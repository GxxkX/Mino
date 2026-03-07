# Configuration

This document provides a complete reference for all environment
variables used by the Mino backend. The backend reads configuration
from a `.env` file in the working directory using the `godotenv`
library.

Copy the example file to get started.

```bash
cd backend
cp configs/.env.example .env
```

## Application

These variables control the backend server behavior.

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_ENV` | `development` | Environment mode: `development`, `staging`, or `production` |
| `APP_PORT` | `8000` | HTTP server port |
| `APP_DEBUG` | `true` | Enable debug mode. Set to `false` in production. |

## PostgreSQL

Connection settings for the primary relational database.

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_NAME` | `mino` | Database name |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | | Database password |
| `DB_SSL_MODE` | `disable` | SSL mode: `disable`, `require`, `verify-ca`, `verify-full` |

## Redis

Connection settings for caching and rate limiting.

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_PASSWORD` | | Redis password |
| `REDIS_DB` | `0` | Redis database number |

## JWT authentication

Settings for JSON Web Token generation and validation. Mino uses RSA
RS256 signatures for JWT tokens.

| Variable | Default | Description |
|----------|---------|-------------|
| `JWT_PRIVATE_KEY_PATH` | `./keys/private.pem` | Path to RSA private key (PEM format) |
| `JWT_PUBLIC_KEY_PATH` | `./keys/public.pem` | Path to RSA public key (PEM format) |
| `JWT_ACCESS_TOKEN_EXPIRE` | `15m` | Access token lifetime (Go duration format) |
| `JWT_REFRESH_TOKEN_EXPIRE` | `168h` | Refresh token lifetime (168h = 7 days) |

Generate JWT keys with the following commands.

```bash
mkdir -p keys
openssl genrsa -out keys/private.pem 2048
openssl rsa -in keys/private.pem -pubout -out keys/public.pem
```

## Default admin account

The backend creates a default admin user on startup if one does not
exist.

| Variable | Default | Description |
|----------|---------|-------------|
| `ADMIN_USERNAME` | `mino` | Default admin username |
| `ADMIN_PASSWORD` | `admin` | Default admin password |

Change the admin password immediately after first login in production.

## LLM

Settings for the large language model used by ChatAgent and
ExtractAgent, and the embedding model used for semantic search. Mino
uses LangchainGo to call OpenAI-compatible APIs.

| Variable | Default | Description |
|----------|---------|-------------|
| `LLM_PROVIDER` | `openai` | LLM provider: `openai`, `zhipu`, or `ollama` |
| `LLM_API_KEY` | | API key for the LLM provider |
| `LLM_BASE_URL` | | Custom API base URL (required for `zhipu` and `ollama`) |
| `LLM_MODEL` | `gpt-4o` | Chat model name |
| `LLM_EMBEDDING_MODEL` | `embedding-3` | Embedding model name for vector generation |

The chat model and embedding model are separate. The chat model handles
conversational AI and structured extraction, while the embedding model
generates vector representations for semantic search. Both share the
same API key and base URL.

Common embedding model choices by provider:

| Provider | Embedding model | Dimensions |
|----------|----------------|------------|
| Zhipu | `embedding-3` | 1024 |
| OpenAI | `text-embedding-3-small` | 1536 |
| OpenAI | `text-embedding-ada-002` | 1536 |
| Ollama | `nomic-embed-text` | 768 |

> **Note:** The default Milvus collection dimension is 1024 (matching
> Zhipu `embedding-3`). If you use a model with a different output
> dimension, update the `DefaultEmbeddingDim` constant in
> `internal/pkg/vectordb/milvus.go` before creating collections.

### Provider-specific configuration

**OpenAI**: Set `LLM_PROVIDER=openai` and provide your API key. The
base URL defaults to the OpenAI API.

**Zhipu**: Set `LLM_PROVIDER=zhipu`, provide your API key, and set
`LLM_BASE_URL` to the Zhipu API endpoint.

**Ollama**: Set `LLM_PROVIDER=ollama` and set `LLM_BASE_URL` to your
Ollama instance (for example, `http://localhost:11434`). No API key is
needed.

## Milvus

Connection settings for the vector database. Milvus stores embeddings
for semantic similarity search. On startup, the backend connects to
Milvus, creates any missing collections, and loads them into memory.
If Milvus is unreachable, the backend continues without vector search
and falls back to keyword matching.

| Variable | Default | Description |
|----------|---------|-------------|
| `MILVUS_HOST` | `localhost` | Milvus host |
| `MILVUS_PORT` | `19530` | Milvus port |
| `MILVUS_USER` | | Milvus username |
| `MILVUS_PASSWORD` | | Milvus password |
| `MILVUS_DB_NAME` | `default` | Milvus database name |
| `MILVUS_CONVERSATIONS_COLLECTION` | `conversations` | Collection name for conversation vectors |
| `MILVUS_MEMORIES_COLLECTION` | `memories` | Collection name for memory vectors |

## MinIO

Connection settings for S3-compatible object storage. MinIO stores
audio files.

| Variable | Default | Description |
|----------|---------|-------------|
| `MINIO_ENDPOINT` | `localhost:9000` | MinIO endpoint (host:port) |
| `MINIO_ACCESS_KEY` | | MinIO access key |
| `MINIO_SECRET_KEY` | | MinIO secret key |
| `MINIO_SECURE` | `false` | Use HTTPS for MinIO connections |
| `MINIO_REGION` | `us-east-1` | MinIO region |
| `MINIO_PUBLIC_URL` | | Public URL for accessing stored files |

## Typesense

Connection settings for the full-text search engine. Typesense indexes
conversations and memories for keyword search.

| Variable | Default | Description |
|----------|---------|-------------|
| `TYPESENSE_HOST` | `localhost` | Typesense host |
| `TYPESENSE_HOST_PORT` | `8108` | Typesense port |
| `TYPESENSE_API_KEY` | | Typesense API key |

## LangSmith

Settings for LLM observability and tracing. LangSmith is optional and
disabled by default.

| Variable | Default | Description |
|----------|---------|-------------|
| `LANGSMITH_TRACING` | `false` | Enable LangSmith tracing |
| `LANGSMITH_API_KEY` | | LangSmith API key |
| `LANGSMITH_PROJECT` | `mino-backend-chat` | LangSmith project name |
| `LANGSMITH_ENDPOINT` | `https://api.smith.langchain.com` | LangSmith API endpoint |
| `OMI_LANGSMITH_AGENTIC_PROMPT_NAME` | `mino-agentic-system` | LangSmith prompt name |
| `OMI_LANGSMITH_PROMPT_CACHE_TTL_SECONDS` | `300` | Prompt cache TTL in seconds |

## Example configuration

A minimal `.env` file for local development.

```bash
# Application
APP_ENV=development
APP_PORT=8000
APP_DEBUG=true

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_NAME=mino
DB_USER=postgres
DB_PASSWORD=your_password
DB_SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password
REDIS_DB=0

# JWT
JWT_PRIVATE_KEY_PATH=./keys/private.pem
JWT_PUBLIC_KEY_PATH=./keys/public.pem
JWT_ACCESS_TOKEN_EXPIRE=15m
JWT_REFRESH_TOKEN_EXPIRE=168h

# Admin
ADMIN_USERNAME=mino
ADMIN_PASSWORD=admin

# LLM
LLM_PROVIDER=openai
LLM_API_KEY=your_api_key
LLM_MODEL=gpt-4o
LLM_EMBEDDING_MODEL=text-embedding-3-small

# Typesense
TYPESENSE_HOST=localhost
TYPESENSE_HOST_PORT=8108
TYPESENSE_API_KEY=your_typesense_key

# MinIO
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minio_user
MINIO_SECRET_KEY=minio_password
MINIO_SECURE=false

# Milvus
MILVUS_HOST=localhost
MILVUS_PORT=19530
```

## Next steps

After configuring the backend, follow the
[Getting started](getting-started.md) guide to run the application.
For production-specific settings, see the
[Deployment](deployment.md) guide.
