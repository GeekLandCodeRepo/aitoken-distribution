package usecase

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
	"llm-gateway/internal/shared/uuid"

	"llm-gateway/internal/billing/domain"
	modelDomain "llm-gateway/internal/model/domain"
	"llm-gateway/internal/shared/errcode"
)

type BillingUsecase struct {
	userRepo  domain.UserRepository
	keyRepo   domain.ApiKeyRepository
	txRepo    domain.TransactionRepository
	logRepo   domain.RequestLogRepository
	modelRepo modelDomain.ModelRepository
	redis     *redis.Client
}

func NewBillingUsecase(
	userRepo domain.UserRepository,
	keyRepo domain.ApiKeyRepository,
	txRepo domain.TransactionRepository,
	logRepo domain.RequestLogRepository,
	modelRepo modelDomain.ModelRepository,
	redis *redis.Client,
) *BillingUsecase {
	return &BillingUsecase{
		userRepo:  userRepo,
		keyRepo:   keyRepo,
		txRepo:    txRepo,
		logRepo:   logRepo,
		modelRepo: modelRepo,
		redis:     redis,
	}
}

func (uc *BillingUsecase) RequestLogRepository() domain.RequestLogRepository {
	if uc == nil {
		return nil
	}
	return uc.logRepo
}

type PreConsumeParams struct {
	UserID       string
	KeyID        string
	ChannelID    string
	Model        string
	PromptTokens int
	IsStream     bool
	IPAddress    string
}

type PostConsumeParams struct {
	RequestID        string
	UserID           string
	KeyID            string
	ChannelID        string
	Endpoint         string
	Model            string
	PromptTokens     int
	CompletionTokens int
	ReasoningTokens  int
	CacheHit         bool
	CacheTokens      int
	StatusCode       int
	IsStream         bool
	FirstByteMs      int
	LatencyMs        int
	ErrorMessage     string
	IPAddress        string
	PreConsumed      int64
}

// PreConsume 预扣费
func (uc *BillingUsecase) PreConsume(ctx context.Context, params PreConsumeParams) (int64, error) {
	// 查找模型配置
	model, err := uc.modelRepo.GetByChannelAndModel(params.ChannelID, params.Model)
	if err != nil {
		return 0, errcode.ErrInternal
	}
	if model == nil {
		return 0, errcode.ErrInvalidBody // 没有定价配置
	}

	// 估算成本：prompt_price / prompt_unit * prompt_tokens
	estimatedCost := int64(math.Ceil(model.PromptPrice / float64(model.PromptUnit) * float64(params.PromptTokens) * 1000000))

	// 检查用户余额
	balanceKey := fmt.Sprintf("user_balance:%s", params.UserID)
	exists, err := uc.redis.Exists(ctx, balanceKey).Result()
	if err != nil {
		return 0, errcode.ErrRedis
	}
	if exists == 0 {
		user, err := uc.userRepo.GetByID(params.UserID)
		if err != nil {
			return 0, errcode.ErrInternal
		}
		if user == nil {
			return 0, errcode.ErrUserNotFound
		}
		if err := uc.redis.SetNX(ctx, balanceKey, user.Balance, 0).Err(); err != nil {
			return 0, errcode.ErrRedis
		}
	}

	balance, err := uc.redis.DecrBy(ctx, balanceKey, estimatedCost).Result()
	if err != nil {
		return 0, errcode.ErrRedis
	}

	if balance < 0 {
		// 余额不足，回滚
		uc.redis.IncrBy(ctx, balanceKey, estimatedCost)
		user, err := uc.userRepo.GetByID(params.UserID)
		if err != nil {
			return 0, errcode.ErrInternal
		}
		if user == nil {
			return 0, errcode.ErrUserNotFound
		}
		if user.Balance >= estimatedCost {
			if err := uc.redis.Set(ctx, balanceKey, user.Balance, 0).Err(); err != nil {
				return 0, errcode.ErrRedis
			}
			balance, err = uc.redis.DecrBy(ctx, balanceKey, estimatedCost).Result()
			if err != nil {
				return 0, errcode.ErrRedis
			}
			if balance >= 0 {
				return estimatedCost, nil
			}
			uc.redis.IncrBy(ctx, balanceKey, estimatedCost)
		}
		return 0, errcode.ErrBalanceInsufficient
	}

	return estimatedCost, nil
}

// PostConsume 结算，返回本次请求的实际扣费金额
func (uc *BillingUsecase) PostConsume(ctx context.Context, params PostConsumeParams) (int64, error) {
	// 查找模型配置
	model, err := uc.modelRepo.GetByChannelAndModel(params.ChannelID, params.Model)
	if err != nil || model == nil {
		// 没有定价时用预扣费金额
		return uc.finalize(ctx, params, params.PreConsumed)
	}

	// 计算实际费用。缓存命中的输入 token 使用独立缓存输入价格。
	cacheTokens := params.CacheTokens
	if cacheTokens < 0 {
		cacheTokens = 0
	}
	if cacheTokens > params.PromptTokens {
		cacheTokens = params.PromptTokens
	}
	normalPromptTokens := params.PromptTokens - cacheTokens
	promptCost := model.PromptPrice / float64(model.PromptUnit) * float64(normalPromptTokens) * 1000000
	if cacheTokens > 0 {
		promptCost += model.CachedPromptPrice / float64(model.PromptUnit) * float64(cacheTokens) * 1000000
	}
	completionCost := model.CompletionPrice / float64(model.CompletionUnit) * float64(params.CompletionTokens) * 1000000
	actualCost := int64(math.Ceil(promptCost + completionCost))

	return uc.finalize(ctx, params, actualCost)
}

func (uc *BillingUsecase) finalize(ctx context.Context, params PostConsumeParams, actualCost int64) (int64, error) {
	delta := actualCost - params.PreConsumed

	// 调整差额
	if delta != 0 {
		balanceKey := fmt.Sprintf("user_balance:%s", params.UserID)
		uc.redis.IncrBy(ctx, balanceKey, -delta)
	}
	if err := uc.userRepo.ApplyUsage(params.UserID, actualCost); err != nil {
		return 0, err
	}
	if actualCost > 0 && params.KeyID != "" {
		if err := uc.keyRepo.UpdateUsedQuota(params.KeyID, actualCost); err != nil {
			return actualCost, err
		}
	}

	// 记录交易
	tx := &domain.Transaction{
		ID:            uuid.NewV7String(),
		UserID:        params.UserID,
		Type:          2, // 消费
		Amount:        -actualCost,
		ReferenceType: "request",
		ReferenceID:   params.RequestID,
		Description:   fmt.Sprintf("%s: %d prompt + %d completion tokens", params.Model, params.PromptTokens, params.CompletionTokens),
		CreatedAt:     time.Now(),
	}

	if err := uc.txRepo.Create(tx); err != nil {
		return actualCost, err
	}

	return actualCost, nil
}

// Refund 退款
func (uc *BillingUsecase) Refund(ctx context.Context, userID string, amount int64) error {
	if amount <= 0 {
		return nil
	}

	balanceKey := fmt.Sprintf("user_balance:%s", userID)
	uc.redis.IncrBy(ctx, balanceKey, amount)

	// 记录退款交易
	tx := &domain.Transaction{
		ID:          uuid.NewV7String(),
		UserID:      userID,
		Type:        3, // 退款
		Amount:      amount,
		Description: "Refund for failed request",
		CreatedAt:   time.Now(),
	}

	if err := uc.txRepo.Create(tx); err != nil {
		return err
	}

	return nil
}
