package usecase

import (
	"time"

	"llm-gateway/internal/shared/uuid"

	"llm-gateway/internal/pricing/domain"
	"llm-gateway/internal/shared/errcode"
)

type PricingUsecase struct {
	pricingRepo domain.PricingRepository
}

func NewPricingUsecase(pricingRepo domain.PricingRepository) *PricingUsecase {
	return &PricingUsecase{pricingRepo: pricingRepo}
}

type CreatePricingRequest struct {
	ChannelID         string   `json:"channel_id"`
	ModelName         string   `json:"model_name"`
	PromptPrice       float64  `json:"prompt_price"`
	PromptUnit        int      `json:"prompt_unit"`
	CompletionPrice   float64  `json:"completion_price"`
	CompletionUnit    int      `json:"completion_unit"`
	ImagePrice        *float64 `json:"image_price"`
	AudioPrice        *float64 `json:"audio_price"`
	CachedPromptPrice float64  `json:"cached_prompt_price"`
	Currency          string   `json:"currency"`
	Enabled           *bool    `json:"enabled"`
}

func (uc *PricingUsecase) CreatePricing(req CreatePricingRequest) (*domain.Pricing, error) {
	if req.ChannelID == "" {
		return nil, errcode.ErrInvalidChanType
	}
	if req.ModelName == "" {
		return nil, errcode.ErrModelRequired
	}
	if req.PromptPrice < 0 {
		return nil, errcode.ErrInvalidPromptPrice
	}
	if req.CompletionPrice < 0 {
		return nil, errcode.ErrInvalidCompPrice
	}
	if req.CachedPromptPrice < 0 {
		return nil, errcode.ErrInvalidCachedPromptPrice
	}

	if req.PromptUnit == 0 {
		req.PromptUnit = 1000000
	}
	if req.CompletionUnit == 0 {
		req.CompletionUnit = 1000000
	}
	if req.Currency == "" {
		req.Currency = "USD"
	}

	existing, _ := uc.pricingRepo.GetByChannelAndModel(req.ChannelID, req.ModelName)
	if existing != nil {
		return nil, errcode.ErrPricingExists
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	pricing := &domain.Pricing{
		ID:                uuid.NewV7String(),
		ChannelID:         req.ChannelID,
		ModelName:         req.ModelName,
		PromptPrice:       req.PromptPrice,
		PromptUnit:        req.PromptUnit,
		CompletionPrice:   req.CompletionPrice,
		CompletionUnit:    req.CompletionUnit,
		ImagePrice:        req.ImagePrice,
		AudioPrice:        req.AudioPrice,
		CachedPromptPrice: req.CachedPromptPrice,
		Currency:          req.Currency,
		Enabled:           enabled,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := uc.pricingRepo.Create(pricing); err != nil {
		return nil, errcode.ErrDatabase
	}

	return pricing, nil
}

func (uc *PricingUsecase) ListPricings(channelID *string, enabled *bool, search string) ([]*domain.Pricing, error) {
	return uc.pricingRepo.List(channelID, enabled, search)
}

func (uc *PricingUsecase) GetPricing(id string) (*domain.Pricing, error) {
	pricing, err := uc.pricingRepo.GetByID(id)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	if pricing == nil {
		return nil, errcode.ErrPricingNotFound
	}
	return pricing, nil
}

func (uc *PricingUsecase) UpdatePricing(id string, req CreatePricingRequest) (*domain.Pricing, error) {
	pricing, err := uc.GetPricing(id)
	if err != nil {
		return nil, err
	}
	if req.PromptPrice < 0 {
		return nil, errcode.ErrInvalidPromptPrice
	}
	if req.CompletionPrice < 0 {
		return nil, errcode.ErrInvalidCompPrice
	}
	if req.CachedPromptPrice < 0 {
		return nil, errcode.ErrInvalidCachedPromptPrice
	}

	if req.ChannelID != "" {
		pricing.ChannelID = req.ChannelID
	}
	if req.ModelName != "" {
		pricing.ModelName = req.ModelName
	}
	if req.PromptPrice >= 0 {
		pricing.PromptPrice = req.PromptPrice
	}
	if req.PromptUnit > 0 {
		pricing.PromptUnit = req.PromptUnit
	}
	if req.CompletionPrice >= 0 {
		pricing.CompletionPrice = req.CompletionPrice
	}
	if req.CompletionUnit > 0 {
		pricing.CompletionUnit = req.CompletionUnit
	}
	if req.ImagePrice != nil {
		pricing.ImagePrice = req.ImagePrice
	}
	if req.AudioPrice != nil {
		pricing.AudioPrice = req.AudioPrice
	}
	pricing.CachedPromptPrice = req.CachedPromptPrice
	if req.Currency != "" {
		pricing.Currency = req.Currency
	}
	if req.Enabled != nil {
		pricing.Enabled = *req.Enabled
	}

	pricing.UpdatedAt = time.Now()
	if err := uc.pricingRepo.Update(pricing); err != nil {
		return nil, errcode.ErrDatabase
	}
	return uc.GetPricing(pricing.ID)
}

func (uc *PricingUsecase) TogglePricing(id string, enabled *bool) (*domain.Pricing, error) {
	pricing, err := uc.GetPricing(id)
	if err != nil {
		return nil, err
	}

	if enabled != nil {
		pricing.Enabled = *enabled
	} else {
		pricing.Enabled = !pricing.Enabled
	}
	if err := uc.pricingRepo.UpdateEnabled(pricing.ID, pricing.Enabled); err != nil {
		return nil, errcode.ErrDatabase
	}
	return uc.GetPricing(pricing.ID)
}

func (uc *PricingUsecase) DeletePricing(id string) error {
	pricing, err := uc.GetPricing(id)
	if err != nil {
		return err
	}
	return uc.pricingRepo.Delete(pricing.ID)
}

func (uc *PricingUsecase) GetPricingByChannelAndModel(channelID string, modelName string) (*domain.Pricing, error) {
	return uc.pricingRepo.GetByChannelAndModel(channelID, modelName)
}

func (uc *PricingUsecase) CalculateCost(pricing *domain.Pricing, promptTokens, completionTokens int, cacheHit bool) int64 {
	var promptCost, completionCost float64

	if cacheHit {
		promptCost = float64(promptTokens) / float64(pricing.PromptUnit) * pricing.CachedPromptPrice
	} else {
		promptCost = float64(promptTokens) / float64(pricing.PromptUnit) * pricing.PromptPrice
	}
	completionCost = float64(completionTokens) / float64(pricing.CompletionUnit) * pricing.CompletionPrice

	totalUSD := promptCost + completionCost
	return int64(totalUSD * 1000000)
}
