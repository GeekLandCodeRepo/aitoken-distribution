# API Key 模块

## 我的 API Keys

### GET /api/v1/keys

**响应 200**
```json
{
  "code": 0,
  "data": [
    {
      "id": "01934b5c-7e8a-7c10-9a5f-8b2d3e4f5a6b",
      "name": "My Production Key",
      "key_prefix": "sk-abc1",
      "status": 1,
      "quota_limit": 100000000,
      "used_quota": 12500000,
      "rate_limit": 60,
      "allowed_models": ["gpt-4o", "gpt-3.5-turbo"],
      "allowed_ips": null,
      "expires_at": "2024-12-31T23:59:59Z",
      "last_used_at": "2024-01-15T10:30:00Z",
      "created_at": "2024-01-10T08:00:00Z"
    }
  ]
}
```

---

## 创建 API Key

### POST /api/v1/keys

**请求体**
```json
{
  "name": "My Production Key",
  "quota_limit": 100000000,
  "rate_limit": 60,
  "allowed_models": ["gpt-4o", "gpt-3.5-turbo"],
  "allowed_ips": ["192.168.1.0/24"],
  "expires_at": "2024-12-31T23:59:59Z"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | Key名称 |
| quota_limit | int64 | 否 | 额度限制，-1=不限制 |
| rate_limit | int | 否 | 每分钟请求限制，-1=不限制 |
| allowed_models | []string | 否 | 允许的模型，null=全部 |
| allowed_ips | []string | 否 | IP白名单，null=不限制 |
| expires_at | string | 否 | 过期时间，null=永不过期 |

**响应 201**
```json
{
  "code": 0,
  "data": {
    "id": "01934b5c-7e8a-7c10-9a5f-8b2d3e4f5a6b",
    "name": "My Production Key",
    "key": "sk-abc1234567890xyz",
    "key_prefix": "sk-abc1",
    "quota_limit": 100000000,
    "rate_limit": 60,
    "status": 1,
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**注意**: `key` 字段仅在创建时返回一次，后续无法获取明文。

**错误码**
| 错误码 | 说明 |
|--------|------|
| 400101 | Key 名称为空 |
| 400102 | Key 名称过长 |
| 400103 | 额度限制无效 |
| 400104 | 限流值无效 |
| 400105 | 模型列表格式无效 |
| 400106 | IP 白名单格式无效 |
| 400107 | 过期时间无效 |

---

## 更新 API Key

### PUT /api/v1/keys/{id}

**路径参数**
| 参数 | 类型 | 说明 |
|------|------|------|
| id | uuid | Key ID |

**请求体**
```json
{
  "name": "Updated Key Name",
  "quota_limit": 200000000,
  "rate_limit": 120,
  "allowed_models": ["gpt-4o", "gpt-4o-mini"],
  "status": 1
}
```

**响应 200**
```json
{
  "code": 0,
  "data": {
    "id": "01934b5c-7e8a-7c10-9a5f-8b2d3e4f5a6b",
    "name": "Updated Key Name",
    "key_prefix": "sk-abc1",
    "quota_limit": 200000000,
    "used_quota": 12500000,
    "rate_limit": 120,
    "status": 1
  }
}
```

**错误码**
| 错误码 | 说明 |
|--------|------|
| 400301 | API Key 不存在 |
| 400404 | Key 不属于当前用户 |

---

## 删除 API Key

### DELETE /api/v1/keys/{id}

**路径参数**
| 参数 | 类型 | 说明 |
|------|------|------|
| id | uuid | Key ID |

**响应 200**
```json
{
  "code": 0,
  "data": {
    "message": "api key deleted successfully"
  }
}
```

**错误码**
| 错误码 | 说明 |
|--------|------|
| 400301 | API Key 不存在 |
| 400404 | Key 不属于当前用户 |

---

## 启用/禁用 API Key

### POST /api/v1/keys/{id}/toggle

**路径参数**
| 参数 | 类型 | 说明 |
|------|------|------|
| id | uuid | Key ID |

**响应 200**
```json
{
  "code": 0,
  "data": {
    "id": "01934b5c-7e8a-7c10-9a5f-8b2d3e4f5a6b",
    "status": 0
  }
}
```

**错误码**
| 错误码 | 说明 |
|--------|------|
| 400301 | API Key 不存在 |
| 400404 | Key 不属于当前用户 |
