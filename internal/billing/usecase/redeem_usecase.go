package usecase

import (
	"time"

	"llm-gateway/internal/shared/uuid"

	"llm-gateway/internal/billing/domain"
	"llm-gateway/internal/shared/errcode"
)

type RedeemUsecase struct {
	redeemRepo domain.RedeemCodeRepository
	userRepo   domain.UserRepository
	txRepo     domain.TransactionRepository
}

func NewRedeemUsecase(
	redeemRepo domain.RedeemCodeRepository,
	userRepo domain.UserRepository,
	txRepo domain.TransactionRepository,
) *RedeemUsecase {
	return &RedeemUsecase{
		redeemRepo: redeemRepo,
		userRepo:   userRepo,
		txRepo:     txRepo,
	}
}

type RedeemResponse struct {
	Quota         int64 `json:"quota"`
	BalanceBefore int64 `json:"balance_before"`
	BalanceAfter  int64 `json:"balance_after"`
}

func (uc *RedeemUsecase) Redeem(userID string, code string) (*RedeemResponse, error) {
	if code == "" {
		return nil, errcode.ErrInvalidRedeemCode
	}

	redeem, err := uc.redeemRepo.GetByCode(code)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	if redeem == nil {
		return nil, errcode.ErrCodeNotFound
	}
	if redeem.UsedBy != nil {
		return nil, errcode.ErrCodeUsed
	}
	if redeem.ExpiresAt != nil && redeem.ExpiresAt.Before(time.Now()) {
		return nil, errcode.ErrCodeExpired
	}

	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	if user == nil {
		return nil, errcode.ErrUserNotFound
	}

	balanceBefore := user.Balance

	// 更新余额
	if err := uc.userRepo.UpdateBalance(userID, redeem.Quota); err != nil {
		return nil, errcode.ErrDatabase
	}

	// 标记兑换码已使用
	now := time.Now()
	redeem.UsedBy = &userID
	redeem.UsedAt = &now
	uc.redeemRepo.Update(redeem)

	// 记录交易
	tx := &domain.Transaction{
		ID:            uuid.NewV7String(),
		UserID:        userID,
		Type:          1, // 充值
		Amount:        redeem.Quota,
		BalanceAfter:  balanceBefore + redeem.Quota,
		ReferenceType: "redeem",
		ReferenceID:   redeem.Code,
		Description:   "Redeem code: " + redeem.Code,
		CreatedAt:     time.Now(),
	}
	uc.txRepo.Create(tx)

	return &RedeemResponse{
		Quota:         redeem.Quota,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceBefore + redeem.Quota,
	}, nil
}

func (uc *RedeemUsecase) GenerateCodes(createdBy string, quota int64, count int, expiresAt *time.Time) ([]string, error) {
	if createdBy == "" {
		return nil, errcode.ErrUnauthorized
	}
	if quota <= 0 {
		return nil, errcode.ErrInvalidTopup
	}
	if count <= 0 || count > 1000 {
		return nil, errcode.ErrInvalidBatchCount
	}

	codes := make([]string, 0, count)
	for i := 0; i < count; i++ {
		code := generateCode()
		redeem := &domain.RedeemCode{
			ID:        uuid.NewV7String(),
			Code:      code,
			Quota:     quota,
			CreatedBy: createdBy,
			ExpiresAt: expiresAt,
			CreatedAt: time.Now(),
		}
		if err := uc.redeemRepo.Create(redeem); err != nil {
			return nil, errcode.ErrDatabase
		}
		codes = append(codes, code)
	}

	return codes, nil
}

func (uc *RedeemUsecase) ListCodes(page, size int, status string) ([]*domain.RedeemCode, int64, error) {
	return uc.redeemRepo.List(page, size, status)
}

func (uc *RedeemUsecase) DeleteCode(id string) error {
	return uc.redeemRepo.Delete(id)
}

func generateCode() string {
	return "RC-" + uuid.NewV7String()[:8]
}
