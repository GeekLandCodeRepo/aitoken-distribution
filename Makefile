.PHONY: build run dev clean test migrate-up migrate-down

# 变量
APP_NAME = llm-gateway
BUILD_DIR = ./bin
GO = go

# 构建
build:
	CGO_ENABLED=0 $(GO) build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

# 运行
run: build
	$(BUILD_DIR)/$(APP_NAME)

# 开发模式（热重载需要安装 air）
dev:
	air -c .air.toml

# 清理
clean:
	rm -rf $(BUILD_DIR)

# 测试
test:
	$(GO) test ./... -v

# 数据库迁移
migrate-up:
	migrate -path migrations -database "postgres://llm_gateway:llm_gateway_pass@localhost:5432/llm_gateway?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://llm_gateway:llm_gateway_pass@localhost:5432/llm_gateway?sslmode=disable" down

# Docker
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Go mod
tidy:
	$(GO) mod tidy

# 格式化
fmt:
	$(GO) fmt ./...

# 检查
vet:
	$(GO) vet ./...

# 前端
web-install:
	cd web && npm install

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build
