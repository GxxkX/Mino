# 快速入门

本指南帮助你完成 Mino 的安装配置，并验证其是否正常运行。
按照以下步骤启动你的个人 AI 助手。

你将安装后端服务、配置所需的基础设施，并测试所有组件是否
正确连接。完成本指南后，你将能够登录系统并开始录制语音
笔记。

## 前置条件

在开始之前，请确保你的服务器上已安装以下软件。

- Docker Engine 20.10 或更高版本
- Docker Compose v2

仅此而已。项目根目录下的 `docker-compose.yml` 已包含
所有基础设施服务（PostgreSQL、Redis、Milvus、MinIO、
Typesense、Whisper）以及两个应用服务（后端、Web 前端）。
无需手动安装 Go、Node.js 或任何数据库。

如果你希望单独运行各服务或连接到已有的基础设施，请参阅
下方的[手动部署](#手动部署)章节。

## 一键部署

使用 Docker Compose 是启动 Mino 最快的方式，只需三步：

### 1. 配置环境变量

复制示例环境文件并填入你的配置值。

```bash
cp .env.example .env
```

编辑 `.env`，至少需要更新以下配置：

- `DB_PASSWORD` — PostgreSQL 密码
- `REDIS_PASSWORD` — Redis 密码
- `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` — MinIO 凭据
- `TYPESENSE_API_KEY` — Typesense API 密钥
- `LLM_API_KEY` — LLM 提供商 API 密钥
- `STT_WHISPER_API_KEY` — Whisper API 密钥

有关所有环境变量的完整参考，请参阅
[配置指南](configuration.md)。

### 2. 生成 JWT 密钥

```bash
mkdir -p backend/keys
openssl genrsa -out backend/keys/private.pem 2048
openssl rsa -in backend/keys/private.pem -pubout -out backend/keys/public.pem
```

### 3. 启动所有服务

```bash
docker compose up -d
```

首次运行需要几分钟来拉取镜像并构建后端和 Web 容器。
等待所有健康检查通过。

```bash
docker compose ps
```

当所有服务显示 `healthy` 或 `running` 后，打开浏览器
访问 `http://localhost:3000`。

### 默认端口

| 服务 | 端口 | 说明 |
|------|------|------|
| Web 界面 | 3000 | Next.js 前端 |
| 后端 API | 8000 | Go REST API + WebSocket |
| PostgreSQL | 5432 | 数据库 |
| Redis | 6379 | 缓存和会话 |
| Milvus | 19530 | 向量数据库 |
| MinIO API | 9000 | 对象存储 |
| MinIO 控制台 | 9001 | MinIO Web 管理界面 |
| Typesense | 8108 | 全文搜索 |
| Whisper | 33000 | 语音转文字 |

所有宿主机端口均可通过 `.env` 配置。例如，设置
`WEB_PORT=8080` 可将 Web 界面改为 8080 端口。

### 服务连接方式

在 Docker Compose 内部，后端通过 Docker DNS 名称
（`postgres`、`redis`、`milvus`、`minio`、`typesense`、
`whisper`）自动连接到所有基础设施服务。你无需配置任何
主机地址——`docker-compose.yml` 已自动覆盖这些配置。

Web 前端在内部将 API 请求代理到后端容器。从浏览器发出
的所有 API 调用都通过 3000 端口的 Next.js 服务器转发
到 8000 端口的后端。

## 数据库迁移

后端在启动时会自动运行数据库迁移。迁移过程会创建以下
数据表：`mino_users`、`mino_conversations`、
`mino_memories`、`mino_tasks`、`mino_tags`、
`mino_conversation_tags`、`mino_chat_sessions`、
`mino_chat_messages` 和 `mino_extensions`。

默认管理员用户也会在启动时自动创建，使用的是 `.env`
文件中 `ADMIN_USERNAME` 和 `ADMIN_PASSWORD` 的值。

## 登录并验证

使用默认管理员凭据进行首次登录。

| 字段 | 值 |
|------|------|
| 用户名 | mino |
| 密码 | admin |

登录后，你将被重定向到仪表盘，在这里可以查看你的录音、
记忆、任务，以及访问 AI 对话功能。

## 录制你的第一条笔记

按照以下步骤测试完整的录音流程。

1. 在浏览器中打开仪表盘。
2. 点击顶部栏中的麦克风按钮开始录音。
3. 自然地说几秒钟。
4. 再次点击按钮停止录音。
5. 等待转录和 AI 处理完成。
6. 查看生成的标题、摘要、待办事项和记忆要点。

录音会出现在仪表盘的历史记录中，你可以随时搜索和回顾。

## 验证搜索功能

创建录音后，验证全文搜索是否正常工作。

1. 打开仪表盘。
2. 使用顶部栏中的搜索框（或按 Cmd+K）。
3. 输入录音中的一个关键词。
4. 确认搜索结果是否出现。

如果搜索结果未出现，请通过 API 触发手动重建索引。

```bash
curl -X POST http://localhost:8000/v1/search/reindex \
  -H "Authorization: Bearer <your_token>"
```

## 验证语义搜索

创建录音后，验证 Milvus 向量搜索管道是否正常工作。
后端日志会显示 Milvus 是否在启动时成功连接。请在服务器
输出中查找以下信息。

```
connected to Milvus at localhost:19530 (db=default)
Milvus collections ready
```

要测试语义检索，请打开 AI 对话，使用与原始转录不同的
措辞提出一个与你的录音相关的问题。例如，如果你录制了
一条关于"下周二安排团队会议"的笔记，可以尝试问"下次
小组会议是什么时候？"AI 助手会通过向量相似度而非精确
关键词匹配来检索相关上下文。

## 后续步骤

你现在已经拥有一个可以正常运行的 Mino 实例。以下是一些
你可以进一步探索的内容。

按照[部署指南](deployment.md)将系统部署到生产环境。
该指南涵盖 HTTPS 配置、安全加固和性能调优。

阅读[架构文档](architecture.md)了解系统架构。理解各
组件之间的协作方式有助于故障排查和自定义开发。

查阅完整的 [API 参考文档](api-reference.md)以了解所有
可用的端点。

将其他客户端（如[智能手表应用](watch-client.md)）连接
到你的后端。每个客户端都连接到同一个 API，数据无缝共享。
