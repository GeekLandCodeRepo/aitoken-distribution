package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired = errors.New("token expired")
	ErrTokenInvalid = errors.New("token invalid")
)

type Claims struct {
	UserID    string `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      int       `json:"role"`
	TokenType string    `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret        string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewTokenManager(secret string, accessExpiry, refreshExpiry int) *TokenManager {
	return &TokenManager{
		secret:        secret,
		accessExpiry:  time.Duration(accessExpiry) * time.Second,
		refreshExpiry: time.Duration(refreshExpiry) * time.Second,
	}
}

func (tm *TokenManager) GenerateAccessToken(userID string, username, email string, role int) (string, error) {
	claims := &Claims{
		UserID:    userID,
		Username:  username,
		Email:     email,
		Role:      role,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tm.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(tm.secret))
}

func (tm *TokenManager) GenerateRefreshToken(userID string, username, email string, role int) (string, error) {
	claims := &Claims{
		UserID:    userID,
		Username:  username,
		Email:     email,
		Role:      role,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tm.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(tm.secret))
}

func (tm *TokenManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return []byte(tm.secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}
