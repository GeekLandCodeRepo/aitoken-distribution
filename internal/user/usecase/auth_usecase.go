package usecase

import (
	"context"
	"errors"
	"regexp"
	"time"

	"llm-gateway/internal/shared/uuid"

	"llm-gateway/internal/shared/crypto"
	"llm-gateway/internal/shared/errcode"
	"llm-gateway/internal/shared/jwt"
	"llm-gateway/internal/user/domain"
)

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,64}$`)
)

type AuthUsecase struct {
	userRepo     domain.UserRepository
	tokenManager *jwt.TokenManager
}

func NewAuthUsecase(userRepo domain.UserRepository, tokenManager *jwt.TokenManager) *AuthUsecase {
	return &AuthUsecase{
		userRepo:     userRepo,
		tokenManager: tokenManager,
	}
}

type RegisterRequest struct {
	Email      string `json:"email"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	InviteCode string `json:"invite_code"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (uc *AuthUsecase) Register(ctx context.Context, req RegisterRequest) (*domain.User, error) {
	requireInviteRaw, err := uc.userRepo.GetSetting(settingRequireInvite)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	requireInvite := requireInviteRaw == "true"
	if requireInvite && req.InviteCode == "" {
		return nil, errcode.ErrInviteRequired
	}

	if !emailRegex.MatchString(req.Email) {
		return nil, errcode.ErrInvalidEmail
	}

	if !usernameRegex.MatchString(req.Username) {
		return nil, errcode.ErrInvalidUsername
	}

	if len(req.Password) < 8 || len(req.Password) > 64 {
		return nil, errcode.ErrInvalidPassword
	}

	existingUser, err := uc.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	if existingUser != nil {
		return nil, errcode.ErrEmailExists
	}

	existingUser, err = uc.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	if existingUser != nil {
		return nil, errcode.ErrUsernameExists
	}

	hashedPassword, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, errcode.ErrInternal
	}

	user := &domain.User{
		ID:           uuid.NewV7String(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         1,
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
	if req.InviteCode != "" {
		if _, err := uc.userRepo.UseInviteCode(req.InviteCode, user.ID); err != nil {
			_ = uc.userRepo.Delete(user.ID)
			return nil, errcode.ErrInviteInvalid
		}
	}

	return user, nil
}

func (uc *AuthUsecase) Login(ctx context.Context, req LoginRequest) (*TokenResponse, error) {
	if !emailRegex.MatchString(req.Email) {
		return nil, errcode.ErrInvalidEmail
	}

	user, err := uc.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	if user == nil {
		return nil, errcode.ErrBadCredentials
	}

	if user.Status == 0 {
		return nil, errcode.ErrUserDisabled
	}

	if !crypto.VerifyPassword(req.Password, user.PasswordHash) {
		return nil, errcode.ErrBadCredentials
	}

	accessToken, err := uc.tokenManager.GenerateAccessToken(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		return nil, errcode.ErrInternal
	}

	refreshToken, err := uc.tokenManager.GenerateRefreshToken(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		return nil, errcode.ErrInternal
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    7200,
	}, nil
}

func (uc *AuthUsecase) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	claims, err := uc.tokenManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, errcode.ErrRefreshInvalid
	}

	if claims.TokenType != "refresh" {
		return nil, errcode.ErrRefreshInvalid
	}

	user, err := uc.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	if user == nil || user.Status == 0 {
		return nil, errcode.ErrRefreshInvalid
	}

	newAccessToken, err := uc.tokenManager.GenerateAccessToken(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		return nil, errcode.ErrInternal
	}

	return &TokenResponse{
		AccessToken: newAccessToken,
		ExpiresIn:   7200,
	}, nil
}

func (uc *AuthUsecase) ChangePassword(ctx context.Context, userID string, oldPassword, newPassword string) error {
	if len(newPassword) < 8 || len(newPassword) > 64 {
		return errcode.ErrInvalidPassword
	}

	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return errcode.ErrDatabase
	}
	if user == nil {
		return errcode.ErrUserNotFound
	}

	if !crypto.VerifyPassword(oldPassword, user.PasswordHash) {
		return errcode.ErrWrongPassword
	}

	if oldPassword == newPassword {
		return errcode.ErrSamePassword
	}

	hashedPassword, err := crypto.HashPassword(newPassword)
	if err != nil {
		return errcode.ErrInternal
	}

	user.PasswordHash = hashedPassword
	user.UpdatedAt = time.Now()

	return uc.userRepo.Update(user)
}

func (uc *AuthUsecase) GetCurrentUser(ctx context.Context, userID string) (*domain.User, error) {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	if user == nil {
		return nil, errcode.ErrUserNotFound
	}
	return user, nil
}

var ErrUserNotFound = errors.New("user not found")
