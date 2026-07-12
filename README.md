# Distributed API Rate Limiting Platform (Backend)

A production-grade, highly scalable distributed API Rate Limiting Platform built in Go. Inspired by Cloudflare Rate Limiting, AWS API Gateway throttling, and Upstash Ratelimit, the platform enforces fine-grained API throttling at sub-millisecond speeds.

## Key Architecture & Features

- **Clean Architecture**: Strictly isolates domain logic from HTTP routers (Gin), database access (GORM), caching (Redis), and event streaming (Kafka).
- **Sub-Millisecond Decisions**: Authentication state (API keys) and project subscriptions are cached in Redis. The middleware makes decisions without hitting PostgreSQL.
- **Pluggable Rate Limiting Engine**: Implement multiple rate limit algorithms preloaded as atomic Lua scripts in Redis:
  - **Token Bucket** (Default, available on Free plan)
  - **Fixed Window** (Pro/Enterprise)
  - **Sliding Window Counter** (Pro/Enterprise)
  - **Sliding Window Log** (Pro/Enterprise)
  - **Leaky Bucket** (Pro/Enterprise)
- **Non-Blocking Kafka Pipeline**: Requests are allowed or blocked immediately. Logging and usage tracking events are dispatched to a buffered Go channel post-response, written to Kafka topic `api_logs`, and asynchronously processed in batches by consumer workers to persist into PostgreSQL.
- **Observability**: Exposes structured JSON logging via Uber's `zap` and records Prometheus metrics for API throughput, decisions, and system latency.
- **Deployment Ready**: Fully containerized with Docker, Docker Compose, and Kubernetes manifests (HPA, StatefulSets, Ingress, ConfigMaps, and Secrets).

---

## Technical Stack

- **Language**: Go (v1.21+)
- **HTTP Engine**: Gin
- **DBMS**: PostgreSQL
- **ORM**: GORM
- **Cache / Lock**: Redis
- **Messaging**: Apache Kafka (KRaft mode)
- **Log / Config**: Zap (JSON) & Viper (.env)
- **Metrics**: Prometheus
- **Container / Deployment**: Docker & Kubernetes

---

## Directory Layout

```text
├── cmd/
│   ├── api/            # API Server entrypoint
│   └── consumer/       # Kafka background aggregator consumer
├── internal/
│   ├── config/         # Environment configurations (Viper)
│   ├── database/       # Postgres connections and seeding
│   ├── delivery/
│   │   └── http/       # Gin routing initialization
│   ├── dto/            # Data Transfer Objects (Requests/Responses)
│   ├── handlers/       # Gin HTTP controllers
│   ├── kafka/          # Kafka producer and consumer wrappers
│   ├── middleware/     # Auth, Recovery, Log, Metrics & Rate Limiting middleware
│   ├── models/         # GORM database schemas
│   ├── ratelimiter/    # Rate Limiting interface and Redis implementations
│   ├── redis/          # Redis connection pool and preloaded Lua scripts
│   ├── repository/     # Interfaces and GORM/Redis concrete implementations
│   ├── services/       # Core business logic layer (Auth, Keys, Policies, etc.)
│   └── utils/          # Cryptographic hashing (Bcrypt, SHA-256) and JWT tokens
├── deploy/
│   ├── docker/         # Dockerfile & Docker-Compose (Kafka, Redis, Postgres, Prometheus)
│   └── kubernetes/     # Deployments, Services, ConfigMaps, Secrets, Ingress & HPAs
└── docs/               # API Reference and Architecture diagrams
```

---

## High-Throughput Request Flow

```text
 Client Request
      │
      ▼
[Request ID & Logs Middleware]
      │
      ▼
[API Key Auth Middleware] (Extracts & hashes key -> Redis check -> Postgres fallback -> Context injection)
      │
      ▼
[Rate Limit Middleware] (Fetches policy rules -> Executes atomic Redis Lua script)
      │
 ┌────┴────────────────────────┐
 │                             │
 ▼ (Allowed)                   ▼ (Blocked)
[c.Next() -> Route Handler]   [HTTP 429 Too Many Requests]
 │                             │
 └────┬────────────────────────┘
      │
      ▼
[Send immediate HTTP response to Client]
      │
      ▼ (Post-Response Asynchronous)
[Push event to Go Channel -> Write to Kafka -> Consumer -> Batch SQL write to PostgreSQL]
```

---

## Database ER Diagram

The database uses PostgreSQL for permanent business metadata:

```text
  ┌──────────────┐         ┌───────────────┐
  │    Users     │◄────────│ RefreshTokens │
  └──────┬───────┘         └───────────────┘
         │
         ├───(1:1)───► ┌─────────────────┐         ┌───────────┐
         │             │  Subscriptions  │◄────────│   Plans   │
         │             └─────────────────┘         └───────────┘
         ▼ (1:N)
  ┌──────────────┐
  │   Projects   │
  └──────┬───────┘
         │
         ├───(1:N)───► ┌─────────────────┐
         │             │     APIKeys     │
         │             └─────────────────┘
         ├───(1:N)───► ┌─────────────────┐
         │             │ RateLimitRules  │
         │             └─────────────────┘
         └───(1:N)───► ┌─────────────────┐
                       │  AnalyticsLogs  │
                       └─────────────────┘
```

---

## Local Development Guide

### Prerequisites
- Go 1.21+
- Docker & Docker Compose

### 1. Configure Environment
Copy the example environment template:
```bash
cp .env.example .env
```

### 2. Start Infrastructures using Docker Compose
The Docker Compose script builds the API, background worker, database, cache, and queueing services:
```bash
docker-compose -f deploy/docker/docker-compose.yml up --build -d
```

### 3. Verification & Live Scrape
- **API Server URL**: http://localhost:8080
- **Prometheus UI**: http://localhost:9090
- **Liveness probe**: `GET http://localhost:8080/healthz`
- **Readiness probe**: `GET http://localhost:8080/readyz` (Verifies DB, Redis, and Kafka connection states)

---

## Environment Variables Guide

| Variable | Default Value | Description |
| :--- | :--- | :--- |
| `ENV` | `development` | Environment mode (`development` or `production`) |
| `PORT` | `8080` | Port for the API server |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | PostgreSQL username |
| `DB_PASSWORD`| `postgres` | PostgreSQL password |
| `DB_NAME` | `ratelimiter` | PostgreSQL database name |
| `DB_SSLMODE` | `disable` | PostgreSQL SSL encryption mode |
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `KAFKA_BROKERS`| `localhost:9092`| Comma-separated list of Kafka broker endpoints |
| `KAFKA_TOPIC`| `api_logs` | Kafka topic name for rate limit logs |
| `JWT_SECRET` | `super_secret...` | HMAC secret for signing JWT access tokens |
| `JWT_ACCESS_TTL`| `15m` | Lifetime of access tokens |
| `JWT_REFRESH_TTL`| `168h` | Lifetime of refresh tokens (7 days) |
| `ADMIN_EMAIL`| `admin@ratelimiter.io`| Seeded administrator email |
| `ADMIN_PASSWORD`| `AdminPass123!`| Seeded administrator password |

---

## API Reference

### 1. Authentication
- `POST /api/v1/auth/register` - Create user profile & attach default **Free Plan** subscription.
- `POST /api/v1/auth/login` - Authenticate, generate JWT access and secure refresh token.
- `POST /api/v1/auth/refresh` - Rotate refresh token and issue new access token.
- `POST /api/v1/auth/forgot-password` - Trigger reset instructions.
- `POST /api/v1/auth/logout` (Auth required) - Revoke user's refresh tokens.
- `POST /api/v1/auth/change-password` (Auth required) - Update password.

### 2. Project & API Key Control (Auth required)
- `POST /api/v1/projects` - Create project.
- `GET /api/v1/projects` - List all projects owned by user.
- `DELETE /api/v1/projects/:projectId` - Permanently delete project.
- `POST /api/v1/projects/:projectId/keys` - Create a new API Key (Secret key returned only *once*).
- `GET /api/v1/projects/:projectId/keys` - List all keys (Exposes prefix, hides hashes).
- `POST /api/v1/projects/:projectId/keys/:keyId/rotate` - Invalidate key, issue a new one.
- `POST /api/v1/projects/:projectId/keys/:keyId/revoke` - Instant revocation (Invalidates cache).

### 3. Policy Rules & Throttling Rules (Auth required)
- `POST /api/v1/projects/:projectId/rules` - Add matching route rate limit rule.
- `GET /api/v1/projects/:projectId/rules` - View active rules.
- `PUT /api/v1/projects/:projectId/rules/:ruleId` - Modify rate, algorithm, burst, or status.
- `DELETE /api/v1/projects/:projectId/rules/:ruleId` - Remove rule.

### 4. Subscription upgrades
- `GET /api/v1/subscription` - View active subscription, plan, and rules limitations.
- `POST /api/v1/subscription/upgrade` - Transition plans (e.g. `pro`). Checks limits and invalidates caches instantly.

### 5. Analytics logs
- `GET /api/v1/projects/:projectId/analytics/stats` - Fetch total, allowed, blocked request counts & average latencies for a period (e.g. `?duration=7d`).
- `GET /api/v1/projects/:projectId/analytics/logs` - Fetch paginated requests history logs.

---

## Subscription Plans Limitations

| Plan Property | Free Plan | Pro Plan | Enterprise Plan |
| :--- | :--- | :--- | :--- |
| **Max Projects** | 3 | Unlimited (`-1`) | Unlimited (`-1`) |
| **Max Keys per Project**| 3 | Unlimited (`-1`) | Unlimited (`-1`) |
| **Available Algorithms**| Token Bucket | Token, Fixed, Sliding Windows, Leaky Bucket | All algorithms + custom configurations |
| **Retention Window** | 7 Days | 90 Days | 365 Days |
| **Max Requests** | 100/minute | 10,000/minute | 1,000,000/minute |
