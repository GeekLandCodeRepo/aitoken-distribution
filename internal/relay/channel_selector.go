package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

// Channel 渠道信息
type Channel struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Type      int      `json:"type"`
	BaseURL   string   `json:"base_url"`
	APIKeyEnc string   `json:"api_key_enc"`
	Priority  int      `json:"priority"`
	Weight    int      `json:"weight"`
	Models    []string `json:"models"`
	Groups    []string `json:"groups"`
}

// ChannelSelector 渠道选择器
type ChannelSelector struct {
	redis *redis.Client
}

// NewChannelSelector 创建渠道选择器
func NewChannelSelector(redis *redis.Client) *ChannelSelector {
	return &ChannelSelector{redis: redis}
}

// SelectChannel 选择渠道
func (s *ChannelSelector) SelectChannel(ctx context.Context, model string, userGroup string) (*Channel, error) {
	// 从Redis获取可用渠道列表
	cacheKey := fmt.Sprintf("channel_list:%s:%s", model, userGroup)
	data, err := s.redis.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		// 缓存未命中，返回错误（实际应该从数据库加载）
		return nil, fmt.Errorf("no available channel for model %s", model)
	} else if err != nil {
		return nil, err
	}

	var channels []*Channel
	if err := json.Unmarshal([]byte(data), &channels); err != nil {
		return nil, err
	}

	if len(channels) == 0 {
		return nil, fmt.Errorf("no available channel for model %s", model)
	}

	// 按优先级分组
	priorityGroups := make(map[int][]*Channel)
	for _, ch := range channels {
		priorityGroups[ch.Priority] = append(priorityGroups[ch.Priority], ch)
	}

	// 找到最高优先级
	maxPriority := 0
	for p := range priorityGroups {
		if p > maxPriority {
			maxPriority = p
		}
	}

	// 在最高优先级中按权重选择
	selected := priorityGroups[maxPriority]
	return s.selectByWeight(selected), nil
}

// selectByWeight 按权重选择
func (s *ChannelSelector) selectByWeight(channels []*Channel) *Channel {
	if len(channels) == 1 {
		return channels[0]
	}

	totalWeight := 0
	for _, ch := range channels {
		totalWeight += ch.Weight
	}

	r := rand.Intn(totalWeight)
	for _, ch := range channels {
		r -= ch.Weight
		if r < 0 {
			return ch
		}
	}

	return channels[0]
}

// UpdateChannelList 更新渠道列表缓存
func (s *ChannelSelector) UpdateChannelList(ctx context.Context, model string, userGroup string, channels []*Channel) error {
	cacheKey := fmt.Sprintf("channel_list:%s:%s", model, userGroup)
	data, err := json.Marshal(channels)
	if err != nil {
		return err
	}
	return s.redis.Set(ctx, cacheKey, data, 1*time.Minute).Err()
}

// MarkChannelDown 标记渠道不可用
func (s *ChannelSelector) MarkChannelDown(ctx context.Context, channelID string) error {
	cacheKey := fmt.Sprintf("channel_health:%s", channelID)
	return s.redis.Set(ctx, cacheKey, "down", 30*time.Second).Err()
}

// IsChannelHealthy 检查渠道健康状态
func (s *ChannelSelector) IsChannelHealthy(ctx context.Context, channelID string) bool {
	cacheKey := fmt.Sprintf("channel_health:%s", channelID)
	status, err := s.redis.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return true // 没有记录表示健康
	}
	return status != "down"
}
