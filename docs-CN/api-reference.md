# API 参考

本文档描述了 Mino 后端提供的所有 REST 和 WebSocket
端点。所有 REST 端点均返回使用标准信封格式的 JSON 响应。

## 响应格式

所有 API 响应遵循以下结构。

```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

分页响应包含额外的元数据。

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [],
    "total": 0,
    "page": 1,
    "page_size": 20
  }
}
```

错误响应使用相同的信封格式，并附带相应的 HTTP 状态码和
错误消息。

## 认证

所有受保护的端点需要在 `Authorization` 请求头中提供有效的
JWT 访问令牌。

```
Authorization: Bearer <access_token>
```

WebSocket 端点改为通过查询参数接受令牌。

```
ws://host/v1/ws/audio?token=<access_token>
```

### POST /v1/auth/signin

使用用户名和密码进行认证。返回访问令牌和刷新令牌对。

**请求体：**

```json
{
  "username": "mino",
  "password": "admin"
}
```

**响应 (200)：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJSUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJSUzI1NiIs...",
    "user": {
      "id": "uuid",
      "username": "mino",
      "role": "admin"
    }
  }
}
```

### POST /v1/auth/signout

登出当前用户。这是一个客户端操作；服务器确认该请求。

**响应 (200)：**

```json
{
  "code": 200,
  "message": "signed out"
}
```

### POST /v1/auth/refresh

使用有效的刷新令牌来刷新过期的访问令牌。

**请求体：**

```json
{
  "refresh_token": "eyJhbGciOiJSUzI1NiIs..."
}
```

**响应 (200)：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJSUzI1NiIs..."
  }
}
```

### POST /v1/auth/password

修改当前用户的密码。需要认证。

**请求体：**

```json
{
  "old_password": "current_password",
  "new_password": "new_password"
}
```

**响应 (200)：**

```json
{
  "code": 200,
  "message": "password changed"
}
```

## 对话

对话表示已录制的音频会话及其转录文本和 AI 生成的元数据。

### GET /v1/conversations

列出已认证用户的对话。支持分页。

**查询参数：**

| 参数 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `page` | integer | 1 | 页码 |
| `page_size` | integer | 20 | 每页条目数 |

**响应 (200)：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": "uuid",
        "user_id": "uuid",
        "title": "Meeting notes",
        "summary": "Discussed project timeline...",
        "transcript": "Full transcript text...",
        "audio_url": "https://...",
        "audio_duration": 120,
        "language": "zh",
        "status": "completed",
        "recorded_at": "2026-02-28T10:00:00Z",
        "created_at": "2026-02-28T10:00:00Z",
        "updated_at": "2026-02-28T10:02:00Z"
      }
    ],
    "total": 50,
    "page": 1,
    "page_size": 20
  }
}
```

### GET /v1/conversations/:id

根据 ID 获取单个对话。

**响应 (200)：** 返回对话对象。

**响应 (404)：** 对话未找到。

### DELETE /v1/conversations/:id

根据 ID 删除对话。

**响应 (204)：** 无内容。

**响应 (404)：** 对话未找到。

## 记忆

记忆是从对话中提取的洞察、事实、偏好和事件。

### GET /v1/memories

列出已认证用户的记忆。支持分页和筛选。

**查询参数：**

| 参数 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `page` | integer | 1 | 页码 |
| `page_size` | integer | 20 | 每页条目数 |
| `category` | string | | 按类别筛选 |

**响应 (200)：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": "uuid",
        "user_id": "uuid",
        "conversation_id": "uuid",
        "content": "Prefers morning meetings",
        "category": "preference",
        "importance": 7,
        "embedding_id": null,
        "created_at": "2026-02-28T10:02:00Z",
        "updated_at": "2026-02-28T10:02:00Z"
      }
    ],
    "total": 25,
    "page": 1,
    "page_size": 20
  }
}
```

### GET /v1/memories/:id

根据 ID 获取单条记忆。

### PUT /v1/memories/:id

更新记忆的内容、类别或重要性。

**请求体：**

```json
{
  "content": "Updated memory content",
  "category": "fact",
  "importance": 8
}
```

**响应 (200)：** 返回更新后的记忆对象。

### DELETE /v1/memories/:id

根据 ID 删除记忆。

**响应 (204)：** 无内容。

## 任务

任务是从对话中提取的待办事项，也可以手动创建。

### GET /v1/tasks

列出已认证用户的任务。支持分页和筛选。

**查询参数：**

| 参数 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `page` | integer | 1 | 页码 |
| `page_size` | integer | 20 | 每页条目数 |
| `status` | string | | 按状态筛选 |

**响应 (200)：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": "uuid",
        "user_id": "uuid",
        "conversation_id": "uuid",
        "title": "Review project proposal",
        "description": "Review and provide feedback",
        "status": "pending",
        "priority": "high",
        "due_date": "2026-03-01T00:00:00Z",
        "completed_at": null,
        "created_at": "2026-02-28T10:02:00Z",
        "updated_at": "2026-02-28T10:02:00Z"
      }
    ],
    "total": 10,
    "page": 1,
    "page_size": 20
  }
}
```

### POST /v1/tasks

创建新任务。

**请求体：**

```json
{
  "title": "Follow up with team",
  "description": "Send meeting summary",
  "priority": "medium",
  "due_date": "2026-03-05T00:00:00Z"
}
```

**响应 (201)：** 返回创建的任务对象。

### PUT /v1/tasks/:id

更新任务的标题、描述、状态、优先级或截止日期。

**请求体：**

```json
{
  "status": "completed",
  "completed_at": "2026-02-28T15:00:00Z"
}
```

**响应 (200)：** 返回更新后的任务对象。

### DELETE /v1/tasks/:id

根据 ID 删除任务。

**响应 (204)：** 无内容。

## 聊天

聊天提供按会话组织的 AI 对话交互。

### POST /v1/chat/sessions

创建新的聊天会话。

**响应 (201)：**

```json
{
  "code": 201,
  "message": "created",
  "data": {
    "id": "uuid",
    "user_id": "uuid",
    "title": "New conversation",
    "created_at": "2026-02-28T10:00:00Z",
    "updated_at": "2026-02-28T10:00:00Z"
  }
}
```

### GET /v1/chat/sessions

列出已认证用户的所有聊天会话，按最近更新时间排序。

**响应 (200)：** 返回会话对象数组。

### PUT /v1/chat/sessions/:id

重命名聊天会话。

**请求体：**

```json
{
  "title": "Project discussion"
}
```

**响应 (200)：** 返回更新后的会话对象。

### DELETE /v1/chat/sessions/:id

删除聊天会话及其所有消息。

**响应 (204)：** 无内容。

### GET /v1/chat/sessions/:id/messages

获取聊天会话中的所有消息。

**响应 (200)：**

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": "uuid",
      "session_id": "uuid",
      "user_id": "uuid",
      "role": "user",
      "content": "What did I discuss yesterday?",
      "sources": null,
      "created_at": "2026-02-28T10:00:00Z"
    },
    {
      "id": "uuid",
      "session_id": "uuid",
      "user_id": "uuid",
      "role": "assistant",
      "content": "Based on your recordings...",
      "sources": [
        {
          "conversation_id": "uuid",
          "title": "Morning meeting",
          "excerpt": "Relevant excerpt..."
        }
      ],
      "created_at": "2026-02-28T10:00:01Z"
    }
  ]
}
```

### POST /v1/chat/sessions/:id/messages

向 AI 助手发送消息。助手使用语义向量搜索（Milvus）从您的
对话历史中检索相关上下文，并生成带有来源引用的回复。如果
向量搜索不可用，服务将回退到 PostgreSQL 中的关键词匹配。

**请求体：**

```json
{
  "content": "What tasks did I mention last week?"
}
```

**响应 (200)：** 返回助手的回复消息及来源信息。

## 扩展

扩展是用户定义的附加组件，具有可配置的设置。

### GET /v1/extensions

列出已认证用户的所有扩展。

**响应 (200)：**

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "name": "Calendar sync",
      "description": "Sync tasks to calendar",
      "icon": "calendar",
      "enabled": true,
      "config": {
        "calendar_url": "https://..."
      },
      "created_at": "2026-02-28T10:00:00Z",
      "updated_at": "2026-02-28T10:00:00Z"
    }
  ]
}
```

### GET /v1/extensions/:id

根据 ID 获取单个扩展。

### POST /v1/extensions

创建新扩展。

**请求体：**

```json
{
  "name": "Calendar sync",
  "description": "Sync tasks to calendar",
  "icon": "calendar",
  "enabled": true,
  "config": {}
}
```

**响应 (201)：** 返回创建的扩展对象。

### PUT /v1/extensions/:id

更新扩展的名称、描述、图标、启用状态或配置。

**响应 (200)：** 返回更新后的扩展对象。

### DELETE /v1/extensions/:id

根据 ID 删除扩展。

**响应 (204)：** 无内容。

## 搜索

搜索功能使用 Typesense 提供跨对话和记忆的全文搜索。语义
向量搜索（Milvus）由聊天服务内部用于 RAG 上下文检索，
不作为单独的 API 端点暴露。

### GET /v1/search

跨对话和记忆进行搜索。

**查询参数：**

| 参数 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `q` | string | 必填 | 搜索查询 |
| `limit` | integer | 20 | 最大结果数 |

**响应 (200)：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "results": [
      {
        "collection": "conversations",
        "id": "uuid",
        "title": "Meeting notes",
        "highlight": "...discussed <mark>project</mark>..."
      }
    ]
  }
}
```

### POST /v1/search/reindex

从 PostgreSQL 数据重建 Typesense 搜索索引。此操作会同步
已认证用户的所有对话和记忆。

**响应 (200)：**

```json
{
  "code": 200,
  "message": "reindex completed"
}
```

## WebSocket

WebSocket 端点支持实时音频流传输和转录。

### GET /v1/ws/audio

建立用于音频流传输的 WebSocket 连接。认证通过 `token`
查询参数提供。

```
ws://host/v1/ws/audio?token=<jwt_access_token>
```

### 客户端到服务器的消息

**音频数据：**

```json
{
  "type": "audio",
  "data": "<base64_encoded_audio_chunk>",
  "timestamp": 1234567890,
  "sequence": 1
}
```

**控制命令：**

```json
{
  "type": "control",
  "action": "start"
}
```

有效的 action 值为 `start`、`stop`、`pause` 和 `resume`。

### 服务器到客户端的消息

**实时转录：**

```json
{
  "type": "transcript",
  "text": "Transcribed text so far...",
  "is_final": false,
  "timestamp": 1234567890
}
```

**状态更新：**

```json
{
  "type": "status",
  "status": "recording",
  "message": "Recording started"
}
```

**处理完成：**

```json
{
  "type": "completed",
  "conversation_id": "uuid",
  "title": "Generated title",
  "summary": "Generated summary",
  "action_items": [
    "Review the proposal",
    "Schedule follow-up meeting"
  ],
  "memories": [
    "Prefers morning meetings",
    "Project deadline is March 15"
  ]
}
```

**错误：**

```json
{
  "type": "error",
  "message": "Error description"
}
```

## 速率限制

所有受保护的端点限制为每用户每端点每分钟 100 次请求。
超过限制时，服务器返回 `429 Too Many Requests` 响应。

速率限制器使用基于 Redis 的滑动窗口计数器。如果 Redis
不可用，速率限制器会优雅降级并放行请求。
