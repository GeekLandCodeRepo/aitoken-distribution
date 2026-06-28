# 计费模块

## 兑换充值码

### POST /api/v1/redeem

**请求体**
```json
{
  "code": "RC-ABC12345"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| code | string | 是 | 充值码 |

**响应 200**
```json
{
  "code": 0,
  "data": {
    "quota": 100000000,
    "balance_before": 50000000,
    "balance_after": 150000000
  }
}
```

**错误码**
| 错误码 | 说明 |
|--------|------|
| 700301 | 充值码不存在 |
| 700302 | 充值码已被使用 |
| 700303 | 充值码已过期 |

---

## 充值码列表（管理员）

### GET /api/v1/admin/redeem-codes

**查询参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| size | int | 否 | 每页数量，默认20 |
| status | string | 否 | 状态筛选（unused/used） |

**响应 200**
```json
{
  "code": 0,
  "data": {
    "total": 50,
    "page": 1,
    "size": 20,
    "items": [
      {
        "id": "01934b5c-7e8a-7c10-9a5f-8b2d3e4f5a6b",
        "code": "RC-ABC12345",
        "quota": 100000000,
        "used_by": null,
        "used_at": null,
        "expires_at": "2024-12-31T23:59:59Z",
        "created_by": "admin_user_id",
        "created_at": "2024-01-15T10:30:00Z"
      }
    ]
  }
}
```

---

## 生成充值码（管理员）

### POST /api/v1/admin/redeem-codes

**请求体**
```json
{
  "quota": 100000000,
  "count": 10,
  "expires_at": "2024-12-31T23:59:59Z"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| quota | int64 | 是 | 每个充值码的额度 |
| count | int | 是 | 生成数量（1-1000） |
| expires_at | string | 否 | 过期时间 |

**响应 201**
```json
{
  "code": 0,
  "data": {
    "count": 10,
    "codes": [
      "RC-ABC12345",
      "RC-DEF67890",
      "RC-GHI11223"
    ]
  }
}
```

**错误码**
| 错误码 | 说明 |
|--------|------|
| 700102 | 充值金额无效 |
| 700103 | 批量数量无效（1-1000） |

---

## 删除充值码（管理员）

### DELETE /api/v1/admin/redeem-codes/{id}

**路径参数**
| 参数 | 类型 | 说明 |
|------|------|------|
| id | uuid | 充值码ID |

**响应 200**
```json
{
  "code": 0,
  "data": {
    "message": "redeem code deleted successfully"
  }
}
```

---

## 交易记录

### GET /api/v1/transactions

**查询参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| size | int | 否 | 每页数量，默认20 |
| type | int | 否 | 类型筛选（1=充值, 2=消费, 3=退款, 4=赠送） |
| start | string | 否 | 开始时间 |
| end | string | 否 | 结束时间 |

**响应 200**
```json
{
  "code": 0,
  "data": {
    "total": 150,
    "page": 1,
    "size": 20,
    "items": [
      {
        "id": "01934b5c-7e8a-7c10-9a5f-8b2d3e4f5a6b",
        "type": 2,
        "type_name": "consume",
        "amount": -7500,
        "balance_after": 87492500,
        "reference_type": "request",
        "reference_id": "request_id_here",
        "description": "gpt-4o: 500 prompt + 200 completion tokens",
        "created_at": "2024-01-15T10:30:00Z"
      }
    ]
  }
}
```

---

## 管理员查看交易记录

### GET /api/v1/admin/transactions

**查询参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码 |
| size | int | 否 | 每页数量 |
| user_id | uuid | 否 | 用户ID筛选 |
| type | int | 否 | 类型筛选 |
| start | string | 否 | 开始时间 |
| end | string | 否 | 结束时间 |

**响应 200**
同上，但包含所有用户的交易记录。
