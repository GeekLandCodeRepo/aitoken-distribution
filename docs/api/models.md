# 模型管理模块

## 模型列表（管理员）

### GET /api/v1/admin/models

**查询参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| channel_id | uuid | 否 | 渠道 ID 筛选 |
| enabled | bool | 否 | 启用状态筛选 |
| search | string | 否 | 模型名称搜索 |

## 创建模型（管理员）

### POST /api/v1/admin/models

**请求体**

```json
{
  "channel_id": "01934b5c-7e8a-7c10-9a5f-8b2d3e4f5a6b",
  "model_name": "gpt-4o",
  "prompt_price": 2.5,
  "prompt_unit": 1000000,
  "cached_prompt_price": 0.5,
  "completion_price": 10,
  "completion_unit": 1000000,
  "currency": "USD",
  "enabled": true
}
```

## 更新模型（管理员）

### PUT /api/v1/admin/models/{id}

## 启用/禁用模型（管理员）

### POST /api/v1/admin/models/{id}/toggle

```json
{
  "enabled": false
}
```

## 删除模型（管理员）

### DELETE /api/v1/admin/models/{id}

**响应 200**

```json
{
  "code": 0,
  "data": {
    "message": "model deleted successfully"
  }
}
```
