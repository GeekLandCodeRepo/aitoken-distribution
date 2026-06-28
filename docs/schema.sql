-- LLM Gateway Database Schema
-- 使用 UUIDv7 作为主键

-- 启用 UUIDv7 扩展（PostgreSQL 16+）
-- CREATE EXTENSION IF NOT EXISTS "pg_uuidv7";

-- 1. 用户表
CREATE TABLE users (
    id              VARCHAR(36) PRIMARY KEY,
    username        VARCHAR(64) UNIQUE NOT NULL,
    email           VARCHAR(255) UNIQUE NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    role            SMALLINT DEFAULT 1,          -- 0=禁用, 1=普通用户, 10=管理员
    balance         BIGINT DEFAULT 0,            -- 余额（内部单位，1美元=1,000,000）
    used_quota      BIGINT DEFAULT 0,            -- 已用额度
    request_count   BIGINT DEFAULT 0,            -- 请求次数
    status          SMALLINT DEFAULT 1,          -- 0=禁用, 1=正常
    group_name      VARCHAR(32) DEFAULT 'default', -- 用户组（用于差异化定价）
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE users IS '用户表';
COMMENT ON COLUMN users.role IS '角色: 0=禁用, 1=普通用户, 10=管理员';
COMMENT ON COLUMN users.balance IS '余额（内部单位，1美元=1,000,000）';
COMMENT ON COLUMN users.group_name IS '用户组（用于差异化定价）';

-- 2. API Key表
CREATE TABLE api_keys (
    id              VARCHAR(36) PRIMARY KEY,
    user_id         VARCHAR(36) REFERENCES users(id) ON DELETE CASCADE,
    key_hash        VARCHAR(64) UNIQUE NOT NULL,  -- SHA-256 哈希
    key_prefix      VARCHAR(12) NOT NULL,         -- sk-xxxx 前缀
    name            VARCHAR(128),
    status          SMALLINT DEFAULT 1,          -- 0=禁用, 1=启用
    quota_limit     BIGINT DEFAULT -1,           -- 额度限制，-1=不限制
    used_quota      BIGINT DEFAULT 0,
    rate_limit      INT DEFAULT -1,              -- 每分钟请求限制，-1=不限制
    allowed_models  JSONB,                       -- 允许的模型数组，null=全部
    allowed_ips     JSONB,                       -- IP白名单
    expires_at      TIMESTAMPTZ,                 -- 过期时间，null=永不过期
    last_used_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE api_keys IS 'API Key表';
COMMENT ON COLUMN api_keys.key_hash IS 'SHA-256 哈希，不存明文';
COMMENT ON COLUMN api_keys.quota_limit IS '额度限制，-1=不限制';
COMMENT ON COLUMN api_keys.rate_limit IS '每分钟请求限制，-1=不限制';

-- 3. 供应商渠道表
CREATE TABLE channels (
    id              VARCHAR(36) PRIMARY KEY,
    name            VARCHAR(128) NOT NULL,
    type            SMALLINT NOT NULL,           -- 1=OpenAI, 2=Claude, 3=Gemini, 4=DeepSeek...
    base_url        VARCHAR(512) NOT NULL,
    api_key_enc     TEXT NOT NULL,               -- 加密存储的 API Key
    status          SMALLINT DEFAULT 1,          -- 0=禁用, 1=启用, 2=自动禁用
    priority        INT DEFAULT 0,               -- 优先级，越大越优先
    weight          INT DEFAULT 1,               -- 同优先级内的权重
    balance         DECIMAL(12,4),               -- 渠道余额（美元）
    models          JSONB NOT NULL,              -- ["gpt-4o", "gpt-3.5-turbo"]
    model_mapping   JSONB,                       -- {"gpt-4": "gpt-4-turbo"}
    groups          JSONB DEFAULT '["default"]', -- 适用的用户组
    used_quota      BIGINT DEFAULT 0,
    request_count   BIGINT DEFAULT 0,
    success_count   BIGINT DEFAULT 0,
    config          JSONB,                       -- 供应商特定配置
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE channels IS '供应商渠道表';
COMMENT ON COLUMN channels.type IS '渠道类型: 1=OpenAI, 2=Claude, 3=Gemini, 4=DeepSeek';
COMMENT ON COLUMN channels.priority IS '优先级，越大越优先';
COMMENT ON COLUMN channels.weight IS '同优先级内的权重';

-- 4. 模型定价表
-- 支持不同的计费单位（每1K、每1M、每10M、每200M等）
CREATE TABLE model_pricing (
    id                  VARCHAR(36) PRIMARY KEY,
    model_name          VARCHAR(128) NOT NULL,
    channel_type        SMALLINT NOT NULL,           -- 对应 channel.type
    
    -- 输入价格
    prompt_price        DECIMAL(16,8) NOT NULL,      -- 原始价格数字
    prompt_unit         INT NOT NULL DEFAULT 1000000, -- 对应token数
    
    -- 输出价格
    completion_price    DECIMAL(16,8) NOT NULL,
    completion_unit     INT NOT NULL DEFAULT 1000000,
    
    -- 可选计费
    image_price         DECIMAL(16,8),               -- 每张图片
    audio_price         DECIMAL(16,8),               -- 每分钟音频
    
    -- 缓存折扣
    cache_ratio         DECIMAL(4,2) DEFAULT 0.5,    -- 缓存命中时的折扣
    
    -- 货币
    currency            VARCHAR(3) DEFAULT 'USD',    -- USD, CNY等
    
    enabled             BOOLEAN DEFAULT TRUE,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(model_name, channel_type)
);

COMMENT ON TABLE model_pricing IS '模型定价表';
COMMENT ON COLUMN model_pricing.prompt_price IS '输入价格（原始数字）';
COMMENT ON COLUMN model_pricing.prompt_unit IS '输入价格对应的token数量';
COMMENT ON COLUMN model_pricing.cache_ratio IS '缓存命中时的折扣比例';

-- 5. 请求日志表
CREATE TABLE request_logs (
    id              VARCHAR(36) PRIMARY KEY,
    user_id         VARCHAR(36) REFERENCES users(id),
    api_key_id      VARCHAR(36) REFERENCES api_keys(id),
    channel_id      VARCHAR(36) REFERENCES channels(id),
    model           VARCHAR(128) NOT NULL,
    prompt_tokens   INT DEFAULT 0,
    completion_tokens INT DEFAULT 0,
    total_tokens    INT DEFAULT 0,
    cost            BIGINT DEFAULT 0,            -- 花费（内部单位）
    cache_hit       BOOLEAN DEFAULT FALSE,
    status_code     SMALLINT,
    is_stream       BOOLEAN DEFAULT FALSE,
    latency_ms      INT DEFAULT 0,
    error_message   TEXT,
    request_id      VARCHAR(64),                 -- 链路追踪ID
    ip_address      VARCHAR(45),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE request_logs IS '请求日志表';
COMMENT ON COLUMN request_logs.cost IS '花费（内部单位，1美元=1,000,000）';

-- 6. 交易记录表
CREATE TABLE transactions (
    id              VARCHAR(36) PRIMARY KEY,
    user_id         VARCHAR(36) REFERENCES users(id),
    type            SMALLINT NOT NULL,           -- 1=充值, 2=消费, 3=退款, 4=赠送
    amount          BIGINT NOT NULL,             -- 金额（内部单位）
    balance_after   BIGINT NOT NULL,             -- 操作后余额
    reference_type  VARCHAR(32),                 -- 'redeem','request','admin'...
    reference_id    VARCHAR(128),                -- 关联ID
    description     TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE transactions IS '交易记录表';
COMMENT ON COLUMN transactions.type IS '类型: 1=充值, 2=消费, 3=退款, 4=赠送';

-- 7. 充值码表
CREATE TABLE redemption_codes (
    id              VARCHAR(36) PRIMARY KEY,
    code            VARCHAR(32) UNIQUE NOT NULL,
    quota           BIGINT NOT NULL,             -- 充值额度
    used_by         VARCHAR(36) REFERENCES users(id),
    used_at         TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ,
    created_by      VARCHAR(36) REFERENCES users(id),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE redemption_codes IS '充值码表';

-- 8. 系统配置表
CREATE TABLE system_configs (
    key             VARCHAR(64) PRIMARY KEY,
    value           TEXT NOT NULL,
    description     TEXT,
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE system_configs IS '系统配置表';

-- 索引
CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_user ON api_keys(user_id);
CREATE INDEX idx_request_logs_user_time ON request_logs(user_id, created_at DESC);
CREATE INDEX idx_request_logs_channel ON request_logs(channel_id, created_at DESC);
CREATE INDEX idx_request_logs_model ON request_logs(model, created_at DESC);
CREATE INDEX idx_transactions_user ON transactions(user_id, created_at DESC);
CREATE INDEX idx_channels_status ON channels(status, priority DESC, weight DESC);
CREATE INDEX idx_pricing_model ON model_pricing(model_name, channel_type);
CREATE INDEX idx_redemption_codes ON redemption_codes(code);
