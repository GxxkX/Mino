# 配置

本文档提供了 Mino 后端使用的所有环境变量的完整参考。
后端通过 `godotenv` 库从工作目录下的 `.env` 文件中读取配置。

复制示例文件以快速开始。

```bash
cd backend
cp configs/.env.example .env
```

## 应用

以下变量控制后端服务器的行为。

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `APP_ENV` | `development` | 环境模式：`development`、`staging` 或 `production` |
| `APP_PORT` | `8000` | HTTP 服务器端口 |
| `APP_DEBUG` | `true` | 启用调试模式。在生产环境中应设为 `false`。 |

## PostgreSQL

主关系型数据库的连接设置。

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `DB_HOST` | `localhost` | PostgreSQL 主机地址 |
| `DB_PORT` | `5432` | PostgreSQL 端口 |
| `DB_NAME` | `mino` | 数据库名称 |
| `DB_USER` | `postgres` | 数据库用户名 |
| `DB_PASSWORD` | | 数据库密码 |
| `DB_SSL_MODE` | `disable` | SSL 模式：`disable`、`require`、`verify-ca`、`verify-full` |

## Redis

缓存和速率限制的连接设置。

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `REDIS_HOST` | `localhost` | Redis 主机地址 |
| `REDIS_PORT` | `6379` | Redis 端口 |
| `REDIS_PASSWORD` | | Redis 密码 |
| `REDIS_DB` | `0` | Redis 数据库编号 |

## JWT 认证

JSON Web Token 生成和验证的相关设置。Mino 使用 RSA RS256
签名算法来签发 JWT 令牌。

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `JWT_PRIVATE_KEY_PATH` | `./keys/private.pem` | RSA 私钥路径（PEM 格式） |
| `JWT_PUBLIC_KEY_PATH` | `./keys/public.pem` | RSA 公钥路径（PEM 格式） |
| `JWT_ACCESS_TOKEN_EXPIRE` | `15m` | 访问令牌有效期（Go duration 格式） |
| `JWT_REFRESH_TOKEN_EXPIRE` | `168h` | 刷新令牌有效期（168h = 7 天） |

使用以下命令生成 JWT 密钥。

```bash
mkdir -p keys
openssl genrsa -out keys/private.pem 2048
openssl rsa -in keys/private.pem -pubout -out keys/public.pem
```

## 默认管理员账户

后端在启动时会创建一个默认管理员用户（如果尚不存在）。

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `ADMIN_USERNAME` | `mino` | 默认管理员用户名 |
| `ADMIN_PASSWORD` | `admin` | 默认管理员密码 |

在生产环境中首次登录后，请立即更改管理员密码。

## LLM

ChatAgent 和 ExtractAgent 使用的大语言模型设置，以及用于
语义搜索的嵌入模型设置。Mino 使用 LangchainGo 调用
OpenAI 兼容的 API。

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `LLM_PROVIDER` | `openai` | LLM 提供商：`openai`、`zhipu` 或 `ollama` |
| `LLM_API_KEY` | | LLM 提供商的 API 密钥 |
| `LLM_BASE_URL` | | 自定义 API 基础 URL（`zhipu` 和 `ollama` 必填） |
| `LLM_MODEL` | `gpt-4o` | 对话模型名称 |
| `LLM_EMBEDDING_MODEL` | `embedding-3` | 用于向量生成的嵌入模型名称 |

对话模型和嵌入模型是分开的。对话模型负责对话式 AI 和
结构化提取，而嵌入模型负责生成用于语义搜索的向量表示。
两者共享相同的 API 密钥和基础 URL。

各提供商常用的嵌入模型：

| 提供商 | 嵌入模型 | 维度 |
|--------|----------|------|
| Zhipu | `embedding-3` | 1024 |
| OpenAI | `text-embedding-3-small` | 1536 |
| OpenAI | `text-embedding-ada-002` | 1536 |
| Ollama | `nomic-embed-text` | 768 |

> **注意：** Milvus 集合的默认维度为 1024（与 Zhipu
> `embedding-3` 匹配）。如果使用输出维度不同的模型，
> 请在创建集合之前更新
> `internal/pkg/vectordb/milvus.go` 中的
> `DefaultEmbeddingDim` 常量。

### 各提供商的具体配置

**OpenAI**：设置 `LLM_PROVIDER=openai` 并提供 API 密钥。
基础 URL 默认为 OpenAI API 地址。

**Zhipu**：设置 `LLM_PROVIDER=zhipu`，提供 API 密钥，
并将 `LLM_BASE_URL` 设置为 Zhipu API 端点。

**Ollama**：设置 `LLM_PROVIDER=ollama`，并将
`LLM_BASE_URL` 设置为 Ollama 实例地址
（例如 `http://localhost:11434`）。无需 API 密钥。

## Milvus

向量数据库的连接设置。Milvus 存储用于语义相似度搜索的
嵌入向量。启动时，后端会连接 Milvus，创建缺失的集合，
并将其加载到内存中。如果 Milvus 不可达，后端将继续运行，
但不提供向量搜索功能，转而使用关键词匹配作为降级方案。

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `MILVUS_HOST` | `localhost` | Milvus 主机地址 |
| `MILVUS_PORT` | `19530` | Milvus 端口 |
| `MILVUS_USER` | | Milvus 用户名 |
| `MILVUS_PASSWORD` | | Milvus 密码 |
| `MILVUS_DB_NAME` | `default` | Milvus 数据库名称 |
| `MILVUS_CONVERSATIONS_COLLECTION` | `conversations` | 对话向量的集合名称 |
| `MILVUS_MEMORIES_COLLECTION` | `memories` | 记忆向量的集合名称 |

## MinIO

S3 兼容对象存储的连接设置。MinIO 用于存储音频文件。

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `MINIO_ENDPOINT` | `localhost:9000` | MinIO 端点（host:port） |
| `MINIO_ACCESS_KEY` | | MinIO 访问密钥 |
| `MINIO_SECRET_KEY` | | MinIO 秘密密钥 |
| `MINIO_SECURE` | `false` | 是否使用 HTTPS 连接 MinIO |
| `MINIO_REGION` | `us-east-1` | MinIO 区域 |
| `MINIO_PUBLIC_URL` | | 用于访问存储文件的公开 URL |

## Typesense

全文搜索引擎的连接设置。Typesense 为对话和记忆建立索引，
以支持关键词搜索。

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `TYPESENSE_HOST` | `localhost` | Typesense 主机地址 |
| `TYPESENSE_HOST_PORT` | `8108` | Typesense 端口 |
| `TYPESENSE_API_KEY` | | Typesense API 密钥 |

## 说话人分离（Pyannote）

Mino 支持基于 [pyannote.audio](https://github.com/pyannote/pyannote-audio)
的说话人分离功能。启用后，系统会自动识别录音中的不同说话人，
并将其声纹嵌入与 Milvus 中已存储的声纹进行匹配。

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `PYANNOTE_ENABLED` | `false` | 启用说话人分离 |
| `PYANNOTE_HF_TOKEN` | | Hugging Face 访问令牌 |
| `SPEAKER_SIMILARITY_THRESHOLD` | `0.65` | 声纹匹配的余弦相似度阈值 |

> **⚠️ 注意：** Pyannote 模型属于 Hugging Face 上的**受限模型（Gated Model）**。
> 使用前必须完成以下步骤：
>
> 1. 注册 [Hugging Face](https://huggingface.co) 账号。
> 2. 访问以下每个模型页面，阅读并同意许可协议：
>    - [pyannote/speaker-diarization-3.1](https://huggingface.co/pyannote/speaker-diarization-3.1)
>    - [pyannote/segmentation-3.0](https://huggingface.co/pyannote/segmentation-3.0)
>    - [pyannote/embedding](https://huggingface.co/pyannote/embedding)
> 3. 在 [huggingface.co/settings/tokens](https://huggingface.co/settings/tokens)
>    生成访问令牌（至少需要 `read` 权限）。
> 4. 将令牌设置为 `.env` 文件中的 `PYANNOTE_HF_TOKEN`。
>
> 如果未完成上述步骤，whisper 服务将无法下载模型，
> 说话人分离功能将不可用。

启用后，需要重新构建 whisper 容器以在启动时预加载模型。

```bash
docker compose build --no-cache whisper
docker compose up -d whisper
```

## LangSmith

LLM 可观测性和链路追踪设置。LangSmith 为可选组件，
默认处于禁用状态。

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `LANGSMITH_TRACING` | `false` | 启用 LangSmith 链路追踪 |
| `LANGSMITH_API_KEY` | | LangSmith API 密钥 |
| `LANGSMITH_PROJECT` | `mino-backend-chat` | LangSmith 项目名称 |
| `LANGSMITH_ENDPOINT` | `https://api.smith.langchain.com` | LangSmith API 端点 |
| `OMI_LANGSMITH_AGENTIC_PROMPT_NAME` | `mino-agentic-system` | LangSmith 提示词名称 |
| `OMI_LANGSMITH_PROMPT_CACHE_TTL_SECONDS` | `300` | 提示词缓存 TTL（秒） |

## 配置示例

用于本地开发的最小化 `.env` 文件。

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

# STT
STT_PROVIDER=whisper

# 说话人分离（可选）
PYANNOTE_ENABLED=false
PYANNOTE_HF_TOKEN=hf_your_token_here
```

## 后续步骤

完成后端配置后，请参阅[快速开始](getting-started.md)指南
来运行应用程序。有关生产环境的具体设置，请参阅
[部署](deployment.md)指南。
