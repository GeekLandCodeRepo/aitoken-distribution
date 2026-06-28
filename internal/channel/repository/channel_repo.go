package repository

import (
	"encoding/json"
	"fmt"

	"xorm.io/xorm"

	"llm-gateway/internal/channel/domain"
)

type channelRepository struct {
	db *xorm.Engine
}

func NewChannelRepository(db *xorm.Engine) domain.ChannelRepository {
	return &channelRepository{db: db}
}

func (r *channelRepository) Create(channel *domain.Channel) error {
	_, err := r.db.Insert(channel)
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
	return err
}

func (r *channelRepository) Delete(id string) error {
	_, err := r.db.ID(id).Delete(&domain.Channel{})
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
