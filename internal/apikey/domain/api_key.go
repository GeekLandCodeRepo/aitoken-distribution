package domain

import (
	"time"
)

type ApiKey struct {
	ID            string     `json:"id" xorm:"'id' uuid pk notnull"`
	UserID        string     `json:"user_id" xorm:"uuid user_id notnull index"`
	KeyHash       string     `json:"-" xorm:"varchar(64) unique notnull"`
	KeyPrefix     string     `json:"key_prefix" xorm:"varchar(12) notnull comment('密钥前缀')"`
	KeySuffix     string     `json:"key_suffix" xorm:"varchar(8) comment('密钥后缀')"`
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

type ApiKeyRepository interface {
	Create(key *ApiKey) error
	GetByID(id string) (*ApiKey, error)
	GetByHash(hash string) (*ApiKey, error)
	Update(key *ApiKey) error
	UpdateStatus(id, userID string, status int) error
	Delete(id string) error
	ListByUserID(userID string) ([]*ApiKey, error)
	UpdateUsedQuota(id string, amount int64) error
	UpdateLastUsedAt(id string) error
}
