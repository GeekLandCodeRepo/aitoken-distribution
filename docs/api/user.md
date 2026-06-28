# 用户模块

## 用户列表（管理员）

### GET /api/v1/admin/users

**查询参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| size | int | 否 | 每页数量，默认20 |
| search | string | 否 | 搜索关键词（邮箱/用户名） |
| status | int | 否 | 状态筛选（0/1） |
| role | int | 否 | 角色筛选（1/10） |

**响应 200**
```json
{
  "code": 0,
  "data": {
    "total": 100,
    "page": 1,
    "size": 20,
    "items": [
      {
        "id": "01934b5c-7e8a-7c10-9a5f-8b2d3e4f5a6b",
        "username": "john",
        "email": "john@example.com",
        "role": 1,
        "balance": 50000000,
        "used_quota": 12500000,
        "request_count": 150,
        "status": 1,
        "group_name": "default",
        "created_at": "2024-01-10T08:00:00Z"
      }
    ]
  }
}
```

---

## 更新用户（管理员）

### PUT /api/v1/admin/users/{id}

**路径参数**
| 参数 | 类型 | 说明 |
|------|------|------|
| id | uuid | 用户ID |

**请求体**
```json
{
  "role": 10,
  "status": 1,
  "group_name": "vip"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| role | int | 否 | 角色（1=普通用户, 10=管理员） |
| status | int | 否 | 状态（0=禁用, 1=正常） |
| group_name | string | 否 | 用户组 |

**响应 200**
```json
{
  "code": 0,
  "data": {
    "id": "01934b5c-7e8a-7c10-9a5f-8b2d3e4f5a6b",
    "username": "john",
    "email": "john@example.com",
    "role": 10,
    "status": 1,
    "group_name": "vip"
  }
}
```

**错误码**
| 错误码 | 说明 |
|--------|------|
| 300301 | 用户不存在 |

---

## 手动充值（管理员）

### POST /api/v1/admin/users/{id}/topup

**路径参数**
| 参数 | 类型 | 说明 |
|------|------|------|
| id | uuid | 用户ID |

**请求体**
```json
{
  "amount": 100000000,
  "description": "Manual topup by admin"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| amount | int64 | 是 | 充值金额（内部单位，1美元=1,000,000） |
| description | string | 否 | 备注 |

**响应 200**
```json
{
  "code": 0,
  "data": {
    "user_id": "01934b5c-7e8a-7c10-9a5f-8b2d3e4f5a6b",
    "balance_before": 50000000,
    "balance_after": 150000000,
    "amount": 100000000
  }
}
```

**错误码**
| 错误码 | 说明 |
|--------|------|
| 300301 | 用户不存在 |
| 300401 | 充值金额无效（<=0） |

---

## 删除用户（管理员）

### DELETE /api/v1/admin/users/{id}

**路径参数**
| 参数 | 类型 | 说明 |
|------|------|------|
| id | uuid | 用户ID |

**响应 200**
```json
{
  "code": 0,
  "data": {
    "message": "user deleted successfully"
  }
}
```

**错误码**
| 错误码 | 说明 |
|--------|------|
| 300301 | 用户不存在 |
