package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	apikeyDomain "llm-gateway/internal/apikey/domain"
	"llm-gateway/internal/shared/errcode"
	"llm-gateway/internal/shared/resp"
)

// APIKeyAuth API Key认证中间件
func APIKeyAuth(apiKeyRepo apikeyDomain.ApiKeyRepository, rdb *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 从Authorization头获取API Key
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				resp.OpenAIError(w, errcode.ErrUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				resp.OpenAIError(w, errcode.ErrUnauthorized)
				return
			}

			apiKey := parts[1]
			if !strings.HasPrefix(apiKey, "sk-") {
				resp.OpenAIError(w, errcode.ErrUnauthorized)
				return
			}

			// 计算Key Hash
			hash := sha256.Sum256([]byte(apiKey))
			keyHash := hex.EncodeToString(hash[:])

			// 从数据库获取Key信息
			keyInfo, err := apiKeyRepo.GetByHash(keyHash)
			if err != nil {
				resp.OpenAIError(w, errcode.ErrInternal)
				return
			}
			if keyInfo == nil {
				resp.OpenAIError(w, errcode.ErrUnauthorized)
				return
			}

			// 检查Key状态
			if keyInfo.Status != 1 {
				resp.OpenAIError(w, errcode.ErrKeyDisabled)
				return
			}

			// 检查是否过期
			if keyInfo.ExpiresAt != nil && keyInfo.ExpiresAt.Before(time.Now()) {
				resp.OpenAIError(w, errcode.ErrKeyExpired)
				return
			}

			// 检查额度
			if keyInfo.QuotaLimit != -1 && keyInfo.UsedQuota >= keyInfo.QuotaLimit {
				resp.OpenAIError(w, errcode.ErrKeyQuotaUsed)
				return
			}

			// 检查限流
			if keyInfo.RateLimit != -1 {
				ctx := context.Background()
				rateLimitKey := "rate_limit:" + keyInfo.ID + ":" + time.Now().Format("200601021504")
				count, err := rdb.Incr(ctx, rateLimitKey).Result()
				if err == nil && count == 1 {
					rdb.Expire(ctx, rateLimitKey, time.Minute)
				}
				if count > int64(keyInfo.RateLimit) {
					resp.OpenAIError(w, errcode.ErrKeyRateLimited)
					return
				}
			}

			// 设置上下文
			ctx := context.WithValue(r.Context(), UserIDKey, keyInfo.UserID)
			ctx = context.WithValue(ctx, "key_id", keyInfo.ID)
			ctx = context.WithValue(ctx, "key_models", keyInfo.AllowedModels)

			// best-effort update, avoid blocking relay requests
			go func(keyID string) {
				_ = apiKeyRepo.UpdateLastUsedAt(keyID)
			}(keyInfo.ID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
