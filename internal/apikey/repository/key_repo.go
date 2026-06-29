package repository

import (
	"fmt"
	"time"

	"xorm.io/xorm"

	"llm-gateway/internal/apikey/domain"
)

type apiKeyRepository struct {
	db *xorm.Engine
}

func NewApiKeyRepository(db *xorm.Engine) domain.ApiKeyRepository {
	return &apiKeyRepository{db: db}
}

func (r *apiKeyRepository) Create(key *domain.ApiKey) error {
	_, err := r.db.Insert(key)
	return err
}

func (r *apiKeyRepository) GetByID(id string) (*domain.ApiKey, error) {
	key := &domain.ApiKey{}
	has, err := r.db.ID(id).Get(key)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return key, nil
}

func (r *apiKeyRepository) GetByHash(hash string) (*domain.ApiKey, error) {
	key := &domain.ApiKey{}
	has, err := r.db.Where("key_hash = ?", hash).Get(key)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return key, nil
}

func (r *apiKeyRepository) Update(key *domain.ApiKey) error {
	_, err := r.db.ID(key.ID).Update(key)
	return err
}

func (r *apiKeyRepository) UpdateEditable(key *domain.ApiKey) error {
	result, err := r.db.Exec(
		`UPDATE api_keys
		SET name = ?, quota_limit = ?, rate_limit = ?, allowed_models = ?, allowed_ips = ?, expires_at = ?
		WHERE id = ? AND user_id = ?`,
		key.Name,
		key.QuotaLimit,
		key.RateLimit,
		key.AllowedModels,
		key.AllowedIPs,
		key.ExpiresAt,
		key.ID,
		key.UserID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("api key not found")
	}
	return nil
}

func (r *apiKeyRepository) UpdateStatus(id, userID string, status int) error {
	result, err := r.db.Exec("UPDATE api_keys SET status = ? WHERE id = ? AND user_id = ?", status, id, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("api key not found")
	}
	return nil
}

func (r *apiKeyRepository) Delete(id string) error {
	session := r.db.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	if _, err := session.Exec("UPDATE request_logs SET api_key_id = NULL WHERE api_key_id = ?", id); err != nil {
		_ = session.Rollback()
		return err
	}

	if _, err := session.ID(id).Delete(&domain.ApiKey{}); err != nil {
		_ = session.Rollback()
		return err
	}

	return session.Commit()
}

func (r *apiKeyRepository) ListByUserID(userID string) ([]*domain.ApiKey, error) {
	var keys []*domain.ApiKey
	err := r.db.Where("user_id = ?", userID).Desc("created_at").Find(&keys)
	return keys, err
}

func (r *apiKeyRepository) UpdateUsedQuota(id string, amount int64) error {
	_, err := r.db.Exec("UPDATE api_keys SET used_quota = used_quota + ? WHERE id = ?", amount, id)
	if err != nil {
		return fmt.Errorf("api key not found")
	}
	return nil
}

func (r *apiKeyRepository) UpdateLastUsedAt(id string) error {
	now := time.Now()
	_, err := r.db.ID(id).Cols("last_used_at").Update(&domain.ApiKey{LastUsedAt: &now})
	return err
}
