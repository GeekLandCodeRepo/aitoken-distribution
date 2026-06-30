package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"llm-gateway/internal/shared/cache"
	sharedRedis "llm-gateway/internal/shared/redis"

	"xorm.io/xorm"

	"llm-gateway/internal/channel/domain"
)

type channelRepository struct {
	db *xorm.Engine
}

type cachedChannel struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Type         int       `json:"type"`
	BaseURL      string    `json:"base_url"`
	APIKeyEnc    string    `json:"api_key_enc"`
	Status       int       `json:"status"`
	Priority     int       `json:"priority"`
	Weight       int       `json:"weight"`
	Balance      float64   `json:"balance"`
	Models       string    `json:"models"`
	ModelMapping string    `json:"model_mapping"`
	Groups       string    `json:"groups"`
	UsedQuota    int64     `json:"used_quota"`
	RequestCount int64     `json:"request_count"`
	SuccessCount int64     `json:"success_count"`
	Config       string    `json:"config"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func channelCandidatesCacheKey(userID string, model string) string {
	return fmt.Sprintf("channel_candidates:%s:%s", userID, url.QueryEscape(model))
}

func toCachedChannels(channels []*domain.Channel) []*cachedChannel {
	items := make([]*cachedChannel, 0, len(channels))
	for _, ch := range channels {
		if ch == nil {
			continue
		}
		items = append(items, &cachedChannel{
			ID:           ch.ID,
			Name:         ch.Name,
			Type:         ch.Type,
			BaseURL:      ch.BaseURL,
			APIKeyEnc:    ch.APIKeyEnc,
			Status:       ch.Status,
			Priority:     ch.Priority,
			Weight:       ch.Weight,
			Balance:      ch.Balance,
			Models:       ch.Models,
			ModelMapping: ch.ModelMapping,
			Groups:       ch.Groups,
			UsedQuota:    ch.UsedQuota,
			RequestCount: ch.RequestCount,
			SuccessCount: ch.SuccessCount,
			Config:       ch.Config,
			CreatedAt:    ch.CreatedAt,
			UpdatedAt:    ch.UpdatedAt,
		})
	}
	return items
}

func fromCachedChannels(items []*cachedChannel) []*domain.Channel {
	channels := make([]*domain.Channel, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		channels = append(channels, &domain.Channel{
			ID:           item.ID,
			Name:         item.Name,
			Type:         item.Type,
			BaseURL:      item.BaseURL,
			APIKeyEnc:    item.APIKeyEnc,
			Status:       item.Status,
			Priority:     item.Priority,
			Weight:       item.Weight,
			Balance:      item.Balance,
			Models:       item.Models,
			ModelMapping: item.ModelMapping,
			Groups:       item.Groups,
			UsedQuota:    item.UsedQuota,
			RequestCount: item.RequestCount,
			SuccessCount: item.SuccessCount,
			Config:       item.Config,
			CreatedAt:    item.CreatedAt,
			UpdatedAt:    item.UpdatedAt,
		})
	}
	return channels
}

func NewChannelRepository(db *xorm.Engine) domain.ChannelRepository {
	return &channelRepository{db: db}
}

func (r *channelRepository) Create(channel *domain.Channel) error {
	_, err := r.db.Insert(channel)
	if err == nil {
		cache.DeletePattern(context.Background(), "channel_candidates:*")
	}
	return err
}

func (r *channelRepository) GetByID(id string) (*domain.Channel, error) {
	channel := &domain.Channel{}
	has, err := r.db.ID(id).Get(channel)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return channel, nil
}

func (r *channelRepository) Update(channel *domain.Channel) error {
	_, err := r.db.ID(channel.ID).Update(channel)
	if err == nil {
		cache.DeletePattern(context.Background(), "channel_candidates:*")
	}
	return err
}

func (r *channelRepository) UpdateStatus(id string, status int) error {
	result, err := r.db.Exec("UPDATE channels SET status = ?, updated_at = ? WHERE id = ?", status, time.Now(), id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("channel not found")
	}
	cache.DeletePattern(context.Background(), "channel_candidates:*")
	return nil
}

func (r *channelRepository) Delete(id string) error {
	_, err := r.db.ID(id).Delete(&domain.Channel{})
	if err == nil {
		cache.DeletePattern(context.Background(), "channel_candidates:*")
	}
	return err
}

func (r *channelRepository) List(page, size int, status, channelType *int) ([]*domain.Channel, int64, error) {
	var channels []*domain.Channel
	session := r.db.NewSession().Limit(size, (page-1)*size)
	defer session.Close()

	if status != nil {
		session = session.Where("status = ?", *status)
	}
	if channelType != nil {
		session = session.Where("type = ?", *channelType)
	}

	total, err := session.Desc("priority").Desc("weight").Desc("created_at").FindAndCount(&channels)
	if err != nil {
		return nil, 0, err
	}

	return channels, total, nil
}

func (r *channelRepository) ListActive() ([]*domain.Channel, error) {
	var channels []*domain.Channel
	err := r.db.Where("status = ?", 1).Desc("priority").Desc("weight").Desc("created_at").Find(&channels)
	return channels, err
}

func (r *channelRepository) ListActiveForUser(userID string) ([]*domain.Channel, error) {
	var channels []*domain.Channel
	err := r.db.Table("channels").Alias("c").
		Join("INNER", []string{"user_channel_permissions", "ucp"}, "ucp.channel_id = c.id AND ucp.user_id = ?", userID).
		Where("c.status = ?", 1).
		Desc("c.priority").Desc("c.weight").Desc("c.created_at").
		Find(&channels)
	return channels, err
}

func (r *channelRepository) GetActiveByModel(model string) ([]*domain.Channel, error) {
	var channels []*domain.Channel

	// 查询状态为启用的渠道
	err := r.db.Where("status = 1").Desc("priority").Desc("weight").Find(&channels)
	if err != nil {
		return nil, err
	}

	// 过滤支持指定模型的渠道
	var result []*domain.Channel
	for _, ch := range channels {
		var models []string
		if err := json.Unmarshal([]byte(ch.Models), &models); err != nil {
			continue
		}
		for _, m := range models {
			if m == model {
				result = append(result, ch)
				break
			}
		}
	}

	return result, nil
}

func (r *channelRepository) GetActiveByModelForUser(model string, userID string) ([]*domain.Channel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	key := channelCandidatesCacheKey(userID, model)
	if client := sharedRedis.GetClient(); client != nil {
		if data, err := client.Get(ctx, key).Bytes(); err == nil {
			items := make([]*cachedChannel, 0)
			if err := json.Unmarshal(data, &items); err == nil {
				return fromCachedChannels(items), nil
			}
		}
	}

	channels, err := r.ListActiveForUser(userID)
	if err != nil {
		return nil, err
	}

	var result []*domain.Channel
	for _, ch := range channels {
		var models []string
		if err := json.Unmarshal([]byte(ch.Models), &models); err != nil {
			continue
		}
		for _, m := range models {
			if m == model {
				result = append(result, ch)
				break
			}
		}
	}
	if client := sharedRedis.GetClient(); client != nil {
		if data, err := json.Marshal(toCachedChannels(result)); err == nil {
			_ = client.Set(ctx, key, data, cache.ChannelCandidatesTTL).Err()
		}
	}
	return result, nil
}

func (r *channelRepository) UpdateStats(id string, success bool, quota int64) error {
	sql := "UPDATE channels SET request_count = request_count + 1, used_quota = used_quota + ?"
	args := []interface{}{quota}

	if success {
		sql += ", success_count = success_count + 1"
	}

	sql += " WHERE id = ?"
	args = append(args, id)

	_, err := r.db.Exec(sql, args)
	if err != nil {
		return fmt.Errorf("channel not found")
	}
	return nil
}

func (r *channelRepository) UpdateStatsBatch(id string, requestCount int64, successCount int64, quota int64) error {
	if requestCount <= 0 {
		return nil
	}
	if successCount < 0 {
		successCount = 0
	}
	if quota < 0 {
		quota = 0
	}
	_, err := r.db.Exec(
		"UPDATE channels SET request_count = request_count + ?, success_count = success_count + ?, used_quota = used_quota + ? WHERE id = ?",
		requestCount, successCount, quota, id,
	)
	if err != nil {
		return fmt.Errorf("channel not found")
	}
	return nil
}
