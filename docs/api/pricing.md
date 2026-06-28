# 定价模块

## 定价列表（管理员）

### GET /api/v1/admin/pricing

**查询参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| channel_type | int | 否 | 渠道类型筛选 |
| enabled | bool | 否 | 启用状态筛选 |
| search | string | 否 | 模型名称搜索 |

**响应 200**
```json
{
  "code": 0,
  "data": [
    {
      "id": "01934b5c-7e8a-7c10-9a5f-8b2d3e4f5a6b",
      "model_name": "gpt-4o",
      "channel_type": 1,
      "prompt_price": 2.50,
      "prompt_unit": 1000000,
      "completion_price": 10.00,
      "completion_unit": 1000000,
      "image_price": null,
      "audio_price": null,
      "cache_ratio": 0.5,
      "currency": "USD",
      "enabled": true,
      "created_at": "2024-01-10T08:00:00Z"
    }
  ]
}
```

---

## 创建定价（管理员）

### POST /api/v1/admin/pricing

**请求体**
```json
{
  "model_name": "gpt-4o",
  "channel_type": 1,
  "prompt_price": 2.50,
  "prompt_unit": 1000000,
  "completion_price": 10.00,
  "completion_unit": 1000000,
  "cache_ratio": 0.5,
  "currency": "USD"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| model_name | string | 是 | 模型名称 |
| channel_type | int | 是 | 渠道类型 |
| prompt_price | decimal | 是 | 输入价格 |
| prompt_unit | int | 是 | 输入价格对应的token数 |
| completion_price | decimal | 是 | 输出价格 |
| completion_unit | int | 是 | 输出价格对应的token数 |
| image_price | decimal | 否 | 图片价格（每张） |
| audio_price | decimal | 否 | 音频价格（每分钟） |
| cache_ratio | decimal | 否 | 缓存折扣，默认0.5 |
| currency | string | 否 | 货币，默认USD |

**计费单位示例**
| 供应商 | prompt_unit | completion_unit |
|--------|-------------|-----------------|
| OpenAI | 1000000 | 1000000 |
| Claude | 1000000 | 1000000 |
| Gemini | 1000 | 1000 |
| DeepSeek | 1000000 | 1000000 |
| 国产模型 | 200000000 | 200000000 |

**响应 201**
```json
{
  "code": 0,
  "data": {
    "id": "01934b5c-7e8a-7c10-9a5f-8b2d3e4f5a6b",
    "model_name": "gpt-4o",
    "channel_type": 1,
    "prompt_price": 2.50,
    "prompt_unit": 1000000,
    "completion_price": 10.00,
    "completion_unit": 1000000,
    "enabled": true
  }
}
```

**错误码**
| 错误码 | 说明 |
|--------|------|
| 600101 | 模型名称为空 |
| 600102 | 渠道类型无效 |
| 600103 | 输入价格无效 |
| 600104 | 输出价格无效 |
| 600302 | 模型定价已存在 |

---

## 更新定价（管理员）

### PUT /api/v1/admin/pricing/{id}

**路径参数**
| 参数 | 类型 | 说明 |
|------|------|------|
| id | uuid | 定价ID |

**请求体**
```json
{
  "prompt_price": 3.00,
  "completion_price": 12.00,
  "cache_ratio": 0.4,
  "enabled": true
}
```

**响应 200**
```json
{
  "code": 0,
  "data": {
    "id": "01934b5c-7e8a-7c10-9a5f-8b2d3e4f5a6b",
    "model_name": "gpt-4o",
    "prompt_price": 3.00,
    "completion_price": 12.00,
    "cache_ratio": 0.4,
    "enabled": true
  }
}
```

**错误码**
| 错误码 | 说明 |
|--------|------|
| 600301 | 定价记录不存在 |

---

## 删除定价（管理员）

### DELETE /api/v1/admin/pricing/{id}

**路径参数**
| 参数 | 类型 | 说明 |
|------|------|------|
| id | uuid | 定价ID |

**响应 200**
```json
{
  "code": 0,
  "data": {
    "message": "pricing deleted successfully"
  }
}
```

---

## 从渠道同步模型（管理员）

### POST /api/v1/admin/pricing/sync

**描述**: 从所有启用的渠道中提取模型列表，自动创建缺失的定价记录。

**响应 200**
```json
{
  "code": 0,
  "data": {
    "synced": 5,
    "created": ["gpt-4o-mini", "claude-3-5-sonnet"],
    "skipped": ["gpt-4o", "gpt-3.5-turbo"]
  }
}
```

**错误码**
| 错误码 | 说明 |
|--------|------|
| 600401 | 同步失败（无可用渠道） |
