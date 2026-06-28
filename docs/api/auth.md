# 认证模块

## 注册

### POST /api/v1/auth/register

**请求体**
```json
{
  "email": "user@example.com",
  "username": "john",
  "password": "securePassword123"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 邮箱地址 |
| username | string | 是 | 用户名（3-64字符） |
| password | string | 是 | 密码（8-64字符） |

**响应 201**
```json
{
  "code": 0,
  "data": {
    "id": "01934b5c-7e8a-7c10-9a5f-8b2d3e4f5a6b",
    "email": "user@example.com",
    "username": "john",
    "role": 1,
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**错误码**
| 错误码 | 说明 |
|--------|------|
| 200101 | 邮箱格式无效 |
| 200102 | 用户名格式无效 |
| 200103 | 密码格式无效 |
| 200301 | 邮箱已注册 |
| 200302 | 用户名已存在 |

---

## 登录

### POST /api/v1/auth/login

**请求体**
```json
{
  "email": "user@example.com",
  "password": "securePassword123"
}
```

**响应 200**
```json
{
  "code": 0,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 7200
  }
}
```

**错误码**
| 错误码 | 说明 |
|--------|------|
| 200303 | 邮箱或密码错误 |
| 200304 | 账号已被禁用 |

---

## 刷新 Token

### POST /api/v1/auth/refresh

**请求体**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**响应 200**
```json
{
  "code": 0,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 7200
  }
}
```

**错误码**
| 错误码 | 说明 |
|--------|------|
| 200305 | Refresh Token 无效 |

---

## 获取当前用户

### GET /api/v1/auth/me

**请求头**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**响应 200**
```json
{
  "code": 0,
  "data": {
    "id": "01934b5c-7e8a-7c10-9a5f-8b2d3e4f5a6b",
    "email": "user@example.com",
    "username": "john",
    "role": 1,
    "balance": 87500000,
    "used_quota": 12500000,
    "request_count": 150,
    "group_name": "default",
    "created_at": "2024-01-10T08:00:00Z"
  }
}
```

---

## 修改密码

### PUT /api/v1/auth/password

**请求体**
```json
{
  "old_password": "currentPassword",
  "new_password": "newSecurePassword456"
}
```

**响应 200**
```json
{
  "code": 0,
  "data": {
    "message": "password updated successfully"
  }
}
```

**错误码**
| 错误码 | 说明 |
|--------|------|
| 200401 | 原密码错误 |
| 200402 | 新密码与原密码相同 |
