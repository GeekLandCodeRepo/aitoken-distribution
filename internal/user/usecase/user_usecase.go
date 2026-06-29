package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"llm-gateway/internal/shared/crypto"
	"llm-gateway/internal/shared/uuid"

	"llm-gateway/internal/shared/errcode"
	"llm-gateway/internal/user/domain"
)

type UserUsecase struct {
	userRepo domain.UserRepository
	redis    *redis.Client
}

func NewUserUsecase(userRepo domain.UserRepository, redis *redis.Client) *UserUsecase {
	return &UserUsecase{userRepo: userRepo, redis: redis}
}

type ListUsersRequest struct {
	Page   int    `json:"page"`
	Size   int    `json:"size"`
	Search string `json:"search"`
	Status *int   `json:"status"`
	Role   *int   `json:"role"`
}

type UpdateUserRequest struct {
	Role      *int    `json:"role"`
	Status    *int    `json:"status"`
	GroupName *string `json:"group_name"`
}

type TopUpRequest struct {
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
}

type ReplaceUserChannelsRequest struct {
	ChannelIDs []string `json:"channel_ids"`
}

type CreateUserRequest struct {
	// Email is the new user's unique email address.
	Email string `json:"email"`
	// Username is the new user's unique username.
	Username string `json:"username"`
	// Password is the new user's plain password before hashing.
	Password string `json:"password"`
	// Role is the assigned role, where 1 is normal user and 10 is admin.
	Role int `json:"role"`
	// Balance is the initial balance in internal units.
	Balance int64 `json:"balance"`
}

type InviteSettings struct {
	// RequireInviteRegister controls whether public registration must provide an invite code.
	RequireInviteRegister bool `json:"require_invite_register"`
	// UserInviteEnabled controls whether normal users may create invite codes.
	UserInviteEnabled bool `json:"user_invite_enabled"`
	// RewardAmount is the default inviter cashback amount in internal units.
	RewardAmount int64 `json:"reward_amount"`
	// NewUserBonusAmount is the default bonus copied onto newly created invite codes.
	NewUserBonusAmount int64 `json:"new_user_bonus_amount"`
}

type CreateInviteCodeRequest struct {
	// InviterUserID is the optional owner of the invite code.
	InviterUserID string `json:"inviter_user_id"`
	// RewardAmount overrides the default inviter cashback for this invite code.
	RewardAmount *int64 `json:"reward_amount"`
	// NewUserBonus overrides the default new-user bonus for this invite code.
	NewUserBonus *int64 `json:"new_user_bonus"`
}

const (
	// DefaultResetPassword is the password assigned by admin password reset.
	DefaultResetPassword = "Aa1234567"

	// settingUserInviteEnabled stores whether normal users can create invite codes.
	settingUserInviteEnabled = "user_invite_enabled"
	// settingRequireInvite stores whether public registration requires an invite code.
	settingRequireInvite = "require_invite_register"
	// settingInviteReward stores the default inviter cashback amount in internal units.
	settingInviteReward = "invite_reward_amount"
	// settingNewUserBonus stores the default new-user bonus copied to invite codes.
	settingNewUserBonus = "invite_new_user_bonus_amount"
)

func (uc *UserUsecase) ListUsers(ctx context.Context, req ListUsersRequest) ([]*domain.User, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 || req.Size > 100 {
		req.Size = 20
	}

	return uc.userRepo.List(req.Page, req.Size, req.Search, req.Status, req.Role)
}

func (uc *UserUsecase) CreateUser(ctx context.Context, req CreateUserRequest) (*domain.User, error) {
	if !emailRegex.MatchString(req.Email) {
		return nil, errcode.ErrInvalidEmail
	}
	if !usernameRegex.MatchString(req.Username) {
		return nil, errcode.ErrInvalidUsername
	}
	if len(req.Password) < 8 || len(req.Password) > 64 {
		return nil, errcode.ErrInvalidPassword
	}

	existing, err := uc.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	if existing != nil {
		return nil, errcode.ErrEmailExists
	}
	existing, err = uc.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	if existing != nil {
		return nil, errcode.ErrUsernameExists
	}

	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, errcode.ErrInternal
	}
	if req.Role == 0 {
		req.Role = 1
	}

	user := &domain.User{
		ID:           uuid.NewV7String(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hash,
		Role:         req.Role,
		Balance:      req.Balance,
		Status:       1,
		GroupName:    "default",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := uc.userRepo.Create(user); err != nil {
		return nil, errcode.ErrDatabase
	}
	if err := uc.userRepo.ApplyDefaultChannelsToUser(user.ID); err != nil {
		_ = uc.userRepo.Delete(user.ID)
		return nil, errcode.ErrDatabase
	}
	return user, nil
}

func (uc *UserUsecase) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := uc.userRepo.GetByID(id)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	if user == nil {
		return nil, errcode.ErrUserNotFound
	}
	return user, nil
}

func (uc *UserUsecase) UpdateUser(ctx context.Context, id string, req UpdateUserRequest) (*domain.User, error) {
	user, err := uc.userRepo.GetByID(id)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	if user == nil {
		return nil, errcode.ErrUserNotFound
	}

	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	if req.GroupName != nil {
		user.GroupName = *req.GroupName
	}
	user.UpdatedAt = time.Now()

	if err := uc.userRepo.Update(user); err != nil {
		return nil, errcode.ErrDatabase
	}

	return user, nil
}

func (uc *UserUsecase) ResetPassword(ctx context.Context, id string) error {
	user, err := uc.userRepo.GetByID(id)
	if err != nil {
		return errcode.ErrDatabase
	}
	if user == nil {
		return errcode.ErrUserNotFound
	}

	hash, err := crypto.HashPassword(DefaultResetPassword)
	if err != nil {
		return errcode.ErrInternal
	}
	user.PasswordHash = hash
	user.UpdatedAt = time.Now()
	if err := uc.userRepo.Update(user); err != nil {
		return errcode.ErrDatabase
	}
	return nil
}

func (uc *UserUsecase) TopUp(ctx context.Context, id string, req TopUpRequest) (int64, int64, error) {
	if req.Amount == 0 {
		return 0, 0, errcode.ErrInvalidAmount
	}

	user, err := uc.userRepo.GetByID(id)
	if err != nil {
		return 0, 0, errcode.ErrDatabase
	}
	if user == nil {
		return 0, 0, errcode.ErrUserNotFound
	}

	txType := 4
	defaultDescription := "Admin balance increase"
	if req.Amount < 0 {
		txType = 2
		defaultDescription = "Admin balance decrease"
	}
	description := strings.TrimSpace(req.Description)
	if description == "" {
		description = defaultDescription
	}

	balanceBefore, balanceAfter, err := uc.userRepo.AdminAdjustBalance(id, req.Amount, txType, description)
	if err != nil {
		if errors.Is(err, domain.ErrInsufficientBalance) {
			return 0, 0, errcode.ErrInsufficientBalance
		}
		return 0, 0, errcode.ErrDatabase
	}
	if uc.redis != nil {
		_ = uc.redis.Del(ctx, fmt.Sprintf("user_balance:%s", id)).Err()
	}

	return balanceBefore, balanceAfter, nil
}

func (uc *UserUsecase) DeleteUser(ctx context.Context, id string) error {
	user, err := uc.userRepo.GetByID(id)
	if err != nil {
		return errcode.ErrDatabase
	}
	if user == nil {
		return errcode.ErrUserNotFound
	}

	return uc.userRepo.Delete(id)
}

func (uc *UserUsecase) ListUserChannelOptions(ctx context.Context, id string) ([]*domain.UserChannelOption, error) {
	user, err := uc.userRepo.GetByID(id)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	if user == nil {
		return nil, errcode.ErrUserNotFound
	}
	options, err := uc.userRepo.ListUserChannelOptions(id)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	return options, nil
}

func (uc *UserUsecase) ReplaceUserChannels(ctx context.Context, id string, req ReplaceUserChannelsRequest) ([]*domain.UserChannelOption, error) {
	user, err := uc.userRepo.GetByID(id)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	if user == nil {
		return nil, errcode.ErrUserNotFound
	}
	if err := uc.userRepo.ReplaceUserChannels(id, req.ChannelIDs); err != nil {
		return nil, errcode.ErrDatabase
	}
	return uc.userRepo.ListUserChannelOptions(id)
}

func (uc *UserUsecase) ListDefaultUserChannelOptions(ctx context.Context) ([]*domain.UserChannelOption, error) {
	options, err := uc.userRepo.ListDefaultUserChannelOptions()
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	return options, nil
}

func (uc *UserUsecase) SaveDefaultUserChannels(ctx context.Context, req ReplaceUserChannelsRequest) ([]*domain.UserChannelOption, error) {
	if err := uc.userRepo.SaveDefaultUserChannelIDs(req.ChannelIDs); err != nil {
		return nil, errcode.ErrDatabase
	}
	return uc.ListDefaultUserChannelOptions(ctx)
}

func (uc *UserUsecase) ApplyDefaultChannelsToAllUsers(ctx context.Context) (int64, error) {
	affected, err := uc.userRepo.ApplyDefaultChannelsToAllUsers()
	if err != nil {
		return 0, errcode.ErrDatabase
	}
	return affected, nil
}

func (uc *UserUsecase) GetInviteSettings(ctx context.Context) (*InviteSettings, error) {
	enabledRaw, err := uc.userRepo.GetSetting(settingUserInviteEnabled)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	rewardRaw, err := uc.userRepo.GetSetting(settingInviteReward)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	requireRaw, err := uc.userRepo.GetSetting(settingRequireInvite)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	bonusRaw, err := uc.userRepo.GetSetting(settingNewUserBonus)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	reward, _ := strconv.ParseInt(rewardRaw, 10, 64)
	bonus, _ := strconv.ParseInt(bonusRaw, 10, 64)
	return &InviteSettings{
		RequireInviteRegister: requireRaw == "true",
		UserInviteEnabled:     enabledRaw == "true",
		RewardAmount:          reward,
		NewUserBonusAmount:    bonus,
	}, nil
}

func (uc *UserUsecase) UpdateInviteSettings(ctx context.Context, req InviteSettings) (*InviteSettings, error) {
	if err := uc.userRepo.SetSetting(settingRequireInvite, strconv.FormatBool(req.RequireInviteRegister)); err != nil {
		return nil, errcode.ErrDatabase
	}
	if err := uc.userRepo.SetSetting(settingUserInviteEnabled, strconv.FormatBool(req.UserInviteEnabled)); err != nil {
		return nil, errcode.ErrDatabase
	}
	if err := uc.userRepo.SetSetting(settingInviteReward, strconv.FormatInt(req.RewardAmount, 10)); err != nil {
		return nil, errcode.ErrDatabase
	}
	if err := uc.userRepo.SetSetting(settingNewUserBonus, strconv.FormatInt(req.NewUserBonusAmount, 10)); err != nil {
		return nil, errcode.ErrDatabase
	}
	return uc.GetInviteSettings(ctx)
}

func (uc *UserUsecase) CreateInviteCode(ctx context.Context, req CreateInviteCodeRequest) (*domain.InviteCode, error) {
	reward := int64(0)
	bonus := int64(0)
	if req.RewardAmount == nil || req.NewUserBonus == nil {
		settings, err := uc.GetInviteSettings(ctx)
		if err != nil {
			return nil, err
		}
		if req.RewardAmount == nil {
			reward = settings.RewardAmount
		}
		if req.NewUserBonus == nil {
			bonus = settings.NewUserBonusAmount
		}
	}
	if req.RewardAmount != nil {
		reward = *req.RewardAmount
	}
	if req.NewUserBonus != nil {
		bonus = *req.NewUserBonus
	}

	code := &domain.InviteCode{
		ID:            uuid.NewV7String(),
		Code:          generateInviteCode(),
		InviterUserID: req.InviterUserID,
		RewardAmount:  reward,
		NewUserBonus:  bonus,
		CreatedAt:     time.Now(),
	}
	if err := uc.userRepo.CreateInviteCode(code); err != nil {
		return nil, errcode.ErrDatabase
	}
	return code, nil
}

func (uc *UserUsecase) ListInviteCodes(ctx context.Context, page, size int) ([]*domain.InviteCode, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return uc.userRepo.ListInviteCodes(page, size)
}

func generateInviteCode() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return uuid.NewV7String()
	}
	return hex.EncodeToString(buf)
}
