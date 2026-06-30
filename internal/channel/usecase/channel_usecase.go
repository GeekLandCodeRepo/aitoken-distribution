package usecase

import (
	"encoding/json"
	"time"

	"llm-gateway/internal/shared/uuid"

	"llm-gateway/internal/channel/domain"
	modelDomain "llm-gateway/internal/model/domain"
	"llm-gateway/internal/shared/crypto"
	"llm-gateway/internal/shared/errcode"
)

type ChannelUsecase struct {
	channelRepo domain.ChannelRepository
	modelRepo   modelDomain.ModelRepository
	keyCrypto   *crypto.ChaCha20Poly1305Crypto
}

func NewChannelUsecase(channelRepo domain.ChannelRepository, modelRepo modelDomain.ModelRepository, keyCrypto *crypto.ChaCha20Poly1305Crypto) *ChannelUsecase {
	return &ChannelUsecase{
		channelRepo: channelRepo,
		modelRepo:   modelRepo,
		keyCrypto:   keyCrypto,
	}
}

type CreateChannelRequest struct {
	Name         string            `json:"name"`
	Type         int               `json:"type"`
	BaseURL      string            `json:"base_url"`
	APIKey       string            `json:"api_key"`
	Models       []string          `json:"models"`
	ModelMapping map[string]string `json:"model_mapping"`
	Priority     int               `json:"priority"`
	Weight       int               `json:"weight"`
	Groups       []string          `json:"groups"`
	Config       map[string]any    `json:"config"`
}

type ModelPlazaChannel struct {
	ChannelID   string            `json:"channel_id"`
	ChannelName string            `json:"channel_name"`
	ChannelType int               `json:"channel_type"`
	Models      []ModelPlazaModel `json:"models"`
}

type ModelPlazaModel struct {
	ModelName         string  `json:"model_name"`
	PromptPrice       float64 `json:"prompt_price"`
	PromptUnit        int     `json:"prompt_unit"`
	CompletionPrice   float64 `json:"completion_price"`
	CompletionUnit    int     `json:"completion_unit"`
	CachedPromptPrice float64 `json:"cached_prompt_price"`
	Currency          string  `json:"currency"`
}

func (uc *ChannelUsecase) CreateChannel(req CreateChannelRequest) (*domain.Channel, error) {
	if req.Name == "" {
		return nil, errcode.ErrChanNameRequired
	}
	if req.Type == 0 {
		return nil, errcode.ErrChanTypeInvalid
	}
	if req.BaseURL == "" {
		return nil, errcode.ErrChanURLInvalid
	}
	if req.APIKey == "" {
		return nil, errcode.ErrChanKeyEmpty
	}
	if len(req.Models) == 0 {
		return nil, errcode.ErrChanModelsEmpty
	}

	// 加密 API Key
	encKey, err := uc.keyCrypto.Encrypt(req.APIKey)
	if err != nil {
		return nil, errcode.ErrInternal
	}

	modelsJSON, _ := json.Marshal(req.Models)

	var modelMappingJSON string
	if req.ModelMapping != nil {
		data, _ := json.Marshal(req.ModelMapping)
		modelMappingJSON = string(data)
	}

	groupsJSON, _ := json.Marshal([]string{"default"})
	if req.Groups != nil {
		groupsJSON, _ = json.Marshal(req.Groups)
	}

	var configJSON string
	if req.Config != nil {
		data, _ := json.Marshal(req.Config)
		configJSON = string(data)
	}

	if req.Weight == 0 {
		req.Weight = 1
	}

	channel := &domain.Channel{
		ID:           uuid.NewV7String(),
		Name:         req.Name,
		Type:         req.Type,
		BaseURL:      req.BaseURL,
		APIKeyEnc:    encKey,
		Status:       0,
		Priority:     req.Priority,
		Weight:       req.Weight,
		Models:       string(modelsJSON),
		ModelMapping: modelMappingJSON,
		Groups:       string(groupsJSON),
		Config:       configJSON,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := uc.channelRepo.Create(channel); err != nil {
		return nil, errcode.ErrDatabase
	}
	uc.ensureDefaultModels(channel.ID, req.Models)

	return channel, nil
}

func (uc *ChannelUsecase) ListChannels(page, size int, status, channelType *int) ([]*domain.Channel, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return uc.channelRepo.List(page, size, status, channelType)
}

func (uc *ChannelUsecase) ListModelPlaza(userID string) ([]ModelPlazaChannel, error) {
	channels, err := uc.channelRepo.ListActiveForUser(userID)
	if err != nil {
		return nil, errcode.ErrDatabase
	}

	enabled := true
	result := make([]ModelPlazaChannel, 0, len(channels))
	for _, channel := range channels {
		channelModels, err := uc.modelRepo.List(&channel.ID, &enabled, "")
		if err != nil {
			return nil, errcode.ErrDatabase
		}
		if len(channelModels) == 0 {
			continue
		}

		models := make([]ModelPlazaModel, 0, len(channelModels))
		for _, model := range channelModels {
			models = append(models, ModelPlazaModel{
				ModelName:         model.ModelName,
				PromptPrice:       model.PromptPrice,
				PromptUnit:        model.PromptUnit,
				CompletionPrice:   model.CompletionPrice,
				CompletionUnit:    model.CompletionUnit,
				CachedPromptPrice: model.CachedPromptPrice,
				Currency:          model.Currency,
			})
		}

		result = append(result, ModelPlazaChannel{
			ChannelID:   channel.ID,
			ChannelName: channel.Name,
			ChannelType: channel.Type,
			Models:      models,
		})
	}

	return result, nil
}

func (uc *ChannelUsecase) GetChannel(id string) (*domain.Channel, error) {
	channel, err := uc.channelRepo.GetByID(id)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	if channel == nil {
		return nil, errcode.ErrChanNotFound
	}
	return channel, nil
}

func (uc *ChannelUsecase) UpdateChannel(id string, req CreateChannelRequest) (*domain.Channel, error) {
	channel, err := uc.GetChannel(id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		channel.Name = req.Name
	}
	if req.Type != 0 {
		channel.Type = req.Type
	}
	if req.BaseURL != "" {
		channel.BaseURL = req.BaseURL
	}
	if req.APIKey != "" {
		encKey, err := uc.keyCrypto.Encrypt(req.APIKey)
		if err != nil {
			return nil, errcode.ErrInternal
		}
		channel.APIKeyEnc = encKey
	}
	if req.Models != nil {
		uc.syncModels(channel.ID, channel.Models, req.Models)
		data, _ := json.Marshal(req.Models)
		channel.Models = string(data)
	}
	if req.ModelMapping != nil {
		data, _ := json.Marshal(req.ModelMapping)
		channel.ModelMapping = string(data)
	}
	if req.Priority != 0 {
		channel.Priority = req.Priority
	}
	if req.Weight != 0 {
		channel.Weight = req.Weight
	}
	if req.Groups != nil {
		data, _ := json.Marshal(req.Groups)
		channel.Groups = string(data)
	}
	if req.Config != nil {
		data, _ := json.Marshal(req.Config)
		channel.Config = string(data)
	}

	channel.UpdatedAt = time.Now()

	if err := uc.channelRepo.Update(channel); err != nil {
		return nil, errcode.ErrDatabase
	}
	if err := uc.modelRepo.UpdateEnabledByChannel(channel.ID, channel.Status == 1); err != nil {
		return nil, errcode.ErrDatabase
	}
	return channel, nil
}

func (uc *ChannelUsecase) DeleteChannel(id string) error {
	channel, err := uc.GetChannel(id)
	if err != nil {
		return err
	}
	if err := uc.modelRepo.DeleteByChannel(channel.ID); err != nil {
		return errcode.ErrDatabase
	}
	return uc.channelRepo.Delete(channel.ID)
}

func (uc *ChannelUsecase) ensureDefaultModels(channelID string, models []string) {
	seen := make(map[string]struct{}, len(models))
	for _, model := range models {
		if model == "" {
			continue
		}
		if _, ok := seen[model]; ok {
			continue
		}
		seen[model] = struct{}{}

		existing, err := uc.modelRepo.GetByChannelAndModel(channelID, model)
		if err != nil || existing != nil {
			continue
		}
		_ = uc.modelRepo.Create(&modelDomain.Model{
			ID:                uuid.NewV7String(),
			ChannelID:         channelID,
			ModelName:         model,
			PromptPrice:       0,
			PromptUnit:        1000000,
			CompletionPrice:   0,
			CompletionUnit:    1000000,
			CachedPromptPrice: 0,
			Currency:          "USD",
			Enabled:           false,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		})
	}
}

func (uc *ChannelUsecase) syncModels(channelID string, oldModelsJSON string, newModels []string) {
	var oldModels []string
	_ = json.Unmarshal([]byte(oldModelsJSON), &oldModels)

	oldSet := make(map[string]struct{}, len(oldModels))
	for _, model := range oldModels {
		if model != "" {
			oldSet[model] = struct{}{}
		}
	}

	newSet := make(map[string]struct{}, len(newModels))
	for _, model := range newModels {
		if model != "" {
			newSet[model] = struct{}{}
		}
	}

	for model := range newSet {
		if _, ok := oldSet[model]; !ok {
			uc.ensureDefaultModels(channelID, []string{model})
		}
	}
	for model := range oldSet {
		if _, ok := newSet[model]; !ok {
			_ = uc.modelRepo.DeleteByChannelAndModel(channelID, model)
		}
	}
}

func (uc *ChannelUsecase) SetChannelStatus(id string, enabled bool) (*domain.Channel, error) {
	channel, err := uc.GetChannel(id)
	if err != nil {
		return nil, err
	}

	status := 0
	if enabled {
		status = 1
	}
	if err := uc.channelRepo.UpdateStatus(channel.ID, status); err != nil {
		return nil, errcode.ErrDatabase
	}
	if err := uc.modelRepo.UpdateEnabledByChannel(channel.ID, enabled); err != nil {
		return nil, errcode.ErrDatabase
	}
	channel.Status = status
	channel.UpdatedAt = time.Now()
	return channel, nil
}

func (uc *ChannelUsecase) GetActiveChannelsByModel(model string) ([]*domain.Channel, error) {
	return uc.channelRepo.GetActiveByModel(model)
}

func (uc *ChannelUsecase) GetDecryptedAPIKey(channel *domain.Channel) (string, error) {
	return uc.keyCrypto.Decrypt(channel.APIKeyEnc)
}
