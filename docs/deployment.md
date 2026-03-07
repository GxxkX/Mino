# Deployment

This guide covers deploying Mino to production. It addresses
configuration, security, performance, and ongoing operations.

A production deployment differs from local development in several
important ways. Network traffic must be encrypted, services should be
monitored for health, backups must be configured, and the system should
handle failure gracefully.

## Production requirements

Before deploying to production, ensure you have the following components
available.

A Linux server with at least 4 CPU cores, 8 GB of RAM, and 100 GB of
disk space is recommended for the backend and services. Additional
storage is needed depending on your audio recording volume.

A domain name pointing to your server enables HTTPS certificate
generation and secure access from clients.

Docker Engine 20.10+ and Docker Compose v2 are required.

## Deploy with Docker Compose

The project includes a production-ready `docker-compose.yml` at the
repository root. It orchestrates all infrastructure and application
services in a single command.

### Quick start

```bash
# 1. Configure
cp .env.example .env
vim .env                    # fill in passwords, API keys, etc.

# 2. Generate JWT keys
mkdir -p backend/keys
openssl genrsa -out backend/keys/private.pem 2048
openssl rsa -in backend/keys/private.pem -pubout -out backend/keys/public.pem

# 3. Deploy
docker compose up -d --build
```

### What gets deployed

The compose file starts the following services, all connected via a
shared Docker network:

| Service | Image | Internal Port |
|---------|-------|---------------|
| `postgres` | postgres:16-alpine | 5432 |
| `redis` | redis:7-alpine | 6379 |
| `etcd` | coreos/etcd:v3.5.18 | 2379 (internal) |
| `milvus-minio` | minio (internal to Milvus) | 9000 (internal) |
| `milvus` | milvusdb/milvus:v2.6.0-rc1 | 19530 |
| `minio` | minio (app storage) | 9000 / 9001 |
| `typesense` | typesense:27.1 | 8108 |
| `whisper` | built from `./whisper` | 9000 |
| `backend` | built from `./backend` | 8000 |
| `web` | built from `./web` | 3000 |

The backend container's environment block overrides all `*_HOST`
variables to use Docker service names (`postgres`, `redis`, `milvus`,
etc.), so you only need to configure credentials and ports in `.env`.

### Port configuration

All host-side ports are configurable via `.env`:

```bash
APP_PORT=8000          # Backend API
WEB_PORT=3000          # Web frontend
DB_PORT=5432           # PostgreSQL
REDIS_PORT=6379        # Redis
MILVUS_PORT=19530      # Milvus
MINIO_API_PORT=9000    # MinIO API
MINIO_CONSOLE_PORT=9001 # MinIO Console
TYPESENSE_PORT=8108    # Typesense
WHISPER_PORT=33000     # Whisper STT
```

### Health checks and startup order

All infrastructure services include health checks. The backend waits
for PostgreSQL, Redis, Milvus, Typesense, and MinIO to become healthy
before starting. The web service waits for the backend. This ensures
correct startup order without manual intervention.

### Updating

To update after pulling new code:

```bash
docker compose up -d --build
```

The backend runs database migrations automatically on startup.

### Standalone Dockerfiles

If you prefer to deploy services individually (for example, behind
Kubernetes or a custom orchestrator), standalone Dockerfiles are
available:

- `backend/Dockerfile` — Multi-stage Go build, produces a minimal
  Alpine image exposing port 8000.
- `web/Dockerfile` — Multi-stage Next.js build with standalone output,
  runs as non-root user, exposes port 3000.
- `Dockerfile` (root) — All-in-one image that runs both backend and
  web via supervisord. Useful for single-container deployments.

## Network architecture

In production, you typically deploy behind a reverse proxy that handles
TLS termination and load balancing.

The recommended architecture places Nginx in front of both the backend
API and the web application. Nginx handles SSL/TLS termination, static
file serving for the web app, and proxying API requests to the backend.

For higher availability, run multiple backend instances behind Nginx,
using Redis for session storage to ensure users stay logged in across
instances. The backend's rate limiter already uses Redis, so it works
correctly across multiple instances.

## Configure HTTPS

Production deployments must use HTTPS to protect user data in transit.
You have two options for obtaining certificates.

### Option 1: Certbot

Certbot obtains free certificates from Let's Encrypt. Install it and
run the following commands.

```bash
sudo apt update
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d yourdomain.com
```

Certbot automatically configures Nginx and sets up certificate renewal.
Add a cron job to renew certificates automatically.

```bash
sudo crontab -e
0 0 * * * certbot renew --quiet
```

### Option 2: Manual certificate

If you prefer manual control, obtain certificates from any certificate
authority and configure Nginx yourself.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/ssl/private/yourdomain.key \
  -out /etc/ssl/certs/yourdomain.crt
```

Then configure Nginx to use these files in your server block.

## Secure the backend

Several configuration changes improve security in production.

### Environment variables

Update your `.env` file with production-specific values. For a complete
reference of all variables, see the
[Configuration](configuration.md) guide.

```bash
APP_ENV=production
APP_DEBUG=false
APP_PORT=8000
```

Disable debug mode to prevent leaking sensitive information in error
responses.

### Authentication

In production, use strong passwords for all service accounts. This
includes PostgreSQL, Redis, MinIO, Typesense, and the admin user.

Change the default admin password immediately after first login. The
default credentials (`mino`/`admin`) are intended for initial setup
only.

Generate new JWT keys specifically for production use. Do not reuse
keys from development.

```bash
openssl genrsa -out keys/private.pem 2048
openssl rsa -in keys/private.pem -pubout -out keys/public.pem
```

Keep the private key secure. Anyone with access to it can forge tokens.

### Rate limiting

The default rate limit is 100 requests per minute per user per
endpoint. The rate limiter uses Redis-based sliding window counters and
degrades gracefully if Redis becomes unavailable.

Monitor rate limit violations to detect potential abuse. Rate limit
keys follow the pattern `ratelimit:{userID}:{path}` in Redis.

## Back up data

Regular backups protect against data loss from hardware failure or
accidental deletion.

### PostgreSQL backup

Create a backup script that runs daily.

```bash
#!/bin/bash
BACKUP_DIR="/backups/postgres"
DATE=$(date +%Y%m%d_%H%M%S)
pg_dump -U postgres mino > "$BACKUP_DIR/mino_$DATE.sql"
find "$BACKUP_DIR" -type f -mtime +7 -delete
```

Schedule this script with cron.

```bash
0 2 * * * /path/to/backup.sh
```

Store backups on a separate volume or offsite to protect against disk
failure. The backup includes all nine tables: users, conversations,
memories, tasks, tags, conversation tags, chat sessions, chat messages,
and extensions.

### MinIO backup

MinIO data can be backed up using `mc`, the MinIO client.

```bash
mc mirror mino/data /backups/minio
```

Automate this with cron, similar to the PostgreSQL backup.

### Typesense backup

Typesense data can be rebuilt from PostgreSQL using the reindex
endpoint. However, backing up the Typesense data directory avoids
reindex downtime during recovery.

### Milvus backup

Milvus vector data can be backed up using the Milvus Backup tool or
by copying the underlying storage volumes. For Docker deployments,
back up the `milvus_data` volume.

```bash
docker run --rm -v milvus_data:/data -v /backups:/backup \
  alpine tar czf /backup/milvus_$(date +%Y%m%d).tar.gz /data
```

If you lose Milvus data, the vectors can be regenerated from
PostgreSQL by re-embedding all conversations and memories. This
requires calling the embedding API for each record, which may take
time and incur API costs depending on your data volume.

### Configuration backup

Keep copies of your configuration files, Nginx settings, and
environment variables in a secure location. Version control works well
for this, excluding sensitive values like passwords and API keys.

## Monitor the system

Monitoring helps you detect and respond to issues before they affect
users.

### Health checks

The backend provides a health endpoint that reports the status of
dependencies.

```bash
curl http://localhost:8000/health
```

Configure your load balancer or orchestration system to use this
endpoint for health checks.

### Logging

The backend outputs structured JSON logs via logrus. Each request log
includes the HTTP method, path, query parameters, status code, client
IP, and response latency. Log levels are assigned based on status
codes: error for 5xx, warning for 4xx, and info for successful
requests.

Configure centralized logging to aggregate logs from all services.
Rotate log files regularly to prevent disk exhaustion.

### LangSmith monitoring

If you use LangSmith for LLM observability, monitor the LangSmith
dashboard for LLM call latency, error rates, and token usage. Enable
tracing by setting `LANGSMITH_TRACING=true` in your `.env` file.

### Metrics

Consider collecting metrics using Prometheus and visualizing them with
Grafana. Key metrics to track include request latency, error rates,
database connection pool usage, and storage capacity.

## Scale horizontally

As usage grows, you can scale the system horizontally.

### Backend scaling

Run multiple backend instances behind Nginx. The backend is stateless
(sessions are stored in Redis), so you can add instances without
configuration changes.

```yaml
# docker-compose.yml
services:
  backend:
    deploy:
      replicas: 3
```

Update the Nginx configuration to load balance across instances.

### Database considerations

PostgreSQL can handle significant load with proper indexing. The
migration files create indexes on frequently queried columns
(`user_id`, `status`, `due_date`, `recorded_at`, `category`). For very
high traffic, consider read replicas to distribute query load.

Milvus supports clustering for increased throughput. For production
workloads with large vector collections, consider deploying Milvus in
distributed mode with separate query nodes and data nodes. Monitor
collection sizes and search latency to determine when scaling is
needed.

Typesense can be clustered for high availability. Each node maintains a
copy of the data.

## Update the system

Regular updates patch security vulnerabilities and add features.

### Backend updates

Pull the latest code and rebuild.

```bash
cd backend
git pull
go build -o mino-server ./cmd/server
systemctl restart mino
```

Test updates in a staging environment first when possible. The backend
runs database migrations automatically on startup, so schema changes
are applied when the new version starts.

### Web app updates

Rebuild the web application for production.

```bash
cd web
npm install
npm run build
```

The output can be served by Nginx or any static hosting service.

## Incident response

Prepare for potential incidents with documented procedures.

### Service unavailable

If the backend becomes unavailable, check the service status first.

```bash
systemctl status mino
journalctl -u mino -n 50
```

Common issues include out-of-memory conditions, database connection
failures, and configuration errors.

### Data recovery

If you discover data loss, stop the service immediately to prevent
further damage. Restore from the most recent PostgreSQL backup and
investigate the cause. Typesense data can be rebuilt using the reindex
endpoint after PostgreSQL is restored. Milvus vector data can be
restored from backup volumes, or regenerated by re-embedding
conversations and memories from PostgreSQL.

```bash
curl -X POST http://localhost:8000/v1/search/reindex \
  -H "Authorization: Bearer <admin_token>"
```

### Security incident

If you suspect a security breach, rotate all credentials immediately.
This includes database passwords, API keys, JWT keys, and the admin
password. Review the structured JSON logs to identify unauthorized
activity.

## Performance tuning

Several adjustments improve performance in production.

### Database optimization

Ensure indexes exist on frequently queried columns. Verify them
periodically.

```sql
SELECT indexname, indexdef
FROM pg_indexes
WHERE tablename LIKE 'mino_%';
```

### Redis configuration

Configure Redis to persist data and use adequate memory. Monitor memory
usage and adjust the `maxmemory` setting as needed. The rate limiter
keys expire automatically, so Redis memory usage stays bounded.

### Audio processing

If transcription latency is high, consider using a faster transcription
provider or running a local Whisper instance. The transcription service
supports multiple providers, configurable through the `LLM_PROVIDER`
environment variable.

## Client deployment

### Web application

Build the web application for production.

```bash
cd web
npm run build
```

Configure Nginx to serve the built application and proxy API requests
to the backend. The web app's `next.config.js` includes a rewrite rule
that proxies `/api/v1/*` to the backend, so ensure the `API_URL`
environment variable points to your production backend.

### Watch application

Build the Flutter watch app for release.

```bash
cd watch
flutter build apk --release
```

Distribute the APK through your preferred channel. The watch app
connects to the backend URL configured in
`lib/core/constants/app_config.dart`. Update this value to point to
your production server before building.

### Mobile and desktop apps

For mobile apps, build release versions and distribute through your
preferred channel (App Store, Google Play, or direct download).

Desktop applications can be distributed through your website or app
stores. Build using the platform-specific tooling for each target OS.
