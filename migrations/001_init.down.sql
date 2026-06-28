-- LLM Gateway Database Migration Rollback
-- Version: 001
-- Description: Drop initial schema

DROP TABLE IF EXISTS redemption_codes;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS request_logs;
DROP TABLE IF EXISTS models;
DROP TABLE IF EXISTS channels;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS system_configs;
DROP TABLE IF EXISTS users;
