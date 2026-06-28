package domain

import (
	"time"
)

type RequestLog struct {
	ID               string    `json:"id" xorm:"'id' uuid pk notnull"`
	UserID           string    `json:"user_id" xorm:"uuid user_id index"`
	APIKeyID         string    `json:"api_key_id" xorm:"uuid api_key_id"`
	ChannelID        string    `json:"channel_id" xorm:"uuid channel_id"`
	Endpoint         string    `json:"endpoint" xorm:"varchar(128)"`
	Model            string    `json:"model" xorm:"varchar(128) notnull"`
	PromptTokens     int       `json:"prompt_tokens" xorm:"int default(0)"`
	CompletionTokens int       `json:"completion_tokens" xorm:"int default(0)"`
	TotalTokens      int       `json:"total_tokens" xorm:"int default(0)"`
	ReasoningTokens  int       `json:"reasoning_tokens" xorm:"int default(0)"`
	Cost             int64     `json:"cost" xorm:"bigint default(0)"`
	CacheHit         bool      `json:"cache_hit" xorm:"bool default(false)"`
	CacheTokens      int       `json:"cache_tokens" xorm:"int default(0)"`
	StatusCode       int       `json:"status_code" xorm:"smallint"`
	IsStream         bool      `json:"is_stream" xorm:"bool default(false)"`
	FirstByteMs      int       `json:"first_byte_ms" xorm:"int default(0)"`
	LatencyMs        int       `json:"latency_ms" xorm:"int default(0)"`
	ErrorMessage     string    `json:"error_message" xorm:"text"`
	RequestID        string    `json:"request_id" xorm:"'request_id' varchar(64)"`
	IPAddress        string    `json:"ip_address" xorm:"'ip_address' varchar(45)"`
	CreatedAt        time.Time `json:"created_at" xorm:"created"`
}

func (RequestLog) TableName() string {
	return "request_logs"
}

type RequestLogRepository interface {
	Create(log *RequestLog) error
	GetByID(id string) (*RequestLog, error)
	ListByUserID(userID string, page, size int, model string, key string) ([]*RequestLogItem, int64, error)
	ListAll(page, size int, model string, key string) ([]*RequestLogItem, int64, error)
	GetUserOverview(userID string) (*UsageOverview, error)
	GetUserStats(userID string, days int) (*UsageStatsResponse, error)
	GetGlobalOverview() (*UsageOverview, error)
	GetDailyStats(date time.Time) (*UsageStats, error)
	GetTopModels(limit int) ([]*ModelUsageStats, error)
	GetTopUsers(limit int) ([]*UserUsageStats, error)
}

type UsageStat struct {
	Date             string `json:"date"`
	Requests         int64  `json:"requests"`
	Tokens           int64  `json:"tokens"`
	PromptTokens     int64  `json:"prompt_tokens"`
	CompletionTokens int64  `json:"completion_tokens"`
	Cost             int64  `json:"cost"`
}

type UsageByModel struct {
	Model    string `json:"model"`
	Requests int64  `json:"requests"`
	Tokens   int64  `json:"tokens"`
	Cost     int64  `json:"cost"`
}

type UsageStatsResponse struct {
	Stats   []*UsageStat    `json:"stats"`
	ByModel []*UsageByModel `json:"by_model"`
}

type RequestLogItem struct {
	RequestLog `xorm:"extends"`
	Username   string `json:"username" xorm:"username"`
	Email      string `json:"email" xorm:"email"`
	KeyName    string `json:"key_name" xorm:"key_name"`
	KeyPrefix  string `json:"key_prefix" xorm:"key_prefix"`
	KeySuffix  string `json:"key_suffix" xorm:"key_suffix"`
	Channel    string `json:"channel" xorm:"channel"`
}

type UsageOverview struct {
	Balance      int64      `json:"balance"`
	UsedQuota    int64      `json:"used_quota"`
	RequestCount int64      `json:"request_count"`
	Today        UsageStats `json:"today"`
	ThisMonth    UsageStats `json:"this_month"`
}

type UsageStats struct {
	Requests int64 `json:"requests"`
	Tokens   int64 `json:"tokens"`
	Cost     int64 `json:"cost"`
}

type ModelUsageStats struct {
	Model      string  `json:"model"`
	Requests   int64   `json:"requests"`
	Tokens     int64   `json:"tokens"`
	Cost       int64   `json:"cost"`
	Percentage float64 `json:"percentage"`
}

type UserUsageStats struct {
	UserID     string  `json:"user_id"`
	Username   string  `json:"username"`
	Email      string  `json:"email"`
	Requests   int64   `json:"requests"`
	Tokens     int64   `json:"tokens"`
	Cost       int64   `json:"cost"`
	Percentage float64 `json:"percentage"`
}
