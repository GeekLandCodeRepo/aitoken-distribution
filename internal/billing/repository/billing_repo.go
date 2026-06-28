package repository

import (
	"time"

	"xorm.io/xorm"

	"llm-gateway/internal/billing/domain"
)

type transactionRepository struct {
	db *xorm.Engine
}

func NewTransactionRepository(db *xorm.Engine) domain.TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(tx *domain.Transaction) error {
	_, err := r.db.Insert(tx)
	return err
}

func (r *transactionRepository) GetByID(id string) (*domain.Transaction, error) {
	tx := &domain.Transaction{}
	has, err := r.db.ID(id).Get(tx)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return tx, nil
}

func (r *transactionRepository) ListByUserID(userID string, page, size int, txType *int) ([]*domain.Transaction, int64, error) {
	var transactions []*domain.Transaction
	query := r.db.Where("user_id = ?", userID).Limit(size, (page-1)*size)

	if txType != nil {
		query = query.Where("type = ?", *txType)
	}

	total, err := query.Desc("created_at").FindAndCount(&transactions)
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

type requestLogRepository struct {
	db *xorm.Engine
}

func NewRequestLogRepository(db *xorm.Engine) domain.RequestLogRepository {
	return &requestLogRepository{db: db}
}

func (r *requestLogRepository) Create(log *domain.RequestLog) error {
	_, err := r.db.Insert(log)
	return err
}

func (r *requestLogRepository) CreateBatch(logs []*domain.RequestLog) error {
	if len(logs) == 0 {
		return nil
	}
	_, err := r.db.Insert(logs)
	return err
}

func (r *requestLogRepository) GetByID(id string) (*domain.RequestLog, error) {
	log := &domain.RequestLog{}
	has, err := r.db.ID(id).Get(log)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return log, nil
}

func (r *requestLogRepository) ListByUserID(userID string, page, size int, model string) ([]*domain.RequestLog, int64, error) {
	var logs []*domain.RequestLog
	query := r.db.Where("user_id = ?", userID).Limit(size, (page-1)*size)

	if model != "" {
		query = query.Where("model = ?", model)
	}

	total, err := query.Desc("created_at").FindAndCount(&logs)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *requestLogRepository) ListAll(page, size int, model string) ([]*domain.RequestLog, int64, error) {
	var logs []*domain.RequestLog
	query := r.db.Limit(size, (page-1)*size)

	if model != "" {
		query = query.Where("model = ?", model)
	}

	total, err := query.Desc("created_at").FindAndCount(&logs)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *requestLogRepository) GetUserOverview(userID string) (*domain.UsageOverview, error) {
	overview := &domain.UsageOverview{}

	// 获取用户余额和已用额度
	user := &struct {
		Balance      int64
		UsedQuota    int64
		RequestCount int64
	}{}
	_, err := r.db.Table("users").Select("balance, used_quota, request_count").Where("id = ?", userID).Get(user)
	if err != nil {
		return nil, err
	}
	overview.Balance = user.Balance
	overview.UsedQuota = user.UsedQuota
	overview.RequestCount = user.RequestCount

	// 今日统计
	today := time.Now().UTC().Truncate(24 * time.Hour)
	_, err = r.db.SQL(`
		SELECT COALESCE(SUM(total_tokens), 0), COALESCE(SUM(cost), 0), COUNT(*)
		FROM request_logs WHERE user_id = ? AND created_at >= ?
	`, userID, today).Get(&overview.Today.Tokens, &overview.Today.Cost, &overview.Today.Requests)
	if err != nil {
		return nil, err
	}

	// 本月统计
	firstOfMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	_, err = r.db.SQL(`
		SELECT COALESCE(SUM(total_tokens), 0), COALESCE(SUM(cost), 0), COUNT(*)
		FROM request_logs WHERE user_id = ? AND created_at >= ?
	`, userID, firstOfMonth).Get(&overview.ThisMonth.Tokens, &overview.ThisMonth.Cost, &overview.ThisMonth.Requests)
	if err != nil {
		return nil, err
	}

	return overview, nil
}

func (r *requestLogRepository) GetGlobalOverview() (*domain.UsageOverview, error) {
	overview := &domain.UsageOverview{}

	// 全局统计
	total, _ := r.db.Count(&domain.RequestLog{})
	overview.RequestCount = total

	r.db.SQL("SELECT COALESCE(SUM(cost), 0) FROM request_logs").Get(&overview.UsedQuota)

	// 今日统计
	today := time.Now().UTC().Truncate(24 * time.Hour)
	r.db.SQL(`
		SELECT COALESCE(SUM(total_tokens), 0), COALESCE(SUM(cost), 0), COUNT(*)
		FROM request_logs WHERE created_at >= ?
	`, today).Get(&overview.Today.Tokens, &overview.Today.Cost, &overview.Today.Requests)

	// 本月统计
	firstOfMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	r.db.SQL(`
		SELECT COALESCE(SUM(total_tokens), 0), COALESCE(SUM(cost), 0), COUNT(*)
		FROM request_logs WHERE created_at >= ?
	`, firstOfMonth).Get(&overview.ThisMonth.Tokens, &overview.ThisMonth.Cost, &overview.ThisMonth.Requests)

	return overview, nil
}

type userRepository struct {
	db *xorm.Engine
}

func NewUserRepository(db *xorm.Engine) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByID(id string) (*domain.User, error) {
	user := &domain.User{}
	has, err := r.db.ID(id).Get(user)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return user, nil
}

func (r *userRepository) UpdateBalance(id string, amount int64) error {
	_, err := r.db.ID(id).Incr("balance", amount).Update(&domain.User{})
	return err
}

func (r *userRepository) ApplyUsage(id string, cost int64) error {
	_, err := r.db.Exec(`
		UPDATE users
		SET balance = balance - ?,
		    used_quota = used_quota + ?,
		    request_count = request_count + 1,
		    updated_at = NOW()
		WHERE id = ?
	`, cost, cost, id)
	return err
}

type apiKeyRepository struct {
	db *xorm.Engine
}

func NewApiKeyRepository(db *xorm.Engine) domain.ApiKeyRepository {
	return &apiKeyRepository{db: db}
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

func (r *apiKeyRepository) UpdateUsedQuota(id string, amount int64) error {
	_, err := r.db.ID(id).Incr("used_quota", amount).Update(&domain.ApiKey{})
	return err
}

type redeemCodeRepository struct {
	db *xorm.Engine
}

func NewRedeemCodeRepository(db *xorm.Engine) domain.RedeemCodeRepository {
	return &redeemCodeRepository{db: db}
}

func (r *redeemCodeRepository) GetByCode(code string) (*domain.RedeemCode, error) {
	redeem := &domain.RedeemCode{}
	has, err := r.db.Where("code = ?", code).Get(redeem)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return redeem, nil
}

func (r *redeemCodeRepository) Create(redeem *domain.RedeemCode) error {
	_, err := r.db.Insert(redeem)
	return err
}

func (r *redeemCodeRepository) Update(redeem *domain.RedeemCode) error {
	_, err := r.db.ID(redeem.ID).Update(redeem)
	return err
}

func (r *redeemCodeRepository) Delete(id string) error {
	_, err := r.db.ID(id).Delete(&domain.RedeemCode{})
	return err
}

func (r *redeemCodeRepository) List(page, size int, status string) ([]*domain.RedeemCode, int64, error) {
	var codes []*domain.RedeemCode
	query := r.db.Limit(size, (page-1)*size)

	switch status {
	case "unused":
		query = query.Where("used_by IS NULL")
	case "used":
		query = query.Where("used_by IS NOT NULL")
	}

	total, err := query.Desc("created_at").FindAndCount(&codes)
	if err != nil {
		return nil, 0, err
	}

	return codes, total, nil
}
