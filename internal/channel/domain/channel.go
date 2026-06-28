package domain

import (
	"time"
)

type Channel struct {
	ID           string    `json:"id" xorm:"'id' uuid pk notnull"`
	Name         string    `json:"name" xorm:"varchar(128) notnull"`
	Type         int       `json:"type" xorm:"smallint notnull"`
	BaseURL      string    `json:"base_url" xorm:"'base_url' varchar(512) notnull"`
	APIKeyEnc    string    `json:"-" xorm:"'api_key_enc' text notnull"`
	Status       int       `json:"status" xorm:"smallint default(1)"`
	Priority     int       `json:"priority" xorm:"int default(0)"`
	Weight       int       `json:"weight" xorm:"int default(1)"`
	Balance      float64   `json:"balance" xorm:"decimal(12,4)"`
	Models       string    `json:"models" xorm:"text notnull"`
	ModelMapping string    `json:"model_mapping" xorm:"text"`
	Groups       string    `json:"groups" xorm:"text default('["default"]')"`
	UsedQuota    int64     `json:"used_quota" xorm:"bigint default(0)"`
	RequestCount int64     `json:"request_count" xorm:"bigint default(0)"`
	SuccessCount int64     `json:"success_count" xorm:"bigint default(0)"`
	Config       string    `json:"config" xorm:"text"`
	CreatedAt    time.Time `json:"created_at" xorm:"created"`
	UpdatedAt    time.Time `json:"updated_at" xorm:"updated"`
}

func (Channel) TableName() string {
	return "channels"
}

type ChannelRepository interface {
	Create(channel *Channel) error
	GetByID(id string) (*Channel, error)
	Update(channel *Channel) error
	Delete(id string) error
	List(page, size int, status, channelType *int) ([]*Channel, int64, error)
	ListActive() ([]*Channel, error)
	ListActiveForUser(userID string) ([]*Channel, error)
	GetActiveByModel(model string) ([]*Channel, error)
	GetActiveByModelForUser(model string, userID string) ([]*Channel, error)
	UpdateStats(id string, success bool, quota int64) error
	UpdateStatsBatch(id string, requestCount int64, successCount int64, quota int64) error
}
