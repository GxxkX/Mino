# Getting started

This guide helps you set up Mino and verify it is working correctly.
Follow these steps to get your personal AI assistant running.

You will install the backend services, configure the required
infrastructure, and test that everything is connected properly. By the
end of this guide, you will be able to log in and start recording voice
notes.

## Prerequisites

Before you begin, ensure you have the following installed on your
server.

- Docker Engine 20.10 or later
- Docker Compose v2

That's it. The `docker-compose.yml` in the project root includes all
infrastructure services (PostgreSQL, Redis, Milvus, MinIO, Typesense,
Whisper) and both application services (backend, web). No need to
install Go, Node.js, or any database manually.

If you prefer to run services individually or connect to existing
infrastructure, see the [manual setup](#manual-setup) section below.

## One-command deployment

The fastest way to get Mino running is with Docker Compose. Three
steps:

### 1. Configure environment

Copy the example environment file and fill in your values.

```bash
cp .env.example .env
```

Edit `.env` and update at minimum:

- `DB_PASSWORD` — PostgreSQL password
- `REDIS_PASSWORD` — Redis password
- `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` — MinIO credentials
- `TYPESENSE_API_KEY` — Typesense API key
- `LLM_API_KEY` — your LLM provider API key
- `STT_WHISPER_API_KEY` — Whisper API key

For a complete reference of all variables, see the
[Configuration](configuration.md) guide.

### 2. Generate JWT keys

```bash
mkdir -p backend/keys
openssl genrsa -out backend/keys/private.pem 2048
openssl rsa -in backend/keys/private.pem -pubout -out backend/keys/public.pem
```

### 3. Start everything

```bash
docker compose up -d
```

This builds and starts all services. The first run takes a few minutes
to pull images and build the backend and web containers. Wait for all
health checks to pass.

```bash
docker compose ps
```

Once all services show `healthy` or `running`, open your browser and
navigate to `http://localhost:3000`.

### Default ports

| Service | Port | Description |
|---------|------|-------------|
| Web UI | 3000 | Next.js frontend |
| Backend API | 8000 | Go REST API + WebSocket |
| PostgreSQL | 5432 | Database |
| Redis | 6379 | Cache and sessions |
| Milvus | 19530 | Vector database |
| MinIO API | 9000 | Object storage |
| MinIO Console | 9001 | MinIO web UI |
| Typesense | 8108 | Full-text search |
| Whisper | 33000 | Speech-to-text |

All host-side ports are configurable via `.env`. For example, set
`WEB_PORT=8080` to serve the web UI on port 8080 instead.

### How services connect

Inside Docker Compose, the backend automatically connects to all
infrastructure services using Docker DNS names (`postgres`, `redis`,
`milvus`, `minio`, `typesense`, `whisper`). You do not need to
configure any host addresses — the `docker-compose.yml` overrides
them for you.

The web frontend proxies API requests to the backend container
internally. From the browser, all API calls go through the Next.js
server at port 3000, which forwards them to the backend at port 8000.

## Database migrations

The backend automatically runs database migrations on startup. The
migration creates the following tables: `mino_users`,
`mino_conversations`, `mino_memories`, `mino_tasks`, `mino_tags`,
`mino_conversation_tags`, `mino_chat_sessions`, `mino_chat_messages`,
and `mino_extensions`.

The default admin user is also created automatically during startup
using the `ADMIN_USERNAME` and `ADMIN_PASSWORD` values from your `.env`
file.

## Log in and verify

Use the default administrator credentials to log in for the first time.

| Field | Value |
|-------|-------|
| Username | mino |
| Password | admin |

After logging in, you are redirected to the dashboard where you can see
your recordings, memories, tasks, and access the AI chat.

## Record your first note

To test the complete recording flow, follow these steps.

1. Open the dashboard in your web browser.
2. Click the microphone button in the header to start recording.
3. Speak naturally for a few seconds.
4. Click the button again to stop recording.
5. Wait for the transcription and AI processing to complete.
6. View the generated title, summary, action items, and memory points.

The recording appears in your dashboard history, where you can search
and review it later.

## Verify search

After creating a recording, verify that full-text search is working.

1. Open the dashboard.
2. Use the search bar in the header (or press Cmd+K).
3. Type a keyword from your recording.
4. Verify that search results appear.

If search results don't appear, trigger a manual reindex through the
API.

```bash
curl -X POST http://localhost:8000/v1/search/reindex \
  -H "Authorization: Bearer <your_token>"
```

## Verify semantic search

After creating a recording, verify that the Milvus vector search
pipeline is working. The backend logs indicate whether Milvus
connected successfully on startup. Look for these messages in the
server output.

```
connected to Milvus at localhost:19530 (db=default)
Milvus collections ready
```

To test semantic retrieval, open the AI chat and ask a question
related to your recording using different words than the original
transcript. For example, if you recorded a note about "scheduling a
team meeting next Tuesday," try asking "when is the next group
meeting?" The chat assistant retrieves relevant context through
vector similarity rather than exact keyword matching.

## Next steps

You now have a working Mino installation. Here are some things you can
explore next.

Deploy the system to production by following the
[Deployment](deployment.md) guide. This covers HTTPS configuration,
security hardening, and performance tuning.

Learn about the system architecture by reading the
[Architecture](architecture.md) documentation. Understanding how the
components work together helps with troubleshooting and customization.

Review the complete [API reference](api-reference.md) to understand all
available endpoints.

Connect additional clients such as the
[smartwatch app](watch-client.md) to your backend. Each client connects
to the same API and shares your data seamlessly.
