# AiToken 分发站

AiToken 分发站是一个面向团队、组织和个人开发者的 LLM API 分发与管理平台。它把多个上游大模型渠道统一封装成 OpenAI 兼容接口，同时提供用户、API Key、渠道、模型管理、额度计费、请求日志和权限控制能力，适合用来搭建私有的模型 API 网关。

## 它解决什么问题

在实际使用大模型 API 时，团队通常会遇到这些问题：

1. **多个模型供应商接入方式不统一**
   - OpenAI、DeepSeek、Claude、Gemini 等供应商接口格式、鉴权方式和响应结构不同。
   - 应用侧如果直接对接多个供应商，会导致业务代码复杂、切换成本高。

2. **API Key 难以集中管理**
   - 多个用户共用上游 Key，容易泄露，也难以追踪是谁在消耗额度。
   - 缺少按用户、按 Key、按模型维度的权限控制和用量统计。

3. **成本不可控**
   - 不同模型输入、缓存命中输入、输出价格不同。
   - 流式请求、非流式请求、缓存 token 等场景都需要统一计费。
   - 团队需要知道每个用户、每个 API Key、每个模型到底花了多少钱。

4. **用户能看到/调用哪些模型需要管控**
   - 并不是所有用户都应该看到所有渠道和模型。
   - 平台需要支持“用户可用渠道”白名单，以及新用户默认可用渠道模板。

5. **请求排查困难**
   - 上游报错、超时、流式中断、计费异常等问题需要完整日志。
   - 管理员需要按用户、API Key、模型、状态码、时间等条件快速检索请求。

AiToken 分发站的目标是：

> 用一个统一的 OpenAI 兼容入口，集中管理上游渠道、用户权限、模型价格、余额消耗和请求审计。

## 核心能力

- **OpenAI 兼容接口**
  - 提供 `/v1/chat/completions` 和 `/v1/models`。
  - 应用侧可以按 OpenAI SDK 的方式接入。

- **多渠道管理**
  - 支持 OpenAI、DeepSeek、Claude、Gemini 类型渠道。
  - 支持渠道启用/禁用、优先级、权重、模型列表、模型映射。

- **模型广场**
  - 用户可以查看自己可用渠道下的模型和价格。
  - 模型广场会按用户可用渠道白名单过滤。

- **用户可用渠道控制**
  - 管理员可以给每个用户勾选可用渠道。
  - 未勾选的渠道不会出现在模型广场，也不能通过 `/v1/chat/completions` 调用。
  - 支持“默认渠道模板”，新注册用户会自动应用模板。
  - 支持一键把模板覆盖应用到所有用户。

- **API Key 管理**
  - 用户可创建自己的 API Key。
  - API Key 只展示前后缀，完整 Key 不落库。
  - 支持 Key 启用/禁用、额度限制、速率限制、过期时间。

- **额度与计费**
  - 支持用户余额、充值码、管理员充值。
  - 支持按渠道+模型配置价格。
  - 支持输入价格、缓存命中输入价格、输出价格。
  - 请求结束后按实际 usage 结算。

- **请求日志与统计**
  - 记录用户、API Key、渠道、模型、token、费用、状态码、延迟等信息。
  - 管理端支持全局日志检索和使用统计。
  - 用户端支持查看自己的使用趋势、模型分布和请求明细。

- **安全加密**
  - 用户密码使用 Argon2id 哈希。
  - 上游渠道 API Key 使用 ChaCha20-Poly1305 加密存储。
  - 后端服务启动不会自动修改数据库结构，数据库迁移由独立 CLI 执行。

## 技术架构

### 总体架构

```text
┌──────────────────────┐
│      React Web UI     │
│  管理后台 / 用户控制台 │
└───────────┬──────────┘
            │ /api/v1
┌───────────▼──────────┐
│       Go Backend      │
│  chi router + usecase │
└───────┬───────┬──────┘
        │       │
        │       └──────────────┐
        │                      │
┌───────▼───────┐      ┌───────▼───────┐
│  PostgreSQL   │      │     Redis     │
│  业务数据/日志 │      │ 缓存/限流/Stream │
└───────────────┘      └───────────────┘
        ▲
        │
┌───────┴────────┐
│ /v1 OpenAI API │
│ Relay Gateway  │
└───────┬────────┘
        │
        ▼
┌──────────────────────────────────────┐
│ OpenAI / DeepSeek / Claude / Gemini  │
└──────────────────────────────────────┘
```

### 后端架构

后端使用 Go 编写，按模块分层：

```text
cmd/
  server/          # HTTP 服务入口
  migrate/         # 数据库迁移 CLI
  worker/          # Redis Streams 异步日志/统计 Worker

internal/
  user/            # 用户、邀请、用户可用渠道模板
  apikey/          # API Key 管理
  channel/         # 上游渠道和模型广场
  pricing/         # 模型管理与价格配置
  billing/         # 余额、消费、充值码、交易流水
  relay/           # OpenAI 兼容转发、适配器、流式处理
  usage/           # 请求日志和使用统计
  shared/          # 配置、数据库、Redis、JWT、中间件、响应、加密
```

每个业务模块大致遵循：

```text
domain      # 数据模型和接口定义
repository  # 数据库访问
usecase     # 业务逻辑
handler     # HTTP 处理器
```

### 前端架构

前端使用 React + TypeScript + Vite：

```text
web/src/
  api/             # API 请求封装
  components/      # 布局和 shadcn 风格组件
  pages/           # 页面
  store/           # Zustand 状态
  locales/         # i18n 文案
  lib/             # alova、工具函数、常量
```

主要页面：

- 用户控制台
  - 仪表盘
  - API Key 管理
  - 模型广场
  - 用量统计
  - 充值码兑换

- 管理后台
  - 用户管理
  - 默认渠道模板
  - 渠道管理
  - 模型管理
  - 请求日志
  - 充值码管理
  - 全局统计

### 数据存储

核心数据表：

- `users`：用户信息、余额、角色、状态
- `api_keys`：用户 API Key 哈希、前后缀、额度限制
- `channels`：上游渠道配置和加密后的上游 API Key
- `models`：渠道+模型维度的模型配置和价格
- `user_channel_permissions`：用户可用渠道白名单
- `request_logs`：请求日志、token、费用、状态码、延迟
- `transactions`：交易流水
- `redemption_codes`：充值码
- `app_settings`：系统配置，例如默认渠道模板

### Relay 流程

```text
Client
  │ Authorization: Bearer sk-xxx
  ▼
APIKeyAuth
  │ 校验 API Key、状态、过期时间、额度、限流
  ▼
RelayUsecase.ChatCompletion
  │ 1. 根据用户白名单 + 模型选择可用渠道
  │ 2. 解密渠道 API Key
  │ 3. 预扣费
  │ 4. 适配/转发请求到上游
  │ 5. 处理流式或非流式响应
  │ 6. 按实际 usage 结算
  │ 7. 同步写交易流水，异步投递请求日志/渠道统计事件
  ▼
Upstream Provider
```

OpenAI 和 DeepSeek 属于 OpenAI 兼容渠道，转发时尽量透传请求和响应；Claude/Gemini 通过 adaptor 转换请求和响应结构。请求日志和渠道统计通过 Redis Streams 异步落库，余额扣减、API Key 用量和交易流水保持同步处理。

### 迁移机制

服务启动不会自动修改数据库结构。

数据库初始化、表结构同步、字段注释、示例渠道 seed 都由独立命令执行：

```bash
bin/aitsd-migrate
```

这样可以避免服务启动时隐式执行 DDL 或业务数据回填。

## 快速开始

### 1. 准备配置

复制环境变量示例：

```bash
cp .env.example .env
```

至少修改以下配置：

```env
DB_PASSWORD=your-db-password
JWT_SECRET=replace-with-random-secret
CHACHA20_POLY1305_KEY=replace-with-fixed-random-32-byte-key
ADMIN_PASSWORD=replace-with-admin-password
```

`CHACHA20_POLY1305_KEY` 必须长期固定；如果变更，已保存的渠道 API Key 将无法解密。

### 2. 启动依赖

```bash
docker compose up -d postgres redis
```

### 3. 构建后端

```bash
CGO_ENABLED=0 go build -o bin/aitsd ./cmd/server
CGO_ENABLED=0 go build -o bin/aitsd-migrate ./cmd/migrate
CGO_ENABLED=0 go build -o bin/aitsd-worker ./cmd/worker
```

### 4. 执行迁移

```bash
bin/aitsd-migrate
```

### 5. 构建前端

```bash
cd web
npm install
npm run build
```

### 6. 启动服务

```bash
bin/aitsd
# 另开进程处理请求日志和渠道统计异步事件
bin/aitsd-worker
```

Docker 部署资产位于 `deploy/`。Compose 中间件默认使用 `pgvector/pgvector:pg18-trixie` 和 `valkey/valkey:8-alpine`，应用侧只有一个 `app` 容器，会同时运行 API/Web Server 和 Worker，并在应用数据卷首次启动时自动执行一次迁移。容器网络为 `stone-net`。服务器部署时 `.env` 至少需要配置：

```env
DB_USER=postgres
DB_PASSWORD=your-postgres-password
DB_NAME=aitsd
REDIS_USERNAME=hdward
REDIS_PASSWORD=your-valkey-password
```

启动完整服务：

```bash
docker compose -f deploy/docker-compose.yml up -d
```

首次启动会执行：

```text
aitsd-migrate --skip-env
```

迁移完成后会在 `appdata` 卷中写入 `/srv/data/.migrated`，后续重启不会重复迁移。如需手动强制迁移，可临时设置 `FORCE_MIGRATE=true` 后重启 `app` 容器。

Web 静态页面使用相对目录 `web/dist` 托管。容器内工作目录是 `/srv`，因此实际目录为 `/srv/web/dist`；本地从仓库根目录启动时则使用 `web/dist`。

是否托管 Web 页面由环境变量控制，只使用 `true` / `false`，默认开启：

```env
SERVE_WEB=true
```

如果前面另有 Nginx/CDN 托管前端，只想让后端提供 API 和 Relay，可以设置：

```env
SERVE_WEB=false
```

默认服务端口：

```text
40680
```

## 使用流程

1. 管理员登录后台。
2. 在“渠道管理”中编辑示例渠道，填入真实上游 API Key 并启用。
3. 在“模型管理”中确认模型价格。
4. 在“用户管理 -> 默认渠道模板”中勾选新用户默认可用渠道。
5. 创建或注册用户。
6. 用户创建自己的 API Key。
7. 应用使用 OpenAI 兼容地址调用：

```text
http://<host>:40680/v1
```

示例：

```bash
curl http://localhost:40680/v1/chat/completions \
  -H "Authorization: Bearer sk-xxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "deepseek-v4-flash",
    "messages": [{"role":"user","content":"hello"}]
  }'
```

## 开发命令

后端：

```bash
CGO_ENABLED=0 go build ./...
CGO_ENABLED=0 go test ./...
```

前端：

```bash
cd web
npx tsc --noEmit
npm run build
```

## 注意事项

- `.env`、`bin/`、`web/dist/`、`web/node_modules/` 不应提交到仓库。
- 生产环境必须替换 `.env.example` 中的示例密钥和密码。
- 裸进程部署时服务启动不会自动迁移数据库，需要显式执行 `bin/aitsd-migrate`；Docker `app` 容器会在首次启动时自动执行一次迁移。
- 新用户能看到哪些模型取决于“默认渠道模板”和用户自己的“可用渠道”配置。
