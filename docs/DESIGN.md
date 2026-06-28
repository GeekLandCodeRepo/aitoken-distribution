# LLM Gateway - 项目设计文档

## 1. 项目概述

### 1.1 项目简介
LLM Gateway 是一个AI团队管理平台，支持多供应商API Key的管理和分发。系统提供统一的OpenAI兼容接口，支持多渠道负载均衡、实时计费和用量统计。

### 1.2 核心功能
- 多供应商支持（OpenAI、Claude、Gemini、DeepSeek等）
- API Key管理和分发
- 统一OpenAI兼容接口
- 实时计费和用量统计
- 充值码系统
- 管理员和用户双端界面

### 1.3 技术栈
| 组件 | 技术选型 |
|------|----------|
| 后端框架 | Go + go-chi |
| 数据库 | PostgreSQL |
| 缓存 | Redis |
| ORM | xorm |
| 前端框架 | React + shadcn/ui |
| 图表库 | Recharts |
| 认证 | JWT |

## 2. 系统架构

### 2.1 DDD架构设计
系统采用简化的DDD架构，分为四层：

```
┌─────────────────────────────────────────────────────────────────┐
│                          Handler 层                              │
│  处理HTTP请求，参数验证，调用Usecase，返回响应                      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Usecase 层                               │
│  业务逻辑编排，事务管理，调用Repository                            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                          Domain 层                               │
│  实体定义，Repository接口，业务规则                                │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Repository 层                              │
│  实现Domain接口，xorm数据库操作                                   │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 模块化结构
```
llm-gateway/
├── cmd/
│   └── server/
│       └── main.go           # 主服务入口
├── internal/
│   ├── shared/               # 共享基础设施
│   │   ├── config/
│   │   │   └── config.go
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   ├── ratelimit.go
│   │   │   ├── requestid.go
│   │   │   └── logger.go
│   │   ├── database/
│   │   │   └── xorm.go
│   │   ├── redis/
│   │   │   └── client.go
│   │   ├── crypto/
│   │   │   ├── aes.go
│   │   │   └── hash.go
│   │   ├── jwt/
│   │   │   └── token.go
│   │   ├── errcode/
│   │   │   └── errcode.go
│   │   └── resp/
│   │       └── response.go
│   ├── user/                 # 用户模块
│   │   ├── handler/
│   │   │   ├── auth_handler.go
│   │   │   └── user_handler.go
│   │   ├── usecase/
│   │   │   ├── auth_usecase.go
│   │   │   └── user_usecase.go
│   │   ├── domain/
│   │   │   ├── user.go
│   │   │   └── repository.go
│   │   └── repository/
│   │       └── user_repo.go
│   ├── apikey/               # API Key模块
│   │   ├── handler/
│   │   │   └── key_handler.go
│   │   ├── usecase/
│   │   │   └── key_usecase.go
│   │   ├── domain/
│   │   │   ├── api_key.go
│   │   │   └── repository.go
│   │   └── repository/
│   │       └── key_repo.go
│   ├── channel/              # 渠道模块
│   │   ├── handler/
│   │   │   └── channel_handler.go
│   │   ├── usecase/
│   │   │   └── channel_usecase.go
│   │   ├── domain/
│   │   │   ├── channel.go
│   │   │   └── repository.go
│   │   └── repository/
│   │       └── channel_repo.go
│   ├── pricing/              # 定价模块
│   │   ├── handler/
│   │   │   └── pricing_handler.go
│   │   ├── usecase/
│   │   │   └── pricing_usecase.go
│   │   ├── domain/
│   │   │   ├── pricing.go
│   │   │   └── repository.go
│   │   └── repository/
│   │       └── pricing_repo.go
│   ├── billing/              # 计费模块
│   │   ├── handler/
│   │   │   └── billing_handler.go
│   │   ├── usecase/
│   │   │   ├── billing_usecase.go
│   │   │   └── redeem_usecase.go
│   │   ├── domain/
│   │   │   ├── transaction.go
│   │   │   ├── redeem_code.go
│   │   │   └── repository.go
│   │   └── repository/
│   │       ├── transaction_repo.go
│   │       └── redeem_repo.go
│   ├── relay/                # 代理模块
│   │   ├── handler/
│   │   │   └── relay_handler.go
│   │   ├── usecase/
│   │   │   └── relay_usecase.go
│   │   ├── adaptor/
│   │   │   ├── adaptor.go
│   │   │   ├── openai.go
│   │   │   ├── claude.go
│   │   │   ├── gemini.go
│   │   │   └── deepseek.go
│   │   └── channel_selector.go
│   └── usage/                # 用量模块
│       ├── handler/
│       │   └── usage_handler.go
│       ├── usecase/
│       │   └── usage_usecase.go
│       ├── domain/
│       │   ├── request_log.go
│       │   └── repository.go
│       └── repository/
│           └── log_repo.go
├── migrations/
│   ├── 001_init.up.sql
│   └── 001_init.down.sql
├── web/                      # React前端
│   ├── src/
│   ├── package.json
│   └── vite.config.ts
├── go.mod
├── Makefile
└── docker-compose.yml
```

### 2.3 请求流程
```
Client Request
    │
    ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Chi Middleware Chain                          │
│  RequestID → CORS → Logger → RateLimit → APIKeyAuth             │
└─────────────────────────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Handler: relay.go                           │
│  1. 解析请求，提取 model                                          │
│  2. 调用 BillingService.PreConsume() 预扣费                      │
│  3. 调用 RelayService.Forward() 转发请求                         │
│  4. 成功 → BillingService.PostConsume() 结算差额                 │
│     失败 → BillingService.Refund() 退还预扣                      │
│  5. 异步写入请求日志                                              │
└─────────────────────────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────────────────────────┐
│                    ChannelSelector (Redis)                        │
│  1. 根据 model 查找可用渠道                                       │
│  2. 按 priority + weight 选择                                    │
│  3. 检查渠道健康状态                                              │
│  4. 返回最优渠道                                                  │
└─────────────────────────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Relay Adaptor                               │
│  1. 将请求转换为目标供应商格式                                     │
│  2. 调用上游 API                                                  │
│  3. 流式/非流式响应处理                                            │
│  4. 从响应中提取 token 用量                                        │
│  5. 将响应转换回 OpenAI 格式返回                                   │
└─────────────────────────────────────────────────────────────────┘
```

### 2.4 计费引擎设计

```go
// 核心计费流程
type BillingService struct {
    redis  *redis.Client
    repo   repository.TransactionRepo
}

// 预扣费：估算成本并扣除
func (s *BillingService) PreConsume(ctx context.Context, userID uuid.UUID, keyID uuid.UUID, model string, promptTokens int) (preConsumed int64, err error) {
    pricing := s.pricingRepo.GetByModel(model)
    estimatedCost = pricing.PromptPrice * float64(promptTokens)
    
    // Redis 原子扣减
    newBalance, err := s.redis.DecrBy(ctx, fmt.Sprintf("user_balance:%s", userID), estimatedCost).Result()
    if newBalance < 0 {
        // 余额不足，回滚
        s.redis.IncrBy(ctx, fmt.Sprintf("user_balance:%s", userID), estimatedCost)
        return 0, ErrInsufficientBalance
    }
    return estimatedCost, nil
}

// 结算：根据实际用量计算最终费用
func (s *BillingService) PostConsume(ctx context.Context, params ConsumeParams) error {
    actualCost = calculateCost(pricing, params.PromptTokens, params.CompletionTokens, params.CacheHit)
    delta = actualCost - params.PreConsumed
    
    if delta != 0 {
        // 调整差额
        s.redis.IncrBy(ctx, fmt.Sprintf("user_balance:%s", params.UserID), delta)
    }
    
    // 异步写入数据库
    go s.persistTransaction(params)
    return nil
}

// 退还：请求失败时退还预扣
func (s *BillingService) Refund(ctx context.Context, userID uuid.UUID, amount int64) {
    s.redis.IncrBy(ctx, fmt.Sprintf("user_balance:%s", userID), amount)
}
```

### 2.5 Redis缓存策略

```
缓存 Key 设计：
├── user_balance:{user_id}        # 用户余额（原子操作）
├── api_key:{key_hash}            # API Key 信息（TTL: 5min）
├── channel_list:{model}:{group}  # 可用渠道列表（TTL: 1min）
├── channel_health:{channel_id}   # 渠道健康状态（TTL: 30s）
├── rate_limit:{key_id}:{window}  # 限流计数器（TTL: 1min）
└── pricing:{model}               # 模型定价（TTL: 10min）
```

## 3. 数据库设计

### 3.1 核心表
- users - 用户表
- api_keys - API Key表
- channels - 供应商渠道表
- model_pricing - 模型定价表
- request_logs - 请求日志表
- transactions - 交易记录表
- redemption_codes - 充值码表
- system_configs - 系统配置表

### 3.2 主键设计
所有表使用UUIDv7作为主键，具有时间有序性和全局唯一性。

### 3.3 模型定价设计

各家供应商定价单位不同，系统采用统一存储方案：

```
字段设计：
├── prompt_price      -- 原始价格数字
├── prompt_unit       -- 对应的token数量
├── completion_price  -- 原始价格数字  
├── completion_unit   -- 对应的token数量
```

计费公式：
```
实际费用 = (prompt_tokens / prompt_unit) × prompt_price 
        + (completion_tokens / completion_unit) × completion_price
        × group_ratio  (用户组折扣/加价倍数)
```

详见：[schema.sql](./schema.sql)

## 4. API接口设计

### 4.1 模块划分
- 认证模块 - 登录、注册、Token刷新
- 用户模块 - 用户管理、手动充值
- API Key模块 - Key的CRUD和配置
- 渠道模块 - 供应商渠道管理
- 定价模块 - 模型定价配置
- 计费模块 - 充值码和交易记录
- 代理模块 - OpenAI兼容接口转发
- 用量模块 - 统计和日志查询

详见：[api/](./api/) 目录下各模块文档

## 5. 错误码设计

### 5.1 格式
```
AABBCC
AA = 模块代码 (10-90)
BB = 错误类型 (01-05)
CC = 具体错误序号 (01-99)
```

### 5.2 模块代码
| 代码 | 模块 |
|------|------|
| 10 | 系统 |
| 20 | 认证 |
| 30 | 用户 |
| 40 | API Key |
| 50 | 渠道 |
| 60 | 定价 |
| 70 | 计费 |
| 80 | 代理 |
| 90 | 用量 |

详见：[errors.md](./errors.md)

## 6. 开发计划

### Phase 1: 基础框架（1-2周）
- 项目初始化
- 数据库层
- 认证系统
- 前端基础

### Phase 2: API Key管理（1周）
- API Key CRUD
- Key验证中间件
- 前端页面

### Phase 3: 代理引擎核心（2周）
- 渠道管理
- 渠道选择器
- 供应商适配器
- SSE流式转发

### Phase 4: 计费系统（1-2周）
- 计费引擎
- 充值码系统
- 用量统计

### Phase 5: 完善和优化（1-2周）
- 功能完善
- 性能优化
- 安全加固
- 部署配置

## 7. 配置项

### 7.1 环境变量
```bash
# 数据库
DATABASE_URL=postgres://user:pass@localhost:5432/llm_gateway
DB_MAX_OPEN_CONNS=100
DB_MAX_IDLE_CONNS=10

# Redis
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY=7200

# 服务
PORT=8080
LOG_LEVEL=info

# 计费
QUOTA_PER_USD=1000000
PRE_CONSUME_ENABLED=true
```
