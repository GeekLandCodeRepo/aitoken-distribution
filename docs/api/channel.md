# 渠道模块

## 渠道列表（管理员）

### GET /api/v1/admin/channels

**查询参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| size | int | 否 | 每页数量，默认20 |
| status | int | 否 | 状态筛选（0/1/2） |
| type | int | 否 | 类型筛选（1/2/3/4...） |

**响应 200**
```json
{
  "code": 0,
  "data": {
    "total": 5,
    "page": 1,
    "size": 20,
    "items": [
      {
        "id": "01934b5c-7e8a-7c10-9a5f-8b2d3e4f5a6b",
        "name": "OpenAI Official",
        "type": 1,
        "type_name": "OpenAI",
        "base_url": "https://api.openai.com/v1",
        "status": 1,
        "priority": 10,
        "weight": 1,
        "balance": 50.00,
        "models": ["gpt-4o", "gpt-4o-mini", "gpt-3.5-turbo"],
        "groups": ["default", "vip"],
        "used_quota": 12500000,
        "request_count": 5000,
        "success_count": 4975,
        "success_rate": 0.995,
        "created_at": "2024-01-10T08:00:00Z"
      }
    ]
  }
}
```

---

## 创建渠道（管理员）

### POST /api/v1/admin/channels

**请求体**
```json
{
  "name": "OpenAI Official",
  "type": 1,
  "base_url": "https://api.openai.com/v1",
  "api_key": "sk-proj-xxxxx",
  "models": ["gpt-4o", "gpt-4o-mini", "gpt-3.5-turbo"],
  "model_mapping": {
    "gpt-4": "gpt-4-turbo"
  },
  "priority": 10,
  "weight": 1,
  "groups": ["default", "vip"],
  "config": {
    "api_version": "2024-02-15-preview"
  }
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 渠道名称 |
| type | int | 是 | 渠道类型（1=OpenAI, 2=Claude, 3=Gemini, 4=DeepSeek） |
| base_url | string | 是 | API基础URL |
| api_key | string | 是 | API Key（加密存储） |
| models | []string | 是 | 支持的模型列表 |
| model_mapping | map | 否 | 模型名映射 |
| priority | int | 否 | 优先级，默认0 |
| weight | int | 否 | 权重，默认1 |
| groups | []string | 否 | 适用用户组，默认["default"] |
| config | object | 否 | 供应商特定配置 |

**响应 201**
```json
{
  "code": 0,
  "data": {
    "id": "01934b5c-7e8a-7c10-9a5f-8b2d3e4f5a6b",
    "name": "OpenAI Official",
    "type": 1,
    "status": 1,
    "priority": 10,
    "weight": 1,
    "models": ["gpt-4o", "gpt-4o-mini", "gpt-3.5-turbo"],
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**错误码**
| 错误码 | 说明 |
|--------|------|
| 500101 | 渠道名称为空 |
| 500102 | 渠道类型无效 |
| 500103 | Base URL 格式无效 |
| 500104 | API Key 为空 |
| 500105 | 模型列表为空 |

---

## 更新渠道（管理员）

### PUT /api/v1/admin/channels/{id}

**路径参数**
| 参数 | 类型 | 说明 |
|------|------|------|
| id | uuid | 渠道ID |

**请求体**
```json
{
  "name": "OpenAI Official Updated",
  "priority": 20,
  "weight": 2,
  "models": ["gpt-4o", "gpt-4o-mini"],
  "status": 1
}
```

**响应 200**
```json
{
  "code": 0,
  "data": {
    "id": "01934b5c-7e8a-7c10-9a5f-8b2d3e4f5a6b",
    "name": "OpenAI Official Updated",
    "priority": 20,
    "weight": 2,
    "status": 1
  }
}
```

**错误码**
| 错误码 | 说明 |
|--------|------|
| 500301 | 渠道不存在 |

---

## 删除渠道（管理员）

### DELETE /api/v1/admin/channels/{id}

**路径参数**
| 参数 | 类型 | 说明 |
|------|------|------|
| id | uuid | 渠道ID |

**响应 200**
```json
{
  "code": 0,
  "data": {
    "message": "channel deleted successfully"
  }
}
```

**错误码**
| 错误码 | 说明 |
|--------|------|
| 500301 | 渠道不存在 |
| 500404 | 渠道正在使用中 |

---

## 测试渠道（管理员）

### POST /api/v1/admin/channels/{id}/test

**路径参数**
| 参数 | 类型 | 说明 |
|------|------|------|
| id | uuid | 渠道ID |

**响应 200**
```json
{
  "code": 0,
  "data": {
    "success": true,
    "latency_ms": 285,
    "tested_at": "2024-01-15T10:30:00Z"
  }
}
```

**错误码**
| 错误码 | 说明 |
|--------|------|
| 500301 | 渠道不存在 |
| 500401 | 渠道测试失败 |
| 500402 | 渠道连接超时 |
| 500403 | 渠道认证失败 |

---

## 启用/禁用渠道（管理员）

### POST /api/v1/admin/channels/{id}/toggle

**路径参数**
| 参数 | 类型 | 说明 |
|------|------|------|
| id | uuid | 渠道ID |

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

---

## 批量创建渠道（管理员）

### POST /api/v1/admin/channels/batch

**请求体**
```json
{
  "channels": [
    {
      "name": "OpenAI Key 1",
      "type": 1,
      "base_url": "https://api.openai.com/v1",
      "api_key": "sk-proj-key1",
      "models": ["gpt-4o"]
    },
    {
      "name": "OpenAI Key 2",
      "type": 1,
      "base_url": "https://api.openai.com/v1",
      "api_key": "sk-proj-key2",
      "models": ["gpt-4o"]
    }
  ]
}
```

**响应 201**
```json
{
  "code": 0,
  "data": {
    "created": 2,
    "channels": [
      {"id": "...", "name": "OpenAI Key 1"},
      {"id": "...", "name": "OpenAI Key 2"}
    ]
  }
}
```

**错误码**
| 错误码 | 说明 |
|--------|------|
| 500405 | 批量创建数量超限（>100） |
