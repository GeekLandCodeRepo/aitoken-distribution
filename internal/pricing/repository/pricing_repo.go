package repository

import (
	"fmt"

	"xorm.io/xorm"

	"llm-gateway/internal/pricing/domain"
)

type pricingRepository struct {
	db *xorm.Engine
}

func NewPricingRepository(db *xorm.Engine) domain.PricingRepository {
	return &pricingRepository{db: db}
}

func (r *pricingRepository) Create(pricing *domain.Pricing) error {
	_, err := r.db.Insert(pricing)
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
	pricing := &domain.Pricing{}
	has, err := r.db.Where("channel_id = ? AND model_name = ? AND enabled = true", channelID, modelName).Get(pricing)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return pricing, nil
}

func (r *pricingRepository) Update(pricing *domain.Pricing) error {
	_, err := r.db.ID(pricing.ID).AllCols().Update(pricing)
	return err
}

func (r *pricingRepository) UpdateEnabled(id string, enabled bool) error {
	result, err := r.db.Table("model_pricing").Where("id = ?", id).Update(map[string]interface{}{
		"enabled": enabled,
	})
	if err != nil {
		return err
	}
	if result == 0 {
		return fmt.Errorf("pricing not found")
	}
	return nil
}

func (r *pricingRepository) Delete(id string) error {
	_, err := r.db.ID(id).Delete(&domain.Pricing{})
	return err
}

func (r *pricingRepository) DeleteByChannelAndModel(channelID string, modelName string) error {
	_, err := r.db.Where("channel_id = ? AND model_name = ?", channelID, modelName).Delete(&domain.Pricing{})
	return err
}

func (r *pricingRepository) DeleteByChannel(channelID string) error {
	_, err := r.db.Where("channel_id = ?", channelID).Delete(&domain.Pricing{})
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
