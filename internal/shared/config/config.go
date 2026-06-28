package config

import (
	"os"
	"strconv"
)

type Config struct {
	// 服务配置
	Port     string
	LogLevel string

	// 数据库配置
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	DBMaxOpenConns int
	DBMaxIdleConns int

	// Redis配置
	RedisHost     string
	RedisPort     string
	RedisUsername string
	RedisPassword string
	RedisDB       int

	// JWT配置
	JWTSecret        string
	JWTAccessExpiry  int
	JWTRefreshExpiry int

	// 计费配置
	QuotaPerUSD       int64
	PreConsumeEnabled bool

	// 加密配置
	ChaCha20Poly1305Key string

	// 管理员配置
	AdminEmail    string
	AdminPassword string
}

func Load() *Config {
	return &Config{
		Port:     getEnv("PORT", "40680"),
		LogLevel: getEnv("LOG_LEVEL", "info"),

		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "llm_gateway"),
		DBPassword:     getEnv("DB_PASSWORD", ""),
		DBName:         getEnv("DB_NAME", "llm_gateway"),
		DBSSLMode:      getEnv("DB_SSLMODE", "disable"),
		DBMaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 100),
		DBMaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 10),

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisUsername: getEnv("REDIS_USERNAME", ""),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		JWTSecret:        getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTAccessExpiry:  getEnvInt("JWT_ACCESS_EXPIRY", 7200),
		JWTRefreshExpiry: getEnvInt("JWT_REFRESH_EXPIRY", 604800),

		QuotaPerUSD:       int64(getEnvInt("QUOTA_PER_USD", 1000000)),
		PreConsumeEnabled: getEnvBool("PRE_CONSUME_ENABLED", true),

		ChaCha20Poly1305Key: getEnv("CHACHA20_POLY1305_KEY", "aitsd-chacha20-poly1305-key-32b"),

		AdminEmail:    getEnv("ADMIN_EMAIL", "admin@example.com"),
		AdminPassword: getEnv("ADMIN_PASSWORD", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
