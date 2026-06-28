package usecase

import (
	"encoding/json"
	"time"

	"llm-gateway/internal/shared/uuid"

	"llm-gateway/internal/apikey/domain"
	"llm-gateway/internal/shared/crypto"
	"llm-gateway/internal/shared/errcode"
)

type KeyUsecase struct {
	keyRepo domain.ApiKeyRepository
}

func NewKeyUsecase(keyRepo domain.ApiKeyRepository) *KeyUsecase {
	return &KeyUsecase{keyRepo: keyRepo}
}

type CreateKeyRequest struct {
	Name          string   `json:"name"`
	QuotaLimit    int64    `json:"quota_limit"`
	RateLimit     int      `json:"rate_limit"`
	AllowedModels []string `json:"allowed_models"`
	AllowedIPs    []string `json:"allowed_ips"`
	ExpiresAt     *string  `json:"expires_at"`
}

type CreateKeyResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Key        string    `json:"key"`
	KeyPrefix  string    `json:"key_prefix"`
	KeySuffix  string    `json:"key_suffix"`
	QuotaLimit int64     `json:"quota_limit"`
	RateLimit  int       `json:"rate_limit"`
	Status     int       `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

func (uc *KeyUsecase) CreateKey(userID string, req CreateKeyRequest) (*CreateKeyResponse, error) {
	if req.Name == "" {
		return nil, errcode.ErrKeyNameRequired
	}
	if len(req.Name) > 128 {
		return nil, errcode.ErrKeyNameTooLong
	}

	plainKey, hash, prefix, suffix, err := crypto.GenerateAPIKey()
	if err != nil {
		return nil, errcode.ErrInternal
	}

	var allowedModels, allowedIPs string
	if req.AllowedModels != nil {
		data, _ := json.Marshal(req.AllowedModels)
		allowedModels = string(data)
	}
	if req.AllowedIPs != nil {
		data, _ := json.Marshal(req.AllowedIPs)
		allowedIPs = string(data)
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err == nil {
			expiresAt = &t
		}
	}

	if req.QuotaLimit == 0 {
		req.QuotaLimit = -1
	}
	if req.RateLimit == 0 {
		req.RateLimit = -1
	}

	key := &domain.ApiKey{
		ID:            uuid.NewV7String(),
		UserID:        userID,
		KeyHash:       hash,
		KeyPrefix:     prefix,
		KeySuffix:     suffix,
		Name:          req.Name,
		Status:        1,
		QuotaLimit:    req.QuotaLimit,
		RateLimit:     req.RateLimit,
		AllowedModels: allowedModels,
		AllowedIPs:    allowedIPs,
		ExpiresAt:     expiresAt,
		CreatedAt:     time.Now(),
	}

	if err := uc.keyRepo.Create(key); err != nil {
		return nil, errcode.ErrDatabase
	}

	return &CreateKeyResponse{
		ID:         key.ID,
		Name:       key.Name,
		Key:        plainKey,
		KeyPrefix:  key.KeyPrefix,
		KeySuffix:  key.KeySuffix,
		QuotaLimit: key.QuotaLimit,
		RateLimit:  key.RateLimit,
		Status:     key.Status,
		CreatedAt:  key.CreatedAt,
	}, nil
}

func (uc *KeyUsecase) ListKeys(userID string) ([]*domain.ApiKey, error) {
	return uc.keyRepo.ListByUserID(userID)
}

func (uc *KeyUsecase) GetKey(id, userID string) (*domain.ApiKey, error) {
	key, err := uc.keyRepo.GetByID(id)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	if key == nil {
		return nil, errcode.ErrKeyNotFound
	}
	if key.UserID != userID {
		return nil, errcode.ErrKeyNotOwned
	}
	return key, nil
}

func (uc *KeyUsecase) UpdateKey(id, userID string, req CreateKeyRequest) (*domain.ApiKey, error) {
	key, err := uc.GetKey(id, userID)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		key.Name = req.Name
	}
	if req.QuotaLimit != 0 {
		key.QuotaLimit = req.QuotaLimit
	}
	if req.RateLimit != 0 {
		key.RateLimit = req.RateLimit
	}
	if req.AllowedModels != nil {
		data, _ := json.Marshal(req.AllowedModels)
		key.AllowedModels = string(data)
	}
	if req.AllowedIPs != nil {
		data, _ := json.Marshal(req.AllowedIPs)
		key.AllowedIPs = string(data)
	}

	if err := uc.keyRepo.Update(key); err != nil {
		return nil, errcode.ErrDatabase
	}
	return key, nil
}

func (uc *KeyUsecase) DeleteKey(id, userID string) error {
	key, err := uc.GetKey(id, userID)
	if err != nil {
		return err
	}
	return uc.keyRepo.Delete(key.ID)
}

func (uc *KeyUsecase) ToggleKey(id, userID string) (*domain.ApiKey, error) {
	key, err := uc.GetKey(id, userID)
	if err != nil {
		return nil, err
	}

	if key.Status == 1 {
		key.Status = 0
	} else {
		key.Status = 1
	}

	if err := uc.keyRepo.Update(key); err != nil {
		return nil, errcode.ErrDatabase
	}
	return key, nil
}

func (uc *KeyUsecase) GetKeyByHash(hash string) (*domain.ApiKey, error) {
	return uc.keyRepo.GetByHash(hash)
}
