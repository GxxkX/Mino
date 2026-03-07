 <p align="center">
  <img src="./logo.png" alt="Mino" width="120" />
</p>

# Mino

Mino is a privacy-first, fully self-deployable personal AI assistant. It
transforms everyday voice into structured digital assets through real-time
transcription and deep semantic analysis, automatically extracting structured
insights and daily memos.

This documentation provides comprehensive guidance on understanding,
installing, and using Mino. Whether you are a developer looking to
contribute or an end user setting up your personal AI assistant, you will
find the information you need here.

## What you can do with Mino

Mino enables you to capture voice notes from any device and automatically
transform them into organized, searchable content. The system handles the
entire pipeline from audio recording through transcription to structured
data extraction, all while keeping your data under your complete control.

Record voice notes from your smartwatch, phone, web browser, or desktop
application. Each recording is automatically transcribed in real-time and
processed by AI to extract titles, summaries, action items, and memory
points. Later, you can search through your entire history using natural
language queries or chat with your AI assistant about past conversations.

The system supports multiple devices working together seamlessly. Start a
recording on your watch and review the results on your phone or computer.
All your data stays synchronized across devices while remaining entirely
under your control on your own infrastructure.

## Core features

Mino provides a complete suite of features for voice note capture and
intelligent retrieval.

**Real-time voice transcription** captures audio from multiple input
sources and transcribes it with minimal latency. The system uses WebSocket
connections for low-latency streaming and supports offline recording when
network connectivity is unavailable.

**Structured information extraction** uses large language models to analyze
your recordings and automatically generate meaningful titles, concise
summaries, actionable task lists, and important memory points. The
ExtractAgent processes transcripts and outputs structured JSON data.

**Semantic search and full-text search** combine multiple retrieval
strategies to help you find exactly what you need. Milvus provides
semantic similarity search by generating vector embeddings for
conversations and memories using LangchainGo. Typesense provides fast
full-text keyword matching as a complementary retrieval path.

**AI-powered chat** lets you ask questions about your past conversations.
The ChatAgent retrieves relevant context from your history using a
retrieval-augmented generation (RAG) pipeline and provides answers with
source citations. Chat is organized into sessions that you can create,
rename, and delete.

**Extensions** allow you to add custom functionality to Mino. You can
create, enable, disable, and configure extensions through the web
interface or API.

**Multi-device support** ensures you can record and review from any
device. The supported clients are smartwatches (Flutter), phones
(Flutter), web browsers (Next.js), and desktop applications (Go + Wails).

## Technology stack

Mino is built with a modern, scalable architecture designed for
self-deployment.

The backend is written in Go 1.24 using the Gin framework, providing
RESTful APIs and WebSocket endpoints for real-time communication.
PostgreSQL stores structured data such as user information, conversations,
memories, tasks, chat sessions, and extensions. Milvus stores vector
embeddings for semantic similarity search across conversations and
memories. Typesense provides full-text search capabilities across
conversations and memories. MinIO stores audio files, and Redis caches
session data and provides rate limiting.

AI capabilities use LangchainGo for LLM orchestration and embedding
generation, with support for OpenAI, Zhipu, and Ollama providers.
LangSmith integration provides observability for LLM calls. The system
implements a dual-agent architecture: ChatAgent for conversational
interactions and ExtractAgent for structured information extraction
from transcripts. An `EmbeddingProvider` interface generates vector
embeddings that power the semantic search pipeline.

The web interface uses Next.js 14 with React 18, Zustand for state
management, and Tailwind CSS for styling. The watch application is built
with Flutter using the Provider pattern for state management. Mobile and
desktop applications follow similar architectures.

## Quick reference

This table shows which features are available on each client platform.

| Feature | Watch | Phone | Web | Desktop |
|---------|-------|-------|-----|---------|
| Record and transcribe | Yes | Yes | Yes | Yes |
| Review history | No | Yes | Yes | Yes |
| Manage memories | No | Yes | Yes | Yes |
| Manage tasks | No | Yes | Yes | Yes |
| Browse audio files | No | Yes | Yes | Yes |
| AI chat | No | Yes | Yes | Yes |
| App extensions | No | Yes | Yes | Yes |
| Global search | No | Yes | Yes | Yes |
| System settings | No | Yes | Yes | Yes |
| Cloud settings | No | Yes | Yes | Yes |
| MCP configuration | No | Yes | Yes | Yes |

## Default access

When you first deploy Mino, you can log in with the default administrator
account.

| Field | Value |
|-------|-------|
| Username | mino |
| Password | admin |
| Role | admin |

Change this password immediately after your first login for security.

## Documentation overview

This documentation is organized into the following guides.

| Guide | Description |
|-------|-------------|
| [Getting started](./docs/getting-started.md) | Install and run Mino |
| [Architecture](./docs/architecture.md) | System design and components |
| [API reference](./docs/api-reference.md) | Complete REST and WebSocket API |
| [Configuration](./docs/configuration.md) | Environment variables reference |
| [Web client](./docs/web-client.md) | Web application guide |
| [Watch client](./docs/watch-client.md) | Smartwatch application guide |
| [Deployment](./docs/deployment.md) | Production deployment guide |

## Next steps

If you want to get Mino running quickly, see the
[Getting started](./docs/getting-started.md) guide. It walks through the
installation process and shows you how to verify your deployment is
working correctly.

If you need to understand how the system is designed, see the
[Architecture](./docs/architecture.md) documentation. It explains the system
components, data flows, and design decisions.

If you are preparing for production deployment, see the
[Deployment](./docs/deployment.md) guide. It covers configuration, security,
and operational considerations.
