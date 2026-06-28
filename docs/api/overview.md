# API 概述

## 基础信息

- Base URL: `/api/v1`
- 认证方式: Bearer Token (JWT)
- Content-Type: `application/json`

## 统一响应格式

### 成功响应
```json
{
  "code": 0,
  "data": { ... }
}
```

### 错误响应
```json
{
  "code": 40001,
  "message": "invalid parameter",
  "details": "email is required"
}
```

## 认证方式

### 1. JWT Token (用户端)
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

### 2. API Key (代理端)
```
Authorization: Bearer sk-xxxxxxxxxxxx
```

## 分页参数

```
GET /api/v1/users?page=1&size=20&sort=created_at&order=desc
```

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| size | int | 20 | 每页数量（最大100） |
| sort | string | created_at | 排序字段 |
| order | string | desc | 排序方式（asc/desc） |

### 分页响应
```json
{
  "code": 0,
  "data": {
    "total": 100,
    "page": 1,
    "size": 20,
    "items": [...]
  }
}
```

## 时间格式

所有时间字段使用 ISO 8601 格式：
```
2024-01-15T10:30:00Z
```

## 错误码

详见 [errors.md](../errors.md)

## API 模块

| 模块 | 文档 | 说明 |
|------|------|------|
| 认证 | [auth.md](./auth.md) | 登录、注册、Token刷新 |
| 用户 | [user.md](./user.md) | 用户管理、手动充值 |
| API Key | [key.md](./key.md) | Key的CRUD和配置 |
| 渠道 | [channel.md](./channel.md) | 供应商渠道管理 |
| 模型管理 | [models.md](./models.md) | 模型配置和价格配置 |
| 计费 | [billing.md](./billing.md) | 充值码和交易记录 |
| 代理 | [relay.md](./relay.md) | OpenAI兼容接口转发 |
| 用量 | [usage.md](./usage.md) | 统计和日志查询 |
