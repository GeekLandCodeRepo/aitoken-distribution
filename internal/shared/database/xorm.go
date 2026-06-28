package database

import (
	"fmt"
	"time"

	"github.com/lib/pq"
	"xorm.io/xorm"
	"xorm.io/xorm/names"

	"llm-gateway/internal/shared/config"
)

var DB *xorm.Engine

func EnsureDatabase(cfg *config.Config) error {
	initDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBSSLMode)

	initEngine, err := xorm.NewEngine("postgres", initDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer initEngine.Close()

	// 检查数据库是否存在，不存在则创建
	var exists bool
	_, err = initEngine.SQL("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = ?)", cfg.DBName).Get(&exists)
	if err != nil {
		return fmt.Errorf("failed to check database: %w", err)
	}
	if !exists {
		if _, err := initEngine.Exec("CREATE DATABASE " + pq.QuoteIdentifier(cfg.DBName)); err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
	}
	return nil
}

func Init(cfg *config.Config) error {
	// 连接到目标数据库
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	engine, err := xorm.NewEngine("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}
	DB = engine
	DB.SetColumnMapper(names.GonicMapper{})

	// 设置连接池
	DB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	DB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	DB.SetConnMaxLifetime(time.Hour)

	return nil
}

func GetDB() *xorm.Engine {
	return DB
}
