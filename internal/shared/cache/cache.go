package cache

import (
	"context"
	"time"

	"llm-gateway/internal/shared/redis"
)

const (
	ChannelCandidatesTTL = 45 * time.Second
	ModelPricingTTL      = 2 * time.Minute
)

func DeleteKeys(ctx context.Context, keys ...string) {
	client := redis.GetClient()
	if client == nil || len(keys) == 0 {
		return
	}
	_ = client.Del(ctx, keys...).Err()
}

func DeletePattern(ctx context.Context, pattern string) {
	client := redis.GetClient()
	if client == nil || pattern == "" {
		return
	}
	var cursor uint64
	for {
		keys, next, err := client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return
		}
		if len(keys) > 0 {
			_ = client.Del(ctx, keys...).Err()
		}
		cursor = next
		if cursor == 0 {
			return
		}
	}
}
