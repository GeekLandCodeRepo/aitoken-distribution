package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"xorm.io/xorm"

	"llm-gateway/internal/pricing/domain"
	"llm-gateway/internal/shared/cache"
	sharedRedis "llm-gateway/internal/shared/redis"
)

type pricingRepository struct {
	db *xorm.Engine
}

func modelPricingCacheKey(channelID string, modelName string) string {
	return fmt.Sprintf("model_pricing:%s:%s", channelID, url.QueryEscape(modelName))
}

func invalidateModelPricing(channelID string, modelName string) {
	if channelID != "" && modelName != "" {
		cache.DeleteKeys(context.Background(), modelPricingCacheKey(channelID, modelName))
		return
	}
	if channelID != "" {
		cache.DeletePattern(context.Background(), fmt.Sprintf("model_pricing:%s:*", channelID))
		return
	}
	cache.DeletePattern(context.Background(), "model_pricing:*")
}

func NewPricingRepository(db *xorm.Engine) domain.PricingRepository {
	return &pricingRepository{db: db}
}

func (r *pricingRepository) Create(pricing *domain.Pricing) error {
	_, err := r.db.Insert(pricing)
	if err == nil {
		invalidateModelPricing(pricing.ChannelID, pricing.ModelName)
	}
	return err
}

func (r *pricingRepository) GetByID(id string) (*domain.Pricing, error) {
	pricing := &domain.Pricing{}
	has, err := r.db.ID(id).Get(pricing)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return pricing, nil
}

func (r *pricingRepository) GetByChannelAndModel(channelID string, modelName string) (*domain.Pricing, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	key := modelPricingCacheKey(channelID, modelName)
	if client := sharedRedis.GetClient(); client != nil {
		if data, err := client.Get(ctx, key).Bytes(); err == nil {
			pricing := &domain.Pricing{}
			if err := json.Unmarshal(data, pricing); err == nil {
				return pricing, nil
			}
		}
	}

	pricing := &domain.Pricing{}
	has, err := r.db.Where("channel_id = ? AND model_name = ? AND enabled = true", channelID, modelName).Get(pricing)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	if client := sharedRedis.GetClient(); client != nil {
		if data, err := json.Marshal(pricing); err == nil {
			_ = client.Set(ctx, key, data, cache.ModelPricingTTL).Err()
		}
	}
	return pricing, nil
}

func (r *pricingRepository) Update(pricing *domain.Pricing) error {
	old, _ := r.GetByID(pricing.ID)
	_, err := r.db.ID(pricing.ID).AllCols().Update(pricing)
	if err == nil {
		if old != nil {
			invalidateModelPricing(old.ChannelID, old.ModelName)
		}
		invalidateModelPricing(pricing.ChannelID, pricing.ModelName)
	}
	return err
}

func (r *pricingRepository) UpdateEnabled(id string, enabled bool) error {
	pricing, _ := r.GetByID(id)
	result, err := r.db.Table("model_pricing").Where("id = ?", id).Update(map[string]interface{}{
		"enabled": enabled,
	})
	if err != nil {
		return err
	}
	if result == 0 {
		return fmt.Errorf("pricing not found")
	}
	if pricing != nil {
		invalidateModelPricing(pricing.ChannelID, pricing.ModelName)
	}
	return nil
}

func (r *pricingRepository) Delete(id string) error {
	pricing, _ := r.GetByID(id)
	_, err := r.db.ID(id).Delete(&domain.Pricing{})
	if err == nil && pricing != nil {
		invalidateModelPricing(pricing.ChannelID, pricing.ModelName)
	}
	return err
}

func (r *pricingRepository) DeleteByChannelAndModel(channelID string, modelName string) error {
	_, err := r.db.Where("channel_id = ? AND model_name = ?", channelID, modelName).Delete(&domain.Pricing{})
	if err == nil {
		invalidateModelPricing(channelID, modelName)
	}
	return err
}

func (r *pricingRepository) DeleteByChannel(channelID string) error {
	_, err := r.db.Where("channel_id = ?", channelID).Delete(&domain.Pricing{})
	if err == nil {
		invalidateModelPricing(channelID, "")
	}
	return err
}

func (r *pricingRepository) List(channelID *string, enabled *bool, search string) ([]*domain.Pricing, error) {
	var pricings []*domain.Pricing
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

	err := session.Asc("model_name").Find(&pricings)
	return pricings, err
}
