# 代理模块

## 聊天补全

### POST /v1/chat/completions

**描述**: OpenAI兼容的聊天补全接口。

**请求头**
```
Authorization: Bearer sk-xxxxxxxxxxxx
Content-Type: application/json
```

**请求体**（标准OpenAI格式）
```json
{
  "model": "gpt-4o",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Hello!"}
  ],
  "stream": false,
  "temperature": 0.7,
  "max_tokens": 1000
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| model | string | 是 | 模型名称 |
| messages | []object | 是 | 消息数组 |
| stream | bool | 否 | 是否流式，默认false |
| temperature | float | 否 | 温度参数 |
| max_tokens | int | 否 | 最大token数 |

**响应 200**（非流式）
```json
{
  "id": "chatcmpl-abc123",
  "object": "chat.completion",
  "created": 1705312200,
  "model": "gpt-4o",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! How can I help you today?"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 25,
    "completion_tokens": 10,
    "total_tokens": 35
  }
}
```

**响应 200**（流式 SSE）
```
data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1705312200,"model":"gpt-4o","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1705312200,"model":"gpt-4o","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1705312200,"model":"gpt-4o","choices":[{"index":0,"delta":{"content":"!"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1705312200,"model":"gpt-4o","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]
```

**错误码**
| 错误码 | 说明 |
|--------|------|
| 800102 | model 参数缺失 |
| 800103 | messages 格式无效 |
| 800301 | 模型不存在或不可用 |
| 800302 | 无可用渠道 |
| 800303 | 所有渠道均失败 |
| 800304 | 上游请求超时 |
| 400302 | API Key 已被禁用 |
| 400303 | API Key 已过期 |
| 400304 | API Key 额度已用完 |
| 400401 | Key 不允许访问该模型 |
| 400403 | Key 请求频率超限 |
| 700401 | 用户余额不足 |

---

## 文本补全

### POST /v1/completions

**请求体**
```json
{
  "model": "gpt-3.5-turbo-instruct",
  "prompt": "Say hello:",
  "max_tokens": 50
}
```

**响应 200**
```json
{
  "id": "cmpl-abc123",
  "object": "text_completion",
  "created": 1705312200,
  "model": "gpt-3.5-turbo-instruct",
  "choices": [
    {
      "text": "Hello! How are you today?",
      "index": 0,
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 5,
    "completion_tokens": 8,
    "total_tokens": 13
  }
}
```

---

## 向量嵌入

### POST /v1/embeddings

**请求体**
```json
{
  "model": "text-embedding-3-small",
  "input": "The food was delicious and the waiter..."
}
```

**响应 200**
```json
{
  "object": "list",
  "data": [
    {
      "object": "embedding",
      "index": 0,
      "embedding": [0.0023064255, -0.009327292, ...]
    }
  ],
  "model": "text-embedding-3-small",
  "usage": {
    "prompt_tokens": 8,
    "total_tokens": 8
  }
}
```

---

## 图片生成

### POST /v1/images/generations

**请求体**
```json
{
  "model": "dall-e-3",
  "prompt": "A cute baby sea otter",
  "n": 1,
  "size": "1024x1024"
}
```

**响应 200**
```json
{
  "created": 1705312200,
  "data": [
    {
      "url": "https://..."
    }
  ]
}
```

---

## 语音转文字

### POST /v1/audio/transcriptions

**请求** (multipart/form-data)
```
file: (audio file)
model: whisper-1
```

**响应 200**
```json
{
  "text": "Hello, how are you today?"
}
```

---

## 文字转语音

### POST /v1/audio/speech

**请求体**
```json
{
  "model": "tts-1",
  "input": "Hello, how are you today?",
  "voice": "alloy"
}
```

**响应 200**
```
Content-Type: audio/mpeg
(binary audio data)
```

---

## 模型列表

### GET /v1/models

**请求头**
```
Authorization: Bearer sk-xxxxxxxxxxxx
```

**响应 200**
```json
{
  "object": "list",
  "data": [
    {
      "id": "gpt-4o",
      "object": "model",
      "created": 1705312200,
      "owned_by": "openai"
    },
    {
      "id": "gpt-3.5-turbo",
      "object": "model",
      "created": 1705312200,
      "owned_by": "openai"
    }
  ]
}
```

**说明**: 返回当前API Key允许访问的所有模型列表。
