package domain

import (
	"time"
)

type Transaction struct {
	ID            string    `json:"id" xorm:"'id' uuid pk notnull"`
	UserID        string    `json:"user_id" xorm:"uuid user_id notnull index"`
	Type          int       `json:"type" xorm:"smallint notnull"` // 1=充值, 2=消费, 3=退款, 4=赠送
	Amount        int64     `json:"amount" xorm:"bigint notnull"`
	BalanceAfter  int64     `json:"balance_after" xorm:"bigint notnull"`
	ReferenceType string    `json:"reference_type" xorm:"varchar(32)"`
	ReferenceID   string    `json:"reference_id" xorm:"'reference_id' varchar(128)"`
	Description   string    `json:"description" xorm:"text"`
	CreatedAt     time.Time `json:"created_at" xorm:"created"`
}

func (Transaction) TableName() string {
	return "transactions"
}

type TransactionRepository interface {
	Create(tx *Transaction) error
	GetByID(id string) (*Transaction, error)
	ListByUserID(userID string, page, size int, txType *int) ([]*Transaction, int64, error)
}

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
	CreateBatch(logs []*RequestLog) error
	GetByID(id string) (*RequestLog, error)
	ListByUserID(userID string, page, size int, model string) ([]*RequestLog, int64, error)
	ListAll(page, size int, model string) ([]*RequestLog, int64, error)
	GetUserOverview(userID string) (*UsageOverview, error)
	GetGlobalOverview() (*UsageOverview, error)
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

// User 用户模型（用于计费模块跨模块调用）
type User struct {
	ID           string    `json:"id" xorm:"'id' uuid pk notnull"`
	Username     string    `json:"username" xorm:"varchar(64) unique notnull"`
	Email        string    `json:"email" xorm:"varchar(255) unique notnull"`
	PasswordHash string    `json:"-" xorm:"varchar(255) notnull"`
	Role         int       `json:"role" xorm:"smallint default(1)"`
	Balance      int64     `json:"balance" xorm:"bigint default(0)"`
	UsedQuota    int64     `json:"used_quota" xorm:"bigint default(0)"`
	RequestCount int64     `json:"request_count" xorm:"bigint default(0)"`
	Status       int       `json:"status" xorm:"smallint default(1)"`
	GroupName    string    `json:"group_name" xorm:"varchar(32) default('default')"`
	CreatedAt    time.Time `json:"created_at" xorm:"created"`
	UpdatedAt    time.Time `json:"updated_at" xorm:"updated"`
}

func (User) TableName() string {
	return "users"
}

// UserRepository 用户仓库接口（用于计费模块跨模块调用）
type UserRepository interface {
	GetByID(id string) (*User, error)
	UpdateBalance(id string, amount int64) error
	ApplyUsage(id string, cost int64) error
}

// ApiKey API Key模型（用于计费模块跨模块调用）
type ApiKey struct {
	ID            string     `json:"id" xorm:"'id' uuid pk notnull"`
	UserID        string     `json:"user_id" xorm:"uuid user_id notnull index"`
	KeyHash       string     `json:"-" xorm:"varchar(64) unique notnull"`
	KeyPrefix     string     `json:"key_prefix" xorm:"varchar(12) notnull"`
	Name          string     `json:"name" xorm:"varchar(128)"`
	Status        int        `json:"status" xorm:"smallint default(1)"`
	QuotaLimit    int64      `json:"quota_limit" xorm:"bigint default(-1)"`
	UsedQuota     int64      `json:"used_quota" xorm:"bigint default(0)"`
	RateLimit     int        `json:"rate_limit" xorm:"int default(-1)"`
	AllowedModels string     `json:"allowed_models" xorm:"text"`
	AllowedIPs    string     `json:"allowed_ips" xorm:"'allowed_ips' text"`
	ExpiresAt     *time.Time `json:"expires_at" xorm:"datetime"`
	LastUsedAt    *time.Time `json:"last_used_at" xorm:"datetime"`
	CreatedAt     time.Time  `json:"created_at" xorm:"created"`
}

func (ApiKey) TableName() string {
	return "api_keys"
}

// ApiKeyRepository API Key仓库接口（用于计费模块跨模块调用）
type ApiKeyRepository interface {
	GetByID(id string) (*ApiKey, error)
	UpdateUsedQuota(id string, amount int64) error
}

// RedeemCode 充值码模型
type RedeemCode struct {
	ID        string     `json:"id" xorm:"'id' uuid pk notnull"`
	Code      string     `json:"code" xorm:"varchar(32) unique notnull"`
	Quota     int64      `json:"quota" xorm:"bigint notnull"`
	UsedBy    *string    `json:"used_by" xorm:"uuid used_by"`
	UsedAt    *time.Time `json:"used_at" xorm:"datetime"`
	ExpiresAt *time.Time `json:"expires_at" xorm:"datetime"`
	CreatedBy string     `json:"created_by" xorm:"uuid created_by"`
	CreatedAt time.Time  `json:"created_at" xorm:"created"`
}

func (RedeemCode) TableName() string {
	return "redemption_codes"
}

// RedeemCodeRepository 充值码仓库接口
type RedeemCodeRepository interface {
	GetByCode(code string) (*RedeemCode, error)
	Create(redeem *RedeemCode) error
	Update(redeem *RedeemCode) error
	Delete(id string) error
	List(page, size int, status string) ([]*RedeemCode, int64, error)
}
