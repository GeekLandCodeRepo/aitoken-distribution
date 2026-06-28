package domain

import (
	"time"
)

type Model struct {
	ID                string    `json:"id" xorm:"'id' uuid pk notnull"`
	ChannelID         string    `json:"channel_id" xorm:"uuid channel_id notnull unique(channel_model)"`
	ModelName         string    `json:"model_name" xorm:"varchar(128) notnull unique(channel_model)"`
	PromptPrice       float64   `json:"prompt_price" xorm:"decimal(16,8) notnull"`
	PromptUnit        int       `json:"prompt_unit" xorm:"int notnull default(1000000)"`
	CompletionPrice   float64   `json:"completion_price" xorm:"decimal(16,8) notnull"`
	CompletionUnit    int       `json:"completion_unit" xorm:"int notnull default(1000000)"`
	ImagePrice        *float64  `json:"image_price" xorm:"decimal(16,8)"`
	AudioPrice        *float64  `json:"audio_price" xorm:"decimal(16,8)"`
	CachedPromptPrice float64   `json:"cached_prompt_price" xorm:"decimal(16,8) default(0) comment('缓存命中输入价格')"`
	Currency          string    `json:"currency" xorm:"varchar(3) default('USD')"`
	Enabled           bool      `json:"enabled" xorm:"bool default(true)"`
	CreatedAt         time.Time `json:"created_at" xorm:"created"`
	UpdatedAt         time.Time `json:"updated_at" xorm:"updated"`
}

func (Model) TableName() string {
	return "models"
}

type ModelRepository interface {
	Create(model *Model) error
	GetByID(id string) (*Model, error)
	GetByChannelAndModel(channelID string, modelName string) (*Model, error)
	Update(model *Model) error
	UpdateEnabled(id string, enabled bool) error
	UpdateEnabledByChannel(channelID string, enabled bool) error
	Delete(id string) error
	DeleteByChannelAndModel(channelID string, modelName string) error
	DeleteByChannel(channelID string) error
	List(channelID *string, enabled *bool, search string) ([]*Model, error)
}
