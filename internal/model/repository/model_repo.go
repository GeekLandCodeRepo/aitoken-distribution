package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"xorm.io/xorm"

	"llm-gateway/internal/model/domain"
	"llm-gateway/internal/shared/cache"
	sharedRedis "llm-gateway/internal/shared/redis"
)

type modelRepository struct {
	db *xorm.Engine
}

func modelCacheKey(channelID string, modelName string) string {
	return fmt.Sprintf("models:%s:%s", channelID, url.QueryEscape(modelName))
}

func invalidateModelCache(channelID string, modelName string) {
	if channelID != "" && modelName != "" {
		cache.DeleteKeys(context.Background(), modelCacheKey(channelID, modelName))
		return
	}
	if channelID != "" {
		cache.DeletePattern(context.Background(), fmt.Sprintf("models:%s:*", channelID))
		return
	}
	cache.DeletePattern(context.Background(), "models:*")
}

func NewModelRepository(db *xorm.Engine) domain.ModelRepository {
	return &modelRepository{db: db}
}

func (r *modelRepository) Create(model *domain.Model) error {
	_, err := r.db.Insert(model)
	if err == nil {
		invalidateModelCache(model.ChannelID, model.ModelName)
	}
	return err
}

func (r *modelRepository) GetByID(id string) (*domain.Model, error) {
	model := &domain.Model{}
	has, err := r.db.ID(id).Get(model)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return model, nil
}

func (r *modelRepository) GetByChannelAndModel(channelID string, modelName string) (*domain.Model, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	key := modelCacheKey(channelID, modelName)
	if client := sharedRedis.GetClient(); client != nil {
		if data, err := client.Get(ctx, key).Bytes(); err == nil {
			model := &domain.Model{}
			if err := json.Unmarshal(data, model); err == nil {
				return model, nil
			}
		}
	}

	model := &domain.Model{}
	has, err := r.db.Where("channel_id = ? AND model_name = ? AND enabled = true", channelID, modelName).Get(model)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	if client := sharedRedis.GetClient(); client != nil {
		if data, err := json.Marshal(model); err == nil {
			_ = client.Set(ctx, key, data, cache.ModelTTL).Err()
		}
	}
	return model, nil
}

func (r *modelRepository) Update(model *domain.Model) error {
	old, _ := r.GetByID(model.ID)
	_, err := r.db.ID(model.ID).AllCols().Update(model)
	if err == nil {
		if old != nil {
			invalidateModelCache(old.ChannelID, old.ModelName)
		}
		invalidateModelCache(model.ChannelID, model.ModelName)
	}
	return err
}

func (r *modelRepository) UpdateEnabled(id string, enabled bool) error {
	model, _ := r.GetByID(id)
	result, err := r.db.Table("models").Where("id = ?", id).Update(map[string]interface{}{
		"enabled": enabled,
	})
	if err != nil {
		return err
	}
	if result == 0 {
		return fmt.Errorf("model not found")
	}
	if model != nil {
		invalidateModelCache(model.ChannelID, model.ModelName)
	}
	return nil
}

func (r *modelRepository) Delete(id string) error {
	model, _ := r.GetByID(id)
	_, err := r.db.ID(id).Delete(&domain.Model{})
	if err == nil && model != nil {
		invalidateModelCache(model.ChannelID, model.ModelName)
	}
	return err
}

func (r *modelRepository) DeleteByChannelAndModel(channelID string, modelName string) error {
	_, err := r.db.Where("channel_id = ? AND model_name = ?", channelID, modelName).Delete(&domain.Model{})
	if err == nil {
		invalidateModelCache(channelID, modelName)
	}
	return err
}

func (r *modelRepository) DeleteByChannel(channelID string) error {
	_, err := r.db.Where("channel_id = ?", channelID).Delete(&domain.Model{})
	if err == nil {
		invalidateModelCache(channelID, "")
	}
	return err
}

func (r *modelRepository) List(channelID *string, enabled *bool, search string) ([]*domain.Model, error) {
	var models []*domain.Model
	session := r.db.NewSession()
	defer session.Close()

	if channelID != nil {
		session = session.Where("channel_id = ?", *channelID)
	}
	if enabled != nil {
		session = session.Where("enabled = ?", *enabled)
	}
	if search != "" {
		session = session.Where("model_name LIKE ?", fmt.Sprintf("%%%s%%", search))
	}

	err := session.Asc("model_name").Find(&models)
	return models, err
}
