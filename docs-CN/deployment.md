# 部署

本指南涵盖将 Mino 部署到生产环境的相关内容，包括配置、安全、
性能和日常运维。

生产环境部署与本地开发有几个重要区别。网络流量必须加密，服务
需要进行健康监控，必须配置备份，并且系统应能优雅地处理故障。

## 生产环境要求

在部署到生产环境之前，请确保具备以下组件。

建议后端和服务使用至少 4 个 CPU 核心、8 GB 内存和 100 GB
磁盘空间的 Linux 服务器。根据音频录制量的不同，还需要额外的
存储空间。

将域名指向您的服务器，以便生成 HTTPS 证书并实现客户端的安全
访问。

需要 Docker Engine 20.10+ 和 Docker Compose v2。

## 使用 Docker Compose 部署

项目根目录下包含一个生产就绪的 `docker-compose.yml`，可通过
一条命令编排所有基础设施和应用服务。

### 快速开始

```bash
# 1. 配置
cp .env.example .env
vim .env                    # 填入密码、API 密钥等

# 2. 生成 JWT 密钥
mkdir -p backend/keys
openssl genrsa -out backend/keys/private.pem 2048
openssl rsa -in backend/keys/private.pem -pubout -out backend/keys/public.pem

# 3. 部署
docker compose up -d --build
```

### 部署内容

compose 文件启动以下服务，所有服务通过共享的 Docker 网络
互相连接：

| 服务 | 镜像 | 内部端口 |
|------|------|----------|
| `postgres` | postgres:16-alpine | 5432 |
| `redis` | redis:7-alpine | 6379 |
| `etcd` | coreos/etcd:v3.5.18 | 2379（内部） |
| `milvus-minio` | minio（Milvus 内部使用） | 9000（内部） |
| `milvus` | milvusdb/milvus:v2.6.0-rc1 | 19530 |
| `minio` | minio（应用存储） | 9000 / 9001 |
| `typesense` | typesense:27.1 | 8108 |
| `whisper` | 从 `./whisper` 构建 | 9000 |
| `backend` | 从 `./backend` 构建 | 8000 |
| `web` | 从 `./web` 构建 | 3000 |

后端容器的 environment 块会将所有 `*_HOST` 变量覆盖为
Docker 服务名称（`postgres`、`redis`、`milvus` 等），
因此你只需在 `.env` 中配置凭据和端口。

### 端口配置

所有宿主机端口均可通过 `.env` 配置：

```bash
APP_PORT=8000          # 后端 API
WEB_PORT=3000          # Web 前端
DB_PORT=5432           # PostgreSQL
REDIS_PORT=6379        # Redis
MILVUS_PORT=19530      # Milvus
MINIO_API_PORT=9000    # MinIO API
MINIO_CONSOLE_PORT=9001 # MinIO 控制台
TYPESENSE_PORT=8108    # Typesense
WHISPER_PORT=33000     # Whisper STT
```

### 健康检查与启动顺序

所有基础设施服务都包含健康检查。后端会等待 PostgreSQL、
Redis、Milvus、Typesense 和 MinIO 变为健康状态后才启动。
Web 服务会等待后端就绪。这确保了正确的启动顺序，无需手动
干预。

### 更新

拉取新代码后更新：

```bash
docker compose up -d --build
```

后端在启动时会自动运行数据库迁移。

### 独立 Dockerfile

如果你希望单独部署各服务（例如在 Kubernetes 或自定义编排器
中），可以使用以下独立 Dockerfile：

- `backend/Dockerfile` — 多阶段 Go 构建，生成最小化的
  Alpine 镜像，暴露 8000 端口。
- `web/Dockerfile` — 多阶段 Next.js 构建，使用 standalone
  输出模式，以非 root 用户运行，暴露 3000 端口。
- `Dockerfile`（根目录）— 一体化镜像，通过 supervisord
  同时运行后端和 Web。适用于单容器部署场景。

## 网络架构

在生产环境中，通常部署在反向代理之后，由反向代理处理 TLS 终止
和负载均衡。

推荐的架构是将 Nginx 放置在后端 API 和 Web 应用程序之前。
Nginx 负责 SSL/TLS 终止、Web 应用的静态文件服务，以及将 API
请求代理到后端。

为了实现更高的可用性，可以在 Nginx 后面运行多个后端实例，
使用 Redis 进行会话存储，以确保用户在不同实例间保持登录状态。
后端的速率限制器已经使用了 Redis，因此可以在多个实例间正确
工作。

## 配置 HTTPS

生产环境部署必须使用 HTTPS 来保护传输中的用户数据。获取证书
有两种方式。

### 方式一：Certbot

Certbot 可以从 Let's Encrypt 获取免费证书。安装后运行以下
命令。

```bash
sudo apt update
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d yourdomain.com
```

Certbot 会自动配置 Nginx 并设置证书续期。添加 cron 任务以
自动续期证书。

```bash
sudo crontab -e
0 0 * * * certbot renew --quiet
```

### 方式二：手动证书

如果您希望手动控制，可以从任何证书颁发机构获取证书，然后自行
配置 Nginx。

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/ssl/private/yourdomain.key \
  -out /etc/ssl/certs/yourdomain.crt
```

然后在 Nginx 的 server 块中配置使用这些文件。

## 加固后端安全

以下几项配置更改可以提升生产环境的安全性。

### 环境变量

使用生产环境专用的值更新您的 `.env` 文件。有关所有变量的完整
参考，请参阅[配置](configuration.md)指南。

```bash
APP_ENV=production
APP_DEBUG=false
APP_PORT=8000
```

禁用调试模式，以防止在错误响应中泄露敏感信息。

### 身份认证

在生产环境中，所有服务账户都应使用强密码。这包括 PostgreSQL、
Redis、MinIO、Typesense 和管理员用户。

首次登录后请立即更改默认管理员密码。默认凭据
（`mino`/`admin`）仅用于初始设置。

为生产环境专门生成新的 JWT 密钥，不要复用开发环境的密钥。

```bash
openssl genrsa -out keys/private.pem 2048
openssl rsa -in keys/private.pem -pubout -out keys/public.pem
```

请妥善保管私钥。任何拥有私钥访问权限的人都可以伪造令牌。

### 速率限制

默认速率限制为每个用户每个端点每分钟 100 次请求。速率限制器
使用基于 Redis 的滑动窗口计数器，在 Redis 不可用时会优雅降级。

监控速率限制违规情况以检测潜在的滥用行为。速率限制键在 Redis
中遵循 `ratelimit:{userID}:{path}` 的模式。

## 数据备份

定期备份可以防止因硬件故障或意外删除导致的数据丢失。

### PostgreSQL 备份

创建一个每日运行的备份脚本。

```bash
#!/bin/bash
BACKUP_DIR="/backups/postgres"
DATE=$(date +%Y%m%d_%H%M%S)
pg_dump -U postgres mino > "$BACKUP_DIR/mino_$DATE.sql"
find "$BACKUP_DIR" -type f -mtime +7 -delete
```

使用 cron 调度此脚本。

```bash
0 2 * * * /path/to/backup.sh
```

将备份存储在单独的卷或异地位置，以防止磁盘故障。备份包含全部
九张表：users、conversations、memories、tasks、tags、
conversation tags、chat sessions、chat messages 和
extensions。

### MinIO 备份

可以使用 MinIO 客户端 `mc` 备份 MinIO 数据。

```bash
mc mirror mino/data /backups/minio
```

与 PostgreSQL 备份类似，使用 cron 自动执行此操作。

### Typesense 备份

Typesense 数据可以通过重建索引端点从 PostgreSQL 重建。但是，
备份 Typesense 数据目录可以避免恢复期间的重建索引停机时间。

### Milvus 备份

Milvus 向量数据可以使用 Milvus Backup 工具或通过复制底层存储
卷进行备份。对于 Docker 部署，请备份 `milvus_data` 卷。

```bash
docker run --rm -v milvus_data:/data -v /backups:/backup \
  alpine tar czf /backup/milvus_$(date +%Y%m%d).tar.gz /data
```

如果丢失了 Milvus 数据，可以通过对所有对话和记忆重新生成
嵌入向量来从 PostgreSQL 重建。这需要为每条记录调用嵌入 API，
根据数据量的不同，可能需要较长时间并产生 API 费用。

### 配置备份

将配置文件、Nginx 设置和环境变量的副本保存在安全的位置。版本
控制非常适合此用途，但需要排除密码和 API 密钥等敏感值。

## 系统监控

监控有助于在问题影响用户之前发现并响应问题。

### 健康检查

后端提供了一个健康检查端点，用于报告各依赖项的状态。

```bash
curl http://localhost:8000/health
```

配置您的负载均衡器或编排系统使用此端点进行健康检查。

### 日志

后端通过 logrus 输出结构化 JSON 日志。每条请求日志包含 HTTP
方法、路径、查询参数、状态码、客户端 IP 和响应延迟。日志级别
根据状态码分配：5xx 为 error，4xx 为 warning，成功请求为
info。

配置集中式日志系统以聚合所有服务的日志。定期轮转日志文件以
防止磁盘空间耗尽。

### LangSmith 监控

如果您使用 LangSmith 进行 LLM 可观测性监控，请关注 LangSmith
仪表板上的 LLM 调用延迟、错误率和 token 使用量。在 `.env`
文件中设置 `LANGSMITH_TRACING=true` 以启用追踪。

### 指标

建议使用 Prometheus 收集指标，并使用 Grafana 进行可视化。
需要跟踪的关键指标包括请求延迟、错误率、数据库连接池使用情况
和存储容量。

## 水平扩展

随着使用量的增长，您可以对系统进行水平扩展。

### 后端扩展

在 Nginx 后面运行多个后端实例。后端是无状态的（会话存储在
Redis 中），因此可以直接添加实例而无需更改配置。

```yaml
# docker-compose.yml
services:
  backend:
    deploy:
      replicas: 3
```

更新 Nginx 配置以在各实例间进行负载均衡。

### 数据库注意事项

PostgreSQL 在适当索引的情况下可以处理大量负载。迁移文件已在
常用查询列上创建了索引（`user_id`、`status`、`due_date`、
`recorded_at`、`category`）。对于非常高的流量，可以考虑使用
只读副本来分散查询负载。

Milvus 支持集群模式以提高吞吐量。对于拥有大型向量集合的生产
工作负载，建议以分布式模式部署 Milvus，使用独立的查询节点和
数据节点。监控集合大小和搜索延迟，以确定何时需要扩展。

Typesense 可以通过集群实现高可用性。每个节点都维护一份数据
副本。

## 系统更新

定期更新可以修补安全漏洞并添加新功能。

### 后端更新

拉取最新代码并重新构建。

```bash
cd backend
git pull
go build -o mino-server ./cmd/server
systemctl restart mino
```

尽可能先在预发布环境中测试更新。后端在启动时会自动运行数据库
迁移，因此新版本启动时会自动应用数据库结构变更。

### Web 应用更新

为生产环境重新构建 Web 应用程序。

```bash
cd web
npm install
npm run build
```

构建输出可以由 Nginx 或任何静态托管服务提供。

## 事件响应

为潜在事件准备好文档化的处理流程。

### 服务不可用

如果后端变得不可用，请先检查服务状态。

```bash
systemctl status mino
journalctl -u mino -n 50
```

常见问题包括内存不足、数据库连接失败和配置错误。

### 数据恢复

如果发现数据丢失，请立即停止服务以防止进一步损害。从最近的
PostgreSQL 备份中恢复数据并调查原因。PostgreSQL 恢复后，可以
使用重建索引端点重建 Typesense 数据。Milvus 向量数据可以从
备份卷恢复，或通过从 PostgreSQL 重新嵌入对话和记忆来重新
生成。

```bash
curl -X POST http://localhost:8000/v1/search/reindex \
  -H "Authorization: Bearer <admin_token>"
```

### 安全事件

如果怀疑发生安全漏洞，请立即轮换所有凭据。这包括数据库密码、
API 密钥、JWT 密钥和管理员密码。查看结构化 JSON 日志以识别
未授权活动。

## 性能调优

以下几项调整可以提升生产环境的性能。

### 数据库优化

确保常用查询列上存在索引，并定期验证。

```sql
SELECT indexname, indexdef
FROM pg_indexes
WHERE tablename LIKE 'mino_%';
```

### Redis 配置

配置 Redis 进行数据持久化并使用足够的内存。监控内存使用情况，
并根据需要调整 `maxmemory` 设置。速率限制器的键会自动过期，
因此 Redis 内存使用量保持在可控范围内。

### 音频处理

如果转录延迟较高，可以考虑使用更快的转录服务提供商或运行本地
Whisper 实例。转录服务支持多个提供商，可通过
`LLM_PROVIDER` 环境变量进行配置。

## 客户端部署

### Web 应用程序

为生产环境构建 Web 应用程序。

```bash
cd web
npm run build
```

配置 Nginx 来提供构建后的应用程序并将 API 请求代理到后端。
Web 应用的 `next.config.js` 包含一条重写规则，将 `/api/v1/*`
代理到后端，因此请确保 `API_URL` 环境变量指向您的生产环境
后端。

### 手表应用

构建 Flutter 手表应用的发布版本。

```bash
cd watch
flutter build apk --release
```

通过您偏好的渠道分发 APK。手表应用连接到
`lib/core/constants/app_config.dart` 中配置的后端 URL。
构建前请将此值更新为指向您的生产服务器。
