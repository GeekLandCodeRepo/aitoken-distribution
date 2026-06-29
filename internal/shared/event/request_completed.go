package event

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"xorm.io/xorm"

	billingDomain "llm-gateway/internal/billing/domain"
)

const (
	RequestCompletedStream = "aitsd:events:request_completed"
	RequestCompletedGroup  = "aitsd-request-workers"
	RequestCompletedDead   = "aitsd:events:request_completed:dead"
	RequestCompletedMaxLen = 100000
	MaxErrorMessageLength  = 2000
)

type RequestCompletedEvent struct {
	RequestID        string    `json:"request_id"`
	UserID           string    `json:"user_id"`
	APIKeyID         string    `json:"api_key_id"`
	ChannelID        string    `json:"channel_id"`
	Endpoint         string    `json:"endpoint"`
	Model            string    `json:"model"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	ReasoningTokens  int       `json:"reasoning_tokens"`
	Cost             int64     `json:"cost"`
	CacheHit         bool      `json:"cache_hit"`
	CacheTokens      int       `json:"cache_tokens"`
	StatusCode       int       `json:"status_code"`
	IsStream         bool      `json:"is_stream"`
	FirstByteMs      int       `json:"first_byte_ms"`
	LatencyMs        int       `json:"latency_ms"`
	ErrorMessage     string    `json:"error_message"`
	IPAddress        string    `json:"ip_address"`
	ChannelSuccess   bool      `json:"channel_success"`
	ChannelQuota     int64     `json:"channel_quota"`
	CreatedAt        time.Time `json:"created_at"`
}

type Publisher struct {
	redis *redislib.Client
}

func NewPublisher(redis *redislib.Client) *Publisher {
	return &Publisher{redis: redis}
}

func (p *Publisher) PublishRequestCompleted(ctx context.Context, evt RequestCompletedEvent) error {
	if p == nil || p.redis == nil {
		return errors.New("redis publisher is nil")
	}
	if evt.CreatedAt.IsZero() {
		evt.CreatedAt = time.Now()
	}
	evt.ErrorMessage = truncateString(evt.ErrorMessage, MaxErrorMessageLength)
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	return p.redis.XAdd(ctx, &redislib.XAddArgs{
		Stream: RequestCompletedStream,
		MaxLen: RequestCompletedMaxLen,
		Approx: true,
		Values: map[string]interface{}{"data": string(data)},
	}).Err()
}

type Consumer struct {
	redis    *redislib.Client
	db       *xorm.Engine
	consumer string
}

func NewConsumer(redis *redislib.Client, db *xorm.Engine, consumer string) *Consumer {
	return &Consumer{redis: redis, db: db, consumer: consumer}
}

func (c *Consumer) Run(ctx context.Context) error {
	if c.redis == nil {
		return errors.New("redis client is nil")
	}
	if c.db == nil {
		return errors.New("database engine is nil")
	}
	if c.consumer == "" {
		c.consumer = fmt.Sprintf("worker-%d", time.Now().UnixNano())
	}
	if err := c.redis.XGroupCreateMkStream(ctx, RequestCompletedStream, RequestCompletedGroup, "0").Err(); err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return err
	}
	c.recoverPending(ctx)
	nextRecovery := time.Now().Add(30 * time.Second)

	for ctx.Err() == nil {
		if time.Now().After(nextRecovery) {
			c.recoverPending(ctx)
			nextRecovery = time.Now().Add(30 * time.Second)
		}
		streams, err := c.redis.XReadGroup(ctx, &redislib.XReadGroupArgs{
			Group:    RequestCompletedGroup,
			Consumer: c.consumer,
			Streams:  []string{RequestCompletedStream, ">"},
			Count:    100,
			Block:    2 * time.Second,
		}).Result()
		if err != nil {
			if errors.Is(err, redislib.Nil) || ctx.Err() != nil {
				continue
			}
			slog.Error("request event read failed", "error", err)
			time.Sleep(time.Second)
			continue
		}
		for _, stream := range streams {
			if err := c.handleMessages(ctx, stream.Messages); err != nil {
				slog.Error("request event batch failed", "error", err)
			}
		}
	}
	return ctx.Err()
}

func (c *Consumer) handleMessages(ctx context.Context, messages []redislib.XMessage) error {
	var firstErr error
	for _, msg := range messages {
		raw, _ := msg.Values["data"].(string)
		var evt RequestCompletedEvent
		if err := json.Unmarshal([]byte(raw), &evt); err != nil {
			if deadErr := c.moveToDead(ctx, msg, err); deadErr == nil {
				if ackErr := c.redis.XAck(ctx, RequestCompletedStream, RequestCompletedGroup, msg.ID).Err(); ackErr != nil && firstErr == nil {
					firstErr = ackErr
				}
			} else if firstErr == nil {
				firstErr = deadErr
			}
			continue
		}
		if err := c.processEvent(evt); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if err := c.redis.XAck(ctx, RequestCompletedStream, RequestCompletedGroup, msg.ID).Err(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (c *Consumer) processEvent(evt RequestCompletedEvent) error {
	session := c.db.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return err
	}

	var existing billingDomain.RequestLog
	has, err := session.Where("id = ? OR request_id = ?", evt.RequestID, evt.RequestID).Get(&existing)
	if err != nil {
		_ = session.Rollback()
		return err
	}
	if has {
		return session.Commit()
	}

	if _, err := session.Insert(evt.RequestLog()); err != nil {
		_ = session.Rollback()
		return err
	}
	if evt.ChannelID != "" {
		successCount := int64(0)
		if evt.ChannelSuccess {
			successCount = 1
		}
		quota := evt.ChannelQuota
		if quota < 0 {
			quota = 0
		}
		if _, err := session.Exec(
			"UPDATE channels SET request_count = request_count + 1, success_count = success_count + ?, used_quota = used_quota + ? WHERE id = ?",
			successCount, quota, evt.ChannelID,
		); err != nil {
			_ = session.Rollback()
			return err
		}
	}
	return session.Commit()
}

func (e RequestCompletedEvent) RequestLog() *billingDomain.RequestLog {
	createdAt := e.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	return &billingDomain.RequestLog{
		ID:               e.RequestID,
		UserID:           e.UserID,
		APIKeyID:         e.APIKeyID,
		ChannelID:        e.ChannelID,
		Endpoint:         e.Endpoint,
		Model:            e.Model,
		PromptTokens:     e.PromptTokens,
		CompletionTokens: e.CompletionTokens,
		TotalTokens:      e.TotalTokens,
		ReasoningTokens:  e.ReasoningTokens,
		Cost:             e.Cost,
		CacheHit:         e.CacheHit,
		CacheTokens:      e.CacheTokens,
		StatusCode:       e.StatusCode,
		IsStream:         e.IsStream,
		FirstByteMs:      e.FirstByteMs,
		LatencyMs:        e.LatencyMs,
		ErrorMessage:     truncateString(e.ErrorMessage, MaxErrorMessageLength),
		RequestID:        e.RequestID,
		IPAddress:        e.IPAddress,
		CreatedAt:        createdAt,
	}
}

func (c *Consumer) moveToDead(ctx context.Context, msg redislib.XMessage, cause error) error {
	return c.redis.XAdd(ctx, &redislib.XAddArgs{
		Stream: RequestCompletedDead,
		MaxLen: RequestCompletedMaxLen,
		Approx: true,
		Values: map[string]interface{}{
			"source_id": msg.ID,
			"error":     cause.Error(),
			"data":      fmt.Sprint(msg.Values["data"]),
		},
	}).Err()
}

func (c *Consumer) recoverPending(ctx context.Context) {
	start := "0-0"
	for ctx.Err() == nil {
		messages, next, err := c.redis.XAutoClaim(ctx, &redislib.XAutoClaimArgs{
			Stream:   RequestCompletedStream,
			Group:    RequestCompletedGroup,
			Consumer: c.consumer,
			MinIdle:  60 * time.Second,
			Start:    start,
			Count:    100,
		}).Result()
		if err != nil {
			if !errors.Is(err, redislib.Nil) {
				slog.Error("request event pending recovery failed", "error", err)
			}
			return
		}
		if len(messages) > 0 {
			if err := c.handleMessages(ctx, messages); err != nil {
				slog.Error("request event pending batch failed", "error", err)
			}
		}
		if len(messages) == 0 || next == "0-0" || next == start {
			return
		}
		start = next
	}
}

func truncateString(value string, max int) string {
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max]
}
