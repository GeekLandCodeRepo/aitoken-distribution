package migration

import (
	"fmt"
	"time"

	"github.com/lib/pq"
	"xorm.io/xorm"

	apikeyDomain "llm-gateway/internal/apikey/domain"
	billingDomain "llm-gateway/internal/billing/domain"
	channelDomain "llm-gateway/internal/channel/domain"
	pricingDomain "llm-gateway/internal/pricing/domain"
	"llm-gateway/internal/shared/config"
	"llm-gateway/internal/shared/crypto"
	"llm-gateway/internal/shared/uuid"
	userDomain "llm-gateway/internal/user/domain"
)

// Up applies schema synchronization and idempotent schema patches.
func Up(db *xorm.Engine, cfg *config.Config) error {
	if err := syncTables(db); err != nil {
		return err
	}
	if err := applySchemaPatches(db); err != nil {
		return err
	}
	if err := seedExampleChannels(db, cfg); err != nil {
		return err
	}
	return nil
}

func syncTables(db *xorm.Engine) error {
	return db.Sync2(
		&userDomain.User{},
		&userDomain.InviteCode{},
		&userDomain.AppSetting{},
		&userDomain.UserChannelPermission{},
		&apikeyDomain.ApiKey{},
		&channelDomain.Channel{},
		&pricingDomain.Pricing{},
		&billingDomain.Transaction{},
		&billingDomain.RequestLog{},
		&billingDomain.RedeemCode{},
	)
}

func applySchemaPatches(db *xorm.Engine) error {
	statements := []string{
		"ALTER TABLE model_pricing ADD COLUMN IF NOT EXISTS cached_prompt_price DECIMAL(16,8) DEFAULT 0",
		"DROP INDEX IF EXISTS user_channel_permissions_user_channel_uidx",
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'user_channel_permissions_user_id_fkey') THEN
				ALTER TABLE user_channel_permissions
				ADD CONSTRAINT user_channel_permissions_user_id_fkey
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
			END IF;
		END $$`,
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'user_channel_permissions_channel_id_fkey') THEN
				ALTER TABLE user_channel_permissions
				ADD CONSTRAINT user_channel_permissions_channel_id_fkey
				FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE;
			END IF;
		END $$`,
	}

	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("apply schema statement %q: %w", statement, err)
		}
	}
	if err := cleanupLegacyColumns(db); err != nil {
		return err
	}
	if err := applySchemaComments(db); err != nil {
		return err
	}
	return nil
}

func applySchemaComments(db *xorm.Engine) error {
	tableComments := map[string]string{
		"users":                    "用户表",
		"invite_codes":             "邀请码表",
		"app_settings":             "系统设置表",
		"user_channel_permissions": "用户可用渠道白名单表",
		"api_keys":                 "API密钥表",
		"channels":                 "渠道表",
		"model_pricing":            "模型定价表",
		"transactions":             "交易流水表",
		"request_logs":             "请求日志表",
		"redemption_codes":         "充值码表",
	}
	for table, comment := range tableComments {
		if err := commentTable(db, table, comment); err != nil {
			return err
		}
	}

	columnComments := map[string]map[string]string{
		"users": {
			"id":            "用户ID",
			"username":      "用户名",
			"email":         "邮箱",
			"password_hash": "Argon2id密码哈希",
			"role":          "角色：1普通用户，10管理员",
			"balance":       "余额，内部额度单位",
			"used_quota":    "已用额度，内部额度单位",
			"request_count": "请求次数",
			"status":        "状态：1启用，0禁用",
			"group_name":    "用户组",
			"created_at":    "创建时间",
			"updated_at":    "更新时间",
		},
		"invite_codes": {
			"id":              "邀请码ID",
			"code":            "邀请码",
			"inviter_user_id": "邀请人用户ID",
			"used_by_user_id": "使用人用户ID",
			"used_at":         "使用时间",
			"reward_amount":   "邀请人返现金额，内部额度单位",
			"new_user_bonus":  "新用户赠送金额，内部额度单位",
			"created_at":      "创建时间",
		},
		"app_settings": {
			"key":        "设置键",
			"value":      "设置值",
			"updated_at": "更新时间",
		},
		"user_channel_permissions": {
			"id":         "用户渠道权限ID",
			"user_id":    "用户ID",
			"channel_id": "渠道ID",
			"created_at": "创建时间",
		},
		"api_keys": {
			"id":             "API密钥ID",
			"user_id":        "所属用户ID",
			"key_hash":       "API密钥哈希",
			"key_prefix":     "密钥前缀",
			"key_suffix":     "密钥后缀",
			"name":           "密钥名称",
			"status":         "状态：1启用，0禁用",
			"quota_limit":    "额度限制，-1表示不限制",
			"used_quota":     "已用额度，内部额度单位",
			"rate_limit":     "速率限制，-1表示不限制",
			"allowed_models": "允许使用的模型列表JSON",
			"allowed_ips":    "允许访问的IP列表JSON",
			"expires_at":     "过期时间",
			"last_used_at":   "最后使用时间",
			"created_at":     "创建时间",
		},
		"channels": {
			"id":            "渠道ID",
			"name":          "渠道名称",
			"type":          "渠道类型：1 OpenAI，2 Claude，3 Gemini，4 DeepSeek",
			"base_url":      "上游基础地址",
			"api_key_enc":   "加密后的上游API Key",
			"status":        "状态：1启用，0禁用",
			"priority":      "优先级",
			"weight":        "权重",
			"balance":       "渠道余额展示值",
			"models":        "支持的模型列表JSON",
			"model_mapping": "模型映射JSON",
			"groups":        "可用用户组JSON",
			"used_quota":    "渠道已消耗额度，内部额度单位",
			"request_count": "渠道请求次数",
			"success_count": "渠道成功请求次数",
			"config":        "渠道扩展配置JSON",
			"created_at":    "创建时间",
			"updated_at":    "更新时间",
		},
		"model_pricing": {
			"id":                  "定价ID",
			"channel_id":          "所属渠道ID",
			"model_name":          "模型名称",
			"prompt_price":        "输入价格",
			"prompt_unit":         "输入价格计价单位token数",
			"cached_prompt_price": "缓存命中输入价格",
			"completion_price":    "输出价格",
			"completion_unit":     "输出价格计价单位token数",
			"image_price":         "图片价格",
			"audio_price":         "音频价格",
			"currency":            "币种",
			"enabled":             "是否启用该定价",
			"created_at":          "创建时间",
			"updated_at":          "更新时间",
		},
		"transactions": {
			"id":             "交易ID",
			"i_d":            "旧兼容交易ID字段",
			"user_id":        "用户ID",
			"type":           "交易类型：1充值，2消费，3退款，4赠送",
			"amount":         "交易金额，内部额度单位",
			"balance_after":  "交易后余额，内部额度单位",
			"reference_type": "关联对象类型",
			"reference_id":   "关联对象ID",
			"reference_i_d":  "旧兼容关联对象ID字段",
			"description":    "交易描述",
			"created_at":     "创建时间",
		},
		"request_logs": {
			"id":                "请求日志ID",
			"i_d":               "旧兼容请求日志ID字段",
			"user_id":           "用户ID",
			"api_key_id":        "API密钥ID",
			"channel_id":        "渠道ID",
			"endpoint":          "请求端点",
			"model":             "模型名称",
			"prompt_tokens":     "输入token数",
			"completion_tokens": "输出token数",
			"total_tokens":      "总token数",
			"cost":              "请求费用，内部额度单位",
			"cache_hit":         "是否命中缓存",
			"cache_tokens":      "缓存命中输入token数",
			"status_code":       "响应状态码",
			"is_stream":         "是否流式请求",
			"first_byte_ms":     "首字节耗时毫秒",
			"latency_ms":        "总耗时毫秒",
			"error_message":     "错误信息",
			"request_id":        "请求ID",
			"request_i_d":       "旧兼容请求ID字段",
			"ip_address":        "客户端IP地址",
			"i_p_address":       "旧兼容客户端IP地址字段",
			"created_at":        "创建时间",
		},
		"redemption_codes": {
			"id":         "充值码ID",
			"i_d":        "旧兼容充值码ID字段",
			"code":       "充值码",
			"quota":      "可兑换额度，内部额度单位",
			"used_by":    "使用人用户ID",
			"used_at":    "使用时间",
			"expires_at": "过期时间",
			"created_by": "创建人用户ID",
			"created_at": "创建时间",
		},
	}
	for table, comments := range columnComments {
		for column, comment := range comments {
			if err := commentColumn(db, table, column, comment); err != nil {
				return err
			}
		}
	}
	return nil
}

func commentTable(db *xorm.Engine, table string, comment string) error {
	exists, err := tableExists(db, table)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	statement := fmt.Sprintf("COMMENT ON TABLE %s IS %s", pq.QuoteIdentifier(table), pq.QuoteLiteral(comment))
	if _, err := db.Exec(statement); err != nil {
		return fmt.Errorf("comment table %s: %w", table, err)
	}
	return nil
}

func commentColumn(db *xorm.Engine, table string, column string, comment string) error {
	exists, err := columnExists(db, table, column)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	statement := fmt.Sprintf("COMMENT ON COLUMN %s.%s IS %s", pq.QuoteIdentifier(table), pq.QuoteIdentifier(column), pq.QuoteLiteral(comment))
	if _, err := db.Exec(statement); err != nil {
		return fmt.Errorf("comment column %s.%s: %w", table, column, err)
	}
	return nil
}

func tableExists(db *xorm.Engine, table string) (bool, error) {
	var exists bool
	_, err := db.SQL("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = ?)", table).Get(&exists)
	if err != nil {
		return false, fmt.Errorf("check table %s: %w", table, err)
	}
	return exists, nil
}

func columnExists(db *xorm.Engine, table string, column string) (bool, error) {
	var exists bool
	_, err := db.SQL("SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = ? AND column_name = ?)", table, column).Get(&exists)
	if err != nil {
		return false, fmt.Errorf("check column %s.%s: %w", table, column, err)
	}
	return exists, nil
}

func cleanupLegacyColumns(db *xorm.Engine) error {
	legacyCopies := []struct {
		table  string
		legacy string
		target string
	}{
		{table: "request_logs", legacy: "request_i_d", target: "request_id"},
		{table: "request_logs", legacy: "i_p_address", target: "ip_address"},
		{table: "transactions", legacy: "reference_i_d", target: "reference_id"},
	}
	for _, item := range legacyCopies {
		if err := copyLegacyColumn(db, item.table, item.legacy, item.target); err != nil {
			return err
		}
	}

	legacyColumns := []struct {
		table  string
		column string
	}{
		{table: "redemption_codes", column: "i_d"},
		{table: "request_logs", column: "i_d"},
		{table: "request_logs", column: "request_i_d"},
		{table: "request_logs", column: "i_p_address"},
		{table: "transactions", column: "i_d"},
		{table: "transactions", column: "reference_i_d"},
	}
	for _, item := range legacyColumns {
		statement := fmt.Sprintf("ALTER TABLE %s DROP COLUMN IF EXISTS %s", pq.QuoteIdentifier(item.table), pq.QuoteIdentifier(item.column))
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("drop legacy column %s.%s: %w", item.table, item.column, err)
		}
	}
	return nil
}

func copyLegacyColumn(db *xorm.Engine, table string, legacyColumn string, targetColumn string) error {
	legacyExists, err := columnExists(db, table, legacyColumn)
	if err != nil {
		return err
	}
	if !legacyExists {
		return nil
	}
	targetExists, err := columnExists(db, table, targetColumn)
	if err != nil {
		return err
	}
	if !targetExists {
		return nil
	}
	statement := fmt.Sprintf(
		"UPDATE %s SET %s = %s WHERE (%s IS NULL OR %s = '') AND %s IS NOT NULL AND %s <> ''",
		pq.QuoteIdentifier(table),
		pq.QuoteIdentifier(targetColumn),
		pq.QuoteIdentifier(legacyColumn),
		pq.QuoteIdentifier(targetColumn),
		pq.QuoteIdentifier(targetColumn),
		pq.QuoteIdentifier(legacyColumn),
		pq.QuoteIdentifier(legacyColumn),
	)
	if _, err := db.Exec(statement); err != nil {
		return fmt.Errorf("copy legacy column %s.%s to %s: %w", table, legacyColumn, targetColumn, err)
	}
	return nil
}

func seedExampleChannels(db *xorm.Engine, cfg *config.Config) error {
	const channelName = "DeepSeek官方"
	const placeholderAPIKey = "replace-with-your-deepseek-api-key"

	modelsJSON := `["deepseek-v4-flash","deepseek-v4-pro"]`
	groupsJSON := `["default"]`

	var channel channelDomain.Channel
	has, err := db.Where("name = ?", channelName).Get(&channel)
	if err != nil {
		return fmt.Errorf("check example channel: %w", err)
	}
	if !has {
		keyCrypto := crypto.NewChaCha20Poly1305Crypto(cfg.ChaCha20Poly1305Key)
		apiKeyEnc, err := keyCrypto.Encrypt(placeholderAPIKey)
		if err != nil {
			return fmt.Errorf("encrypt example channel api key: %w", err)
		}

		now := time.Now()
		channel = channelDomain.Channel{
			ID:        uuid.NewV7String(),
			Name:      channelName,
			Type:      4,
			BaseURL:   "https://api.deepseek.com",
			APIKeyEnc: apiKeyEnc,
			Status:    0,
			Priority:  0,
			Weight:    1,
			Models:    modelsJSON,
			Groups:    groupsJSON,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if _, err := db.Insert(&channel); err != nil {
			return fmt.Errorf("seed example channel: %w", err)
		}
	}

	if err := seedModelPricing(db, channel.ID, "deepseek-v4-flash", 0.14000000, 0.00280000, 0.28000000); err != nil {
		return err
	}
	if err := seedModelPricing(db, channel.ID, "deepseek-v4-pro", 0.43500000, 0.00362500, 0.87000000); err != nil {
		return err
	}
	return nil
}

func seedModelPricing(db *xorm.Engine, channelID string, modelName string, promptPrice float64, cachedPromptPrice float64, completionPrice float64) error {
	var existing pricingDomain.Pricing
	has, err := db.Where("channel_id = ? AND model_name = ?", channelID, modelName).Get(&existing)
	if err != nil {
		return fmt.Errorf("check example pricing %s: %w", modelName, err)
	}
	if has {
		return nil
	}

	now := time.Now()
	pricing := pricingDomain.Pricing{
		ID:                uuid.NewV7String(),
		ChannelID:         channelID,
		ModelName:         modelName,
		PromptPrice:       promptPrice,
		PromptUnit:        1000000,
		CachedPromptPrice: cachedPromptPrice,
		CompletionPrice:   completionPrice,
		CompletionUnit:    1000000,
		Currency:          "USD",
		Enabled:           true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if _, err := db.Insert(&pricing); err != nil {
		return fmt.Errorf("seed example pricing %s: %w", modelName, err)
	}
	return nil
}
