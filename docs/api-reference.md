# API reference

This document describes all REST and WebSocket endpoints provided by
the Mino backend. All REST endpoints return JSON responses using a
standard envelope format.

## Response format

All API responses follow this structure.

```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

Paginated responses include additional metadata.

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

Error responses use the same envelope with an appropriate HTTP status
code and error message.

## Authentication

All protected endpoints require a valid JWT access token in the
`Authorization` header.

```
Authorization: Bearer <access_token>
```

The WebSocket endpoint accepts the token as a query parameter instead.

```
ws://host/v1/ws/audio?token=<access_token>
```

### POST /v1/auth/signin

Authenticate with username and password. Returns an access token and
refresh token pair.

**Request body:**

```json
{
  "username": "mino",
  "password": "admin"
}
```

**Response (200):**

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

Sign out the current user. This is a client-side operation; the server
acknowledges the request.

**Response (200):**

```json
{
  "code": 200,
  "message": "signed out"
}
```

### POST /v1/auth/refresh

Refresh an expired access token using a valid refresh token.

**Request body:**

```json
{
  "refresh_token": "eyJhbGciOiJSUzI1NiIs..."
}
```

**Response (200):**

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

Change the current user's password. Requires authentication.

**Request body:**

```json
{
  "old_password": "current_password",
  "new_password": "new_password"
}
```

**Response (200):**

```json
{
  "code": 200,
  "message": "password changed"
}
```

## Conversations

Conversations represent recorded audio sessions with their transcripts
and AI-generated metadata.

### GET /v1/conversations

List conversations for the authenticated user. Supports pagination.

**Query parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | 1 | Page number |
| `page_size` | integer | 20 | Items per page |

**Response (200):**

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

Get a single conversation by ID.

**Response (200):** Returns the conversation object.

**Response (404):** Conversation not found.

### DELETE /v1/conversations/:id

Delete a conversation by ID.

**Response (204):** No content.

**Response (404):** Conversation not found.

## Memories

Memories are insights, facts, preferences, and events extracted from
conversations.

### GET /v1/memories

List memories for the authenticated user. Supports pagination and
filtering.

**Query parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | 1 | Page number |
| `page_size` | integer | 20 | Items per page |
| `category` | string | | Filter by category |

**Response (200):**

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

Get a single memory by ID.

### PUT /v1/memories/:id

Update a memory's content, category, or importance.

**Request body:**

```json
{
  "content": "Updated memory content",
  "category": "fact",
  "importance": 8
}
```

**Response (200):** Returns the updated memory object.

### DELETE /v1/memories/:id

Delete a memory by ID.

**Response (204):** No content.

## Tasks

Tasks are action items extracted from conversations or created
manually.

### GET /v1/tasks

List tasks for the authenticated user. Supports pagination and
filtering.

**Query parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | 1 | Page number |
| `page_size` | integer | 20 | Items per page |
| `status` | string | | Filter by status |

**Response (200):**

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

Create a new task.

**Request body:**

```json
{
  "title": "Follow up with team",
  "description": "Send meeting summary",
  "priority": "medium",
  "due_date": "2026-03-05T00:00:00Z"
}
```

**Response (201):** Returns the created task object.

### PUT /v1/tasks/:id

Update a task's title, description, status, priority, or due date.

**Request body:**

```json
{
  "status": "completed",
  "completed_at": "2026-02-28T15:00:00Z"
}
```

**Response (200):** Returns the updated task object.

### DELETE /v1/tasks/:id

Delete a task by ID.

**Response (204):** No content.

## Chat

Chat provides conversational AI interactions organized into sessions.

### POST /v1/chat/sessions

Create a new chat session.

**Response (201):**

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

List all chat sessions for the authenticated user, ordered by most
recently updated.

**Response (200):** Returns an array of session objects.

### PUT /v1/chat/sessions/:id

Rename a chat session.

**Request body:**

```json
{
  "title": "Project discussion"
}
```

**Response (200):** Returns the updated session object.

### DELETE /v1/chat/sessions/:id

Delete a chat session and all its messages.

**Response (204):** No content.

### GET /v1/chat/sessions/:id/messages

Get all messages in a chat session.

**Response (200):**

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

Send a message to the AI assistant. The assistant retrieves relevant
context from your conversation history using semantic vector search
(Milvus) and generates a response with source citations. If vector
search is unavailable, the service falls back to keyword matching in
PostgreSQL.

**Request body:**

```json
{
  "content": "What tasks did I mention last week?"
}
```

**Response (200):** Returns the assistant's response message with
sources.

## Extensions

Extensions are user-defined add-ons with configurable settings.

### GET /v1/extensions

List all extensions for the authenticated user.

**Response (200):**

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

Get a single extension by ID.

### POST /v1/extensions

Create a new extension.

**Request body:**

```json
{
  "name": "Calendar sync",
  "description": "Sync tasks to calendar",
  "icon": "calendar",
  "enabled": true,
  "config": {}
}
```

**Response (201):** Returns the created extension object.

### PUT /v1/extensions/:id

Update an extension's name, description, icon, enabled state, or
configuration.

**Response (200):** Returns the updated extension object.

### DELETE /v1/extensions/:id

Delete an extension by ID.

**Response (204):** No content.

## Search

Search provides full-text search across conversations and memories
using Typesense. Semantic vector search (Milvus) is used internally
by the chat service for RAG context retrieval and is not exposed as a
separate API endpoint.

### GET /v1/search

Search across conversations and memories.

**Query parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `q` | string | required | Search query |
| `limit` | integer | 20 | Maximum results |

**Response (200):**

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

Rebuild the Typesense search index from PostgreSQL data. This
synchronizes all conversations and memories for the authenticated user.

**Response (200):**

```json
{
  "code": 200,
  "message": "reindex completed"
}
```

## WebSocket

The WebSocket endpoint enables real-time audio streaming and
transcription.

### GET /v1/ws/audio

Establish a WebSocket connection for audio streaming. Authentication
is provided via the `token` query parameter.

```
ws://host/v1/ws/audio?token=<jwt_access_token>
```

### Client-to-server messages

**Audio data:**

```json
{
  "type": "audio",
  "data": "<base64_encoded_audio_chunk>",
  "timestamp": 1234567890,
  "sequence": 1
}
```

**Control commands:**

```json
{
  "type": "control",
  "action": "start"
}
```

Valid actions are `start`, `stop`, `pause`, and `resume`.

### Server-to-client messages

**Real-time transcript:**

```json
{
  "type": "transcript",
  "text": "Transcribed text so far...",
  "is_final": false,
  "timestamp": 1234567890
}
```

**Status update:**

```json
{
  "type": "status",
  "status": "recording",
  "message": "Recording started"
}
```

**Processing complete:**

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

**Error:**

```json
{
  "type": "error",
  "message": "Error description"
}
```

## Rate limiting

All protected endpoints are rate-limited to 100 requests per minute per
user per endpoint. When the limit is exceeded, the server returns a
`429 Too Many Requests` response.

The rate limiter uses Redis-based sliding window counters. If Redis is
unavailable, the rate limiter degrades gracefully and allows requests
through.
