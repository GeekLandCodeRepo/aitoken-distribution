# 用量模块

## 用量概览

### GET /api/v1/usage/overview

**响应 200**
```json
{
  "code": 0,
  "data": {
    "balance": 87500000,
    "used_quota": 12500000,
    "request_count": 1500,
    "today": {
      "requests": 25,
      "tokens": 50000,
      "cost": 1250000
    },
    "this_month": {
      "requests": 150,
      "tokens": 300000,
      "cost": 7500000
    }
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| balance | int64 | 当前余额（内部单位） |
| used_quota | int64 | 已用额度 |
| request_count | int64 | 总请求数 |
| today.requests | int64 | 今日请求数 |
| today.tokens | int64 | 今日token数 |
| today.cost | int64 | 今日花费 |
| this_month.requests | int64 | 本月请求数 |
| this_month.tokens | int64 | 本月token数 |
| this_month.cost | int64 | 本月花费 |

---

## 用量统计

### GET /api/v1/usage/stats

**查询参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_by | string | 否 | 分组方式（day/week/month），默认day |
| start | string | 否 | 开始时间 |
| end | string | 否 | 结束时间 |
| model | string | 否 | 模型筛选 |

**响应 200**
```json
{
  "code": 0,
  "data": {
    "stats": [
      {
        "date": "2024-01-15",
        "requests": 25,
        "tokens": 50000,
        "prompt_tokens": 35000,
        "completion_tokens": 15000,
        "cost": 1250000
      },
      {
        "date": "2024-01-14",
        "requests": 30,
        "tokens": 60000,
        "prompt_tokens": 42000,
        "completion_tokens": 18000,
        "cost": 1500000
      }
    ],
    "by_model": [
      {
        "model": "gpt-4o",
        "requests": 40,
        "tokens": 80000,
        "cost": 2000000
      },
      {
        "model": "gpt-3.5-turbo",
        "requests": 15,
        "tokens": 30000,
        "cost": 300000
      }
    ]
  }
}
```

---

## 请求日志

### GET /api/v1/usage/logs

**查询参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| size | int | 否 | 每页数量，默认20 |
| model | string | 否 | 模型筛选 |
| status | int | 否 | 状态码筛选 |
| start | string | 否 | 开始时间 |
| end | string | 否 | 结束时间 |

**响应 200**
```json
{
  "code": 0,
  "data": {
    "total": 1500,
    "page": 1,
    "size": 20,
    "items": [
      {
        "id": "01934b5c-7e8a-7c10-9a5f-8b2d3e4f5a6b",
        "model": "gpt-4o",
        "prompt_tokens": 500,
        "completion_tokens": 200,
        "total_tokens": 700,
        "cost": 7000,
        "cache_hit": false,
        "status_code": 200,
        "is_stream": true,
        "latency_ms": 1200,
        "error_message": null,
        "request_id": "req-abc123",
        "ip_address": "192.168.1.1",
        "created_at": "2024-01-15T10:30:00Z"
      }
    ]
  }
}
```

---

## 管理员全局概览

### GET /api/v1/admin/usage/overview

**响应 200**
```json
{
  "code": 0,
  "data": {
    "total_users": 128,
    "total_requests": 45678,
    "total_revenue": 156800000,
    "active_channels": 12,
    "total_channels": 15,
    "today": {
      "requests": 1250,
      "tokens": 2500000,
      "cost": 62500000
    },
    "this_month": {
      "requests": 45678,
      "tokens": 9135600,
      "cost": 228390000
    }
  }
}
```

---

## 管理员请求日志

### GET /api/v1/admin/usage/logs

**查询参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码 |
| size | int | 否 | 每页数量 |
| user_id | uuid | 否 | 用户ID筛选 |
| model | string | 否 | 模型筛选 |
| channel_id | uuid | 否 | 渠道ID筛选 |
| status | int | 否 | 状态码筛选 |
| start | string | 否 | 开始时间 |
| end | string | 否 | 结束时间 |

**响应 200**
同用户端日志格式，但包含所有用户的日志。

---

## 管理员用户用量排名

### GET /api/v1/admin/usage/users

**查询参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| limit | int | 否 | 返回数量，默认10 |
| start | string | 否 | 开始时间 |
| end | string | 否 | 结束时间 |

**响应 200**
```json
{
  "code": 0,
  "data": [
    {
      "user_id": "01934b5c-7e8a-7c10-9a5f-8b2d3e4f5a6b",
      "username": "john",
      "email": "john@example.com",
      "requests": 12500,
      "tokens": 2500000,
      "cost": 45000000
    }
  ]
}
```

---

## 管理员渠道用量排名

### GET /api/v1/admin/usage/channels

**查询参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| limit | int | 否 | 返回数量，默认10 |
| start | string | 否 | 开始时间 |
| end | string | 否 | 结束时间 |

**响应 200**
```json
{
  "code": 0,
  "data": [
    {
      "channel_id": "01934b5c-7e8a-7c10-9a5f-8b2d3e4f5a6b",
      "name": "OpenAI Official",
      "requests": 5000,
      "tokens": 1000000,
      "cost": 25000000,
      "success_rate": 0.995
    }
  ]
}
```

---

## 管理员全局统计

### GET /api/v1/admin/usage/stats

**查询参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_by | string | 否 | 分组方式（day/week/month） |
| start | string | 否 | 开始时间 |
| end | string | 否 | 结束时间 |

**响应 200**
```json
{
  "code": 0,
  "data": {
    "stats": [
      {
        "date": "2024-01-15",
        "requests": 1250,
        "tokens": 2500000,
        "cost": 62500000,
        "new_users": 5
      }
    ],
    "by_model": [
      {
        "model": "gpt-4o",
        "requests": 800,
        "tokens": 1600000,
        "cost": 40000000
      }
    ],
    "by_channel": [
      {
        "channel_id": "...",
        "name": "OpenAI Official",
        "requests": 1000,
        "cost": 50000000
      }
    ]
  }
}
```
