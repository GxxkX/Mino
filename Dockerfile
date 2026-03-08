# =============================================================================
# Mino All-in-One Dockerfile
# Builds both backend (Go) and web (Next.js) into a single image.
# The container runs both services via a lightweight process manager.
# =============================================================================

# ---- Backend build ----
FROM golang:1.24-alpine AS backend-builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src/backend

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/mino-server ./cmd/server

# ---- Web dependencies ----
FROM node:20-alpine AS web-deps
WORKDIR /app
COPY web/package.json web/package-lock.json ./
RUN npm ci

# ---- Web build ----
FROM node:20-alpine AS web-builder
WORKDIR /app
COPY --from=web-deps /app/node_modules ./node_modules
COPY web/ .

ARG API_URL=http://localhost:8000
ENV API_URL=${API_URL}

RUN npm run build

# ---- Runtime ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata nodejs supervisor
ENV TZ=Asia/Shanghai

WORKDIR /app

# Backend
COPY --from=backend-builder /bin/mino-server ./backend/mino-server
COPY backend/migrations/ ./backend/migrations/
COPY backend/keys/ ./backend/keys/

# Web
COPY --from=web-builder /app/public ./web/public
COPY --from=web-builder /app/.next/standalone ./web/
COPY --from=web-builder /app/.next/static ./web/.next/static

# Supervisor config to run both processes
RUN mkdir -p /etc/supervisor.d
COPY <<'EOF' /etc/supervisor.d/mino.ini
[supervisord]
nodaemon=true
logfile=/dev/stdout
logfile_maxbytes=0

[program:backend]
command=/app/backend/mino-server
directory=/app/backend
autostart=true
autorestart=true
stdout_logfile=/dev/stdout
stdout_logfile_maxbytes=0
stderr_logfile=/dev/stderr
stderr_logfile_maxbytes=0

[program:web]
command=node /app/web/server.js
directory=/app/web
environment=NODE_ENV="production",PORT="3000",HOSTNAME="0.0.0.0"
autostart=true
autorestart=true
stdout_logfile=/dev/stdout
stdout_logfile_maxbytes=0
stderr_logfile=/dev/stderr
stderr_logfile_maxbytes=0
EOF

EXPOSE 8000 3000

CMD ["supervisord", "-c", "/etc/supervisor.d/mino.ini"]
