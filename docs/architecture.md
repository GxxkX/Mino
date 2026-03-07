# Architecture

This document explains how Mino is designed and how the components work
together. Understanding the architecture helps you with troubleshooting,
customization, and extending the system.

## System overview

Mino follows a layered architecture with clear separation between
clients, the API gateway, services, and data storage. Each layer
communicates through well-defined interfaces, making it possible to
replace or modify components without affecting the rest of the system.

The system consists of four main layers.

- The **client layer** includes all user-facing applications:
  smartwatch, phone, web, and desktop clients.
- The **gateway layer** handles authentication, rate limiting, CORS,
  structured logging, and request routing.
- The **service layer** contains the business logic for audio
  processing, transcription, AI analysis, chat, search, and extensions.
- The **data layer** stores information in PostgreSQL, Milvus, MinIO,
  Typesense, and Redis.

## Client applications

Mino supports four client platforms, each optimized for its device.

The **watch app** built with Flutter is designed for quick voice capture
on wearable devices. It streams audio to the backend via WebSocket and
displays real-time transcription as you speak. When offline, it stores
recordings locally in SQLite and syncs when connectivity returns. The
app uses the Provider pattern for state management and supports
configurable audio quality settings.

The **phone app** also built with Flutter provides a complete interface
for recording, reviewing history, managing memories and tasks, and
chatting with the AI assistant. It supports both recording and playback
of audio files.

The **web app** uses Next.js 14 with React 18 to deliver a responsive
browser-based interface. It includes all features available on mobile
plus additional capabilities like extensions management, global search
(with Cmd+K shortcut), and MCP configuration. The app uses Zustand for
state management and communicates with the backend through a client-side
API layer with automatic JWT token refresh.

The **desktop app** built with Go and Wails offers native performance on
Windows, macOS, and Linux. It provides the same feature set as the web
app in a standalone application.

## Backend services

The backend is written in Go 1.24 using the Gin web framework. It
provides both REST APIs for standard operations and WebSocket endpoints
for real-time streaming. The application compiles to a single binary
that connects to external infrastructure services.

### API gateway

All client requests go through the API gateway first. The gateway
handles the following responsibilities.

- **Authentication**: validates JWT tokens using RSA RS256 signatures.
  Access tokens expire after 15 minutes, refresh tokens after 7 days.
- **Rate limiting**: uses Redis-based sliding window counters, keyed by
  user ID and endpoint path. Default limit is 100 requests per minute.
  Gracefully degrades if Redis is unavailable.
- **CORS**: allows all origins for development flexibility.
- **Logging**: structured JSON logging via logrus, capturing status
  codes, methods, paths, query parameters, client IPs, and latency.

### Audio service

The audio service manages the recording-to-insight pipeline. When a
WebSocket recording session ends, the audio service receives the
complete transcript and orchestrates the processing workflow.

The service creates a conversation record in PostgreSQL, then calls
the ExtractAgent to extract structured information (title, summary,
action items, and memory points). Extracted memories and tasks are
batch-inserted into their respective tables using database
transactions. After extraction completes, the service asynchronously
generates vector embeddings for the conversation transcript and
extracted memories, then stores them in Milvus for semantic search.

The service accepts audio in PCM16 format at 16kHz sampling rate. Opus
compression is supported to reduce bandwidth on mobile connections.

### Transcription service

The transcription service converts speech to text. It integrates with
speech-to-text providers such as Zhipu ASR or self-hosted Whisper. The
service receives streaming audio and returns transcription results in
real-time, enabling the client to display text as the user speaks.

You can configure which provider to use through environment variables.
The service is designed to support multiple providers, making it easy to
switch or add new transcription engines.

### AI service (dual-agent architecture)

The AI service implements a dual-agent architecture using LangchainGo.
Two specialized agents handle different tasks.

**ChatAgent** powers the conversational AI assistant. It receives user
messages along with retrieved context from conversation history and
generates natural language responses. The system prompt instructs the
agent to answer based on historical conversations, avoid fabricating
content, naturally cite sources, and support multiple languages.

**ExtractAgent** processes completed transcripts and extracts structured
data. It outputs strict JSON containing a title, summary, action items,
and memory points. This agent runs automatically after each recording
session completes.

**SummarizeTitle** is a utility method that generates short titles for
chat sessions based on the first user message and assistant reply.

The `LLMProvider` interface defines the contract for these agents.

```go
type LLMProvider interface {
    ChatAgent(ctx context.Context, userMessage string,
        retrievedContext string) (string, error)
    ExtractAgent(ctx context.Context, transcript string) (
        *StructuredResult, error)
    SummarizeTitle(ctx context.Context, userMessage string,
        assistantReply string) (string, error)
}
```

The `EmbeddingProvider` interface handles vector embedding generation
for the semantic search pipeline.

```go
type EmbeddingProvider interface {
    EmbedQuery(ctx context.Context, text string) ([]float32, error)
    EmbedDocuments(ctx context.Context, texts []string) (
        [][]float32, error)
}
```

The `LangchainLLMService` implements both interfaces. It uses a
dedicated embedding model (configured via `LLM_EMBEDDING_MODEL`) that
is separate from the chat model. For example, when using Zhipu as the
provider, the chat model might be `glm-4.7-flash` while the embedding
model is `embedding-3`.

The LangchainGo implementation supports three LLM providers: OpenAI,
Zhipu, and Ollama. All use the OpenAI-compatible API format, configured
through `LLM_PROVIDER`, `LLM_API_KEY`, `LLM_BASE_URL`, `LLM_MODEL`,
and `LLM_EMBEDDING_MODEL` environment variables.

### Chat service

The chat service manages conversational interactions with the AI
assistant. It provides session-based chat with full CRUD operations on
sessions and messages.

When a user sends a message, the service performs the following steps.

1. Saves the user message to the database.
2. Searches for relevant context using Milvus semantic vector search.
   If vector search is unavailable or returns no results, falls back
   to keyword matching in PostgreSQL.
3. Calls the ChatAgent with the user message and retrieved context.
4. Saves the assistant response (with source citations) to the
   database.
5. Auto-generates a session title using SummarizeTitle if the session
   is new.

Chat messages include a `sources` field (stored as JSONB in PostgreSQL)
that contains citations linking back to original conversations.

### Search service

The search service provides full-text search across conversations and
memories using Typesense. It manages two Typesense collections.

- **conversations**: indexed fields include title, summary, and
  transcript.
- **memories**: indexed fields include content, category, and
  importance.

The service supports multi-collection search and provides a reindex
endpoint that synchronizes all user data from PostgreSQL to Typesense.
Documents are automatically synced when conversations or memories are
created.

### Vector store service

The vector store service bridges the embedding provider and Milvus to
provide semantic search capabilities. It handles three responsibilities.

**Indexing** generates vector embeddings for conversations and memories
using the `EmbeddingProvider`, then stores them in Milvus. Conversation
text is truncated to 6,000 characters before embedding to stay within
API limits. Memory embeddings are generated in batches for efficiency.

**Searching** embeds the user's query text, then performs approximate
nearest neighbor (ANN) search in Milvus filtered by `user_id`. The
`SearchAll` method queries both the conversations and memories
collections, merges results by cosine similarity score, and returns
the top-K matches.

**Deletion** removes vectors from Milvus when conversations or
memories are deleted, keeping the vector index consistent with
PostgreSQL.

The vector store service is optional. If Milvus or the embedding
provider is unavailable at startup, the service is set to `nil` and
the chat service falls back to keyword search automatically.

### Extension service

The extension service manages user-defined extensions. Each extension
has a name, description, icon, enabled state, and a JSON configuration
object. Extensions are scoped per user and support full CRUD operations.

## Data storage

Mino uses multiple specialized databases to handle different types of
data effectively.

### PostgreSQL

PostgreSQL stores all structured data. The schema includes nine tables.

| Table | Purpose |
|-------|---------|
| `mino_users` | User accounts and authentication |
| `mino_conversations` | Recording metadata and transcripts |
| `mino_memories` | Extracted insights and facts |
| `mino_tasks` | Action items and to-do entries |
| `mino_tags` | Content organization labels |
| `mino_conversation_tags` | Many-to-many tag associations |
| `mino_chat_sessions` | Chat conversation sessions |
| `mino_chat_messages` | Individual chat messages with sources |
| `mino_extensions` | User-defined extensions |

All tables use UUID primary keys. User data is strictly isolated through
`user_id` foreign keys on every query.

### Milvus

Milvus is a vector database that stores embeddings for semantic
similarity search. The backend manages two collections.

- **conversations**: stores vector embeddings of conversation
  transcripts and summaries.
- **memories**: stores vector embeddings of extracted memory content.

Each collection uses the following schema.

| Field | Type | Description |
|-------|------|-------------|
| `id` | Int64 (auto) | Primary key |
| `user_id` | VarChar(64) | Owner user ID for filtering |
| `source_id` | VarChar(64) | Conversation or memory UUID |
| `text` | VarChar(8192) | Original text stored alongside the vector |
| `vector` | FloatVector | Embedding vector |

The default vector dimension is 1024 (matching Zhipu `embedding-3`).
If you use a different embedding model, adjust the dimension
accordingly. Collections use IVF_FLAT indexing with COSINE similarity
metric.

On startup, the backend connects to Milvus, creates any missing
collections, builds indexes, and loads them into memory for search.
If Milvus is unreachable, the backend logs a warning and continues
without vector search.

### MinIO

MinIO provides S3-compatible object storage for audio files. Audio
files are stored with unique identifiers and associated with their
conversation records in PostgreSQL.

> **Note:** MinIO upload is configured but the audio upload pipeline
> currently returns placeholder URLs. Full audio storage integration is
> planned.

### Typesense

Typesense provides full-text search capabilities. Two collections are
maintained: `conversations` (with title, summary, and transcript
fields) and `memories` (with content, category, and importance fields).
The search service ensures these collections stay synchronized with
PostgreSQL data.

### Redis

Redis handles caching, session storage, and rate limiting. The rate
limiter uses a sliding window algorithm keyed by user ID and endpoint
path. Rate limit counters expire automatically. If Redis becomes
unavailable, the rate limiter degrades gracefully and allows requests
through.

## LangSmith observability

Mino integrates with LangSmith for LLM call tracing and observability.
When enabled via the `LANGSMITH_TRACING` environment variable, all LLM
calls are traced with input/output capture and sent asynchronously to
the LangSmith API.

The integration uses a custom callback handler that implements the
LangchainGo callback interface. Traces are organized under the project
name configured in `LANGSMITH_PROJECT`.

## Data flows

Understanding how data moves through the system helps with debugging
and optimization.

### Recording flow

When you record audio from any client, the following steps occur.

1. The client establishes a WebSocket connection to `/v1/ws/audio`,
   passing a JWT token via the `token` query parameter.
2. The gateway validates the token and upgrades the connection.
3. The client sends a control message
   `{"type":"control","action":"start"}` to begin recording.
4. The client streams audio chunks (typically every 100 milliseconds)
   as binary WebSocket messages with base64-encoded audio data.
5. The backend forwards chunks to the transcription service.
6. The transcription service returns partial results in real-time. The
   backend pushes these to the client as
   `{"type":"transcript","text":"...","is_final":false}`.
7. When the client sends `{"type":"control","action":"stop"}`, the
   transcription service returns the complete transcript.
8. The backend sends the transcript to the AudioService for
   asynchronous processing.
9. The ExtractAgent generates the title, summary, action items, and
    memories.
10. Structured data is stored in PostgreSQL. Memories and tasks are
    batch-inserted in transactions.
11. The backend asynchronously generates vector embeddings for the
    conversation transcript and extracted memories, then stores them
    in Milvus for semantic search.
12. The backend notifies the client with a
    `{"type":"completed","conversation_id":"..."}` message.

### Chat flow

When you chat with the AI assistant, the following steps occur.

1. The client sends a message to
   `POST /v1/chat/sessions/:id/messages`.
2. The gateway authenticates the request and passes it to the chat
   service.
3. The chat service saves the user message to the database.
4. The service generates a vector embedding for the user's query and
   performs an ANN search in Milvus across both conversations and
   memories, filtered by user ID. If Milvus is unavailable or returns
   no results, the service falls back to keyword matching in
   PostgreSQL.
5. The ChatAgent receives the user message and retrieved context, then
   generates a response with source citations.
6. The assistant response is saved to the database with its sources.
7. If this is the first exchange in the session, SummarizeTitle
   generates a session title automatically.
8. The response is returned to the client with citations.

### Search flow

When you search for content, the following steps occur.

1. The client sends a query to `GET /v1/search?q=keyword&limit=20`.
2. The search service performs a multi-collection search across
   Typesense, querying both conversations and memories.
3. Results are returned ranked by relevance, with each result
   indicating its source collection.

## Security

Mino implements several security measures to protect your data.

All communication uses HTTPS and WSS for encrypted transmission.
Passwords are hashed using bcrypt with a cost factor of 12 or higher.
JWT tokens are signed with RSA RS256 key pairs and have short
expiration times.

User data is strictly isolated. Every database query filters by
`user_id`, ensuring users can only access their own recordings,
memories, tasks, chat sessions, and extensions.

API endpoints are protected by Redis-based rate limiting. The system
tracks usage per user and per endpoint, blocking excessive requests at
100 requests per minute by default.

## Extension points

Mino is designed to be extensible in several ways.

The **Model Context Protocol (MCP)** support allows you to connect
external tools and services to the AI assistant. You can configure MCP
through the web interface settings.

**Application extensions** enable additional functionality that can be
enabled or disabled per user. Each extension carries a JSON
configuration object for flexible customization.

The **modular service architecture** makes it possible to swap out
components. For example, you can replace the transcription provider
(Zhipu ASR, Whisper, or others), change the LLM model or provider
(OpenAI, Zhipu, Ollama), or add new storage backends without affecting
other parts of the system.

## Technology choices

Each technology in Mino was chosen for specific reasons.

**Go** was selected for the backend because of its excellent
performance, strong concurrency support, and mature ecosystem for
building networked services. The Gin framework provides fast HTTP
routing with middleware support.

**Flutter** enables code sharing between watch and phone apps while
delivering native performance on both platforms. The Provider pattern
simplifies state management for the watch app.

**Next.js 14** provides the App Router for the web interface, with
server-side rendering capabilities and a clean file-based routing
structure. Zustand provides lightweight state management.

**PostgreSQL** offers reliable relational data storage with powerful
query capabilities. Its JSONB support handles semi-structured data like
chat message sources and extension configurations.

**LangchainGo** provides a Go-native interface for LLM orchestration,
supporting multiple providers through a unified API. Combined with
LangSmith, it offers full observability into AI operations.

**Typesense** delivers fast full-text search with typo tolerance and
relevance ranking, suitable for searching across conversation
transcripts and memory content.

**Milvus** provides high-performance vector similarity search for the
semantic retrieval pipeline. Its IVF_FLAT indexing and COSINE metric
enable fast approximate nearest neighbor queries, and its filtering
capabilities allow per-user data isolation at the vector level.
