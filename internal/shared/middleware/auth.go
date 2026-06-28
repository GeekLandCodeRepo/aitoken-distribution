package middleware

import (
	"context"
	"net/http"
	"strings"

	"llm-gateway/internal/shared/errcode"
	"llm-gateway/internal/shared/jwt"
	"llm-gateway/internal/shared/resp"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UsernameKey contextKey = "username"
	EmailKey    contextKey = "email"
	RoleKey     contextKey = "role"
)

func Auth(tokenManager *jwt.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				resp.Error(w, errcode.ErrUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				resp.Error(w, errcode.ErrUnauthorized)
				return
			}

			claims, err := tokenManager.ValidateToken(parts[1])
			if err != nil {
				if err == jwt.ErrTokenExpired {
					resp.Error(w, errcode.ErrTokenExpired)
				} else {
					resp.Error(w, errcode.ErrTokenInvalid)
				}
				return
			}

			if claims.TokenType != "access" {
				resp.Error(w, errcode.ErrTokenInvalid)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UsernameKey, claims.Username)
			ctx = context.WithValue(ctx, EmailKey, claims.Email)
			ctx = context.WithValue(ctx, RoleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminOnly() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(RoleKey).(int)
			if !ok || role < 10 {
				resp.Error(w, errcode.ErrForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func GetUserID(ctx context.Context) string {
	if val, ok := ctx.Value(UserIDKey).(string); ok {
		return val
	}
	return ""
}

func GetUsername(ctx context.Context) string {
	if val, ok := ctx.Value(UsernameKey).(string); ok {
		return val
	}
	return ""
}

func GetEmail(ctx context.Context) string {
	if val, ok := ctx.Value(EmailKey).(string); ok {
		return val
	}
	return ""
}

func GetRole(ctx context.Context) int {
	if val, ok := ctx.Value(RoleKey).(int); ok {
		return val
	}
	return 0
}
