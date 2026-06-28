package usecase

import (
	"time"

	"llm-gateway/internal/shared/uuid"

	"llm-gateway/internal/model/domain"
	"llm-gateway/internal/shared/errcode"
)

type ModelUsecase struct {
	modelRepo domain.ModelRepository
}

func NewModelUsecase(modelRepo domain.ModelRepository) *ModelUsecase {
	return &ModelUsecase{modelRepo: modelRepo}
}

type CreateModelRequest struct {
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

func (uc *ModelUsecase) CreateModel(req CreateModelRequest) (*domain.Model, error) {
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

	existing, _ := uc.modelRepo.GetByChannelAndModel(req.ChannelID, req.ModelName)
	if existing != nil {
		return nil, errcode.ErrModelExists
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	model := &domain.Model{
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

	if err := uc.modelRepo.Create(model); err != nil {
		return nil, errcode.ErrDatabase
	}

	return model, nil
}

func (uc *ModelUsecase) ListModels(channelID *string, enabled *bool, search string) ([]*domain.Model, error) {
	return uc.modelRepo.List(channelID, enabled, search)
}

func (uc *ModelUsecase) GetModel(id string) (*domain.Model, error) {
	model, err := uc.modelRepo.GetByID(id)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	if model == nil {
		return nil, errcode.ErrModelNotFound
	}
	return model, nil
}

func (uc *ModelUsecase) UpdateModel(id string, req CreateModelRequest) (*domain.Model, error) {
	model, err := uc.GetModel(id)
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
		model.ChannelID = req.ChannelID
	}
	if req.ModelName != "" {
		model.ModelName = req.ModelName
	}
	if req.PromptPrice >= 0 {
		model.PromptPrice = req.PromptPrice
	}
	if req.PromptUnit > 0 {
		model.PromptUnit = req.PromptUnit
	}
	if req.CompletionPrice >= 0 {
		model.CompletionPrice = req.CompletionPrice
	}
	if req.CompletionUnit > 0 {
		model.CompletionUnit = req.CompletionUnit
	}
	if req.ImagePrice != nil {
		model.ImagePrice = req.ImagePrice
	}
	if req.AudioPrice != nil {
		model.AudioPrice = req.AudioPrice
	}
	model.CachedPromptPrice = req.CachedPromptPrice
	if req.Currency != "" {
		model.Currency = req.Currency
	}
	if req.Enabled != nil {
		model.Enabled = *req.Enabled
	}

	model.UpdatedAt = time.Now()
	if err := uc.modelRepo.Update(model); err != nil {
		return nil, errcode.ErrDatabase
	}
	return uc.GetModel(model.ID)
}

func (uc *ModelUsecase) ToggleModel(id string, enabled *bool) (*domain.Model, error) {
	model, err := uc.GetModel(id)
	if err != nil {
		return nil, err
	}

	if enabled != nil {
		model.Enabled = *enabled
	} else {
		model.Enabled = !model.Enabled
	}
	if err := uc.modelRepo.UpdateEnabled(model.ID, model.Enabled); err != nil {
		return nil, errcode.ErrDatabase
	}
	return uc.GetModel(model.ID)
}

func (uc *ModelUsecase) DeleteModel(id string) error {
	model, err := uc.GetModel(id)
	if err != nil {
		return err
	}
	return uc.modelRepo.Delete(model.ID)
}

func (uc *ModelUsecase) GetModelByChannelAndModel(channelID string, modelName string) (*domain.Model, error) {
	return uc.modelRepo.GetByChannelAndModel(channelID, modelName)
}

func (uc *ModelUsecase) CalculateCost(pricing *domain.Model, promptTokens, completionTokens int, cacheHit bool) int64 {
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
