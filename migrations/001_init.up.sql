-- LLM Gateway Database Migration
-- Version: 001
-- Description: Initial schema

-- 1. 用户表
CREATE TABLE IF NOT EXISTS users (
    id              VARCHAR(36) PRIMARY KEY,
    username        VARCHAR(64) UNIQUE NOT NULL,
    email           VARCHAR(255) UNIQUE NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    role            SMALLINT DEFAULT 1,
    balance         BIGINT DEFAULT 0,
    used_quota      BIGINT DEFAULT 0,
    request_count   BIGINT DEFAULT 0,
    status          SMALLINT DEFAULT 1,
    group_name      VARCHAR(32) DEFAULT 'default',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE users IS '用户表';
COMMENT ON COLUMN users.role IS '角色: 0=禁用, 1=普通用户, 10=管理员';
COMMENT ON COLUMN users.balance IS '余额（内部单位，1美元=1,000,000）';
COMMENT ON COLUMN users.group_name IS '用户组（用于差异化定价）';

-- 2. API Key表
CREATE TABLE IF NOT EXISTS api_keys (
    id              VARCHAR(36) PRIMARY KEY,
    user_id         VARCHAR(36) REFERENCES users(id) ON DELETE CASCADE,
    key_hash        VARCHAR(64) UNIQUE NOT NULL,
    key_prefix      VARCHAR(12) NOT NULL,
    name            VARCHAR(128),
    status          SMALLINT DEFAULT 1,
    quota_limit     BIGINT DEFAULT -1,
    used_quota      BIGINT DEFAULT 0,
    rate_limit      INT DEFAULT -1,
    allowed_models  JSONB,
    allowed_ips     JSONB,
    expires_at      TIMESTAMPTZ,
    last_used_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE api_keys IS 'API Key表';
COMMENT ON COLUMN api_keys.key_hash IS 'SHA-256 哈希，不存明文';
COMMENT ON COLUMN api_keys.quota_limit IS '额度限制，-1=不限制';
COMMENT ON COLUMN api_keys.rate_limit IS '每分钟请求限制，-1=不限制';

-- 3. 供应商渠道表
CREATE TABLE IF NOT EXISTS channels (
    id              VARCHAR(36) PRIMARY KEY,
    name            VARCHAR(128) NOT NULL,
    type            SMALLINT NOT NULL,
    base_url        VARCHAR(512) NOT NULL,
    api_key_enc     TEXT NOT NULL,
    status          SMALLINT DEFAULT 1,
    priority        INT DEFAULT 0,
    weight          INT DEFAULT 1,
    balance         DECIMAL(12,4),
    models          JSONB NOT NULL,
    model_mapping   JSONB,
    groups          JSONB DEFAULT '["default"]',
    used_quota      BIGINT DEFAULT 0,
    request_count   BIGINT DEFAULT 0,
    success_count   BIGINT DEFAULT 0,
    config          JSONB,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE channels IS '供应商渠道表';
COMMENT ON COLUMN channels.type IS '渠道类型: 1=OpenAI, 2=Claude, 3=Gemini, 4=DeepSeek';
COMMENT ON COLUMN channels.priority IS '优先级，越大越优先';
COMMENT ON COLUMN channels.weight IS '同优先级内的权重';

-- 4. 模型管理表
CREATE TABLE IF NOT EXISTS models (
    id                  VARCHAR(36) PRIMARY KEY,
    channel_id          VARCHAR(36) NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    model_name          VARCHAR(128) NOT NULL,
    prompt_price        DECIMAL(16,8) NOT NULL,
    prompt_unit         INT NOT NULL DEFAULT 1000000,
    cached_prompt_price DECIMAL(16,8) NOT NULL DEFAULT 0,
    completion_price    DECIMAL(16,8) NOT NULL,
    completion_unit     INT NOT NULL DEFAULT 1000000,
    image_price         DECIMAL(16,8),
    audio_price         DECIMAL(16,8),
    currency            VARCHAR(3) DEFAULT 'USD',
    enabled             BOOLEAN DEFAULT TRUE,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(channel_id, model_name)
);

COMMENT ON TABLE models IS '模型管理表';
COMMENT ON COLUMN models.prompt_price IS '输入价格（原始数字）';
COMMENT ON COLUMN models.cached_prompt_price IS '缓存输入价格（原始数字）';
COMMENT ON COLUMN models.prompt_unit IS '输入价格对应的token数量';

-- 5. 请求日志表
CREATE TABLE IF NOT EXISTS request_logs (
    id              VARCHAR(36) PRIMARY KEY,
    user_id         VARCHAR(36) REFERENCES users(id),
    api_key_id      VARCHAR(36) REFERENCES api_keys(id),
    channel_id      VARCHAR(36) REFERENCES channels(id),
    model           VARCHAR(128) NOT NULL,
    prompt_tokens   INT DEFAULT 0,
    completion_tokens INT DEFAULT 0,
    total_tokens    INT DEFAULT 0,
    reasoning_tokens INT DEFAULT 0,
    cost            BIGINT DEFAULT 0,
    cache_hit       BOOLEAN DEFAULT FALSE,
    cache_tokens    INT DEFAULT 0,
    status_code     SMALLINT,
    is_stream       BOOLEAN DEFAULT FALSE,
    latency_ms      INT DEFAULT 0,
    error_message   TEXT,
    request_id      VARCHAR(64),
    ip_address      VARCHAR(45),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE request_logs IS '请求日志表';
COMMENT ON COLUMN request_logs.cost IS '花费（内部单位，1美元=1,000,000）';

-- 6. 交易记录表
CREATE TABLE IF NOT EXISTS transactions (
    id              VARCHAR(36) PRIMARY KEY,
    user_id         VARCHAR(36) REFERENCES users(id),
    type            SMALLINT NOT NULL,
    amount          BIGINT NOT NULL,
    balance_after   BIGINT NOT NULL,
    reference_type  VARCHAR(32),
    reference_id    VARCHAR(128),
    description     TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE transactions IS '交易记录表';
COMMENT ON COLUMN transactions.type IS '类型: 1=充值, 2=消费, 3=退款, 4=赠送';

-- 7. 充值码表
CREATE TABLE IF NOT EXISTS redemption_codes (
    id              VARCHAR(36) PRIMARY KEY,
    code            VARCHAR(32) UNIQUE NOT NULL,
    quota           BIGINT NOT NULL,
    used_by         VARCHAR(36) REFERENCES users(id),
    used_at         TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ,
    created_by      VARCHAR(36) REFERENCES users(id),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE redemption_codes IS '充值码表';

-- 8. 系统配置表
CREATE TABLE IF NOT EXISTS system_configs (
    key             VARCHAR(64) PRIMARY KEY,
    value           TEXT NOT NULL,
    description     TEXT,
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE system_configs IS '系统配置表';

-- 索引
CREATE INDEX IF NOT EXISTS idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_user ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_request_logs_user_time ON request_logs(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_request_logs_channel ON request_logs(channel_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_request_logs_model ON request_logs(model, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_user ON transactions(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_channels_status ON channels(status, priority DESC, weight DESC);
CREATE INDEX IF NOT EXISTS idx_models_channel_model ON models(channel_id, model_name);
CREATE INDEX IF NOT EXISTS idx_redemption_codes ON redemption_codes(code);

-- 初始管理员用户
INSERT INTO users (id, username, email, password_hash, role, balance, status)
VALUES (
    '01934b5c-7e8a-7c10-9a5f-8b2d3e4f5a6b',
    'admin',
    'admin@example.com',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', -- admin123456
    10,
    1000000000, -- $1000
    1
) ON CONFLICT DO NOTHING;

-- 初始系统配置
INSERT INTO system_configs (key, value, description) VALUES
    ('site_name', 'LLM Gateway', '站点名称'),
    ('register_enabled', 'true', '是否开放注册'),
    ('default_quota', '1000000', '新用户默认额度（$1）')
ON CONFLICT DO NOTHING;
