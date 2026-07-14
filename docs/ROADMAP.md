# Limiter.io — Project Status & Final Roadmap

**This is the canonical, up-to-date roadmap** (supersedes `PROJECT_ROADMAP.md`).
Written after a full pass over the codebase — `internal/`, `cmd/`, `sdk/`, `deploy/`, `landing/`.

---

## ✅ What's already built (the project is ~90% feature-complete)

### Backend (Go · Gin · GORM · Redis · Kafka · Postgres)
- **Auth**: register, login, JWT access + refresh (with rotation & reuse detection), logout (revokes tokens),
  change-password, forgot-password **+ reset-password token flow**.
- **Projects / API keys**: full CRUD; keys create/rotate/revoke/delete, SHA-256 hashed, prefix shown.
- **Rate policies**: 5 algorithms (token bucket, fixed/sliding window, sliding log, leaky bucket), atomic Redis
  Lua eval, **per-key / per-IP / per-header key strategies** (`ratelimit.go`), plan-gated algorithms.
- **Rule dry-run simulator** (`POST /rules/:id/simulate`).
- **Analytics**: stats, logs (paginated), **time-series**, live WebSocket stream, **and direct-to-Postgres
  persistence** (works without the Kafka consumer running).
- **Billing**: subscription, Lemon Squeezy checkout + **webhook audit log**, **usage metering**
  (`GET /subscription/usage`).
- **Teams**: project members (`/projects/:id/members`).
- **Security**: WAF middleware (`waf.go`), JWT shield (`jwt_shield.go`), config-driven CORS.
- **Ops**: Prometheus metrics (`/metrics`), structured zap request logging, health/readiness probes,
  graceful shutdown, Swagger docs (`/swagger`).

### Frontend (Next.js 16 · React 19 · Tailwind v4 · neobrutalism)
- Landing + marketing pages: **platform, enterprise, resources, company**, pricing (Lemon Squeezy wired).
- Auth: login, register, forgot-password, reset-password.
- Dashboard: overview (stats + logs + live feed), **analytics (recharts)**, policies, keys,
  playground (**dry-run**), billing (usage + webhooks), settings (**members**).
- Env-driven config (no hardcoded URLs), 0 TypeScript errors, `next build` passes.

### SDKs & Infra
- SDKs: **Go, Node.js, Python** (fail-open middleware).
- Deploy: Docker, docker-compose, **full Kubernetes manifests** (api, consumer, kafka, redis, postgres),
  Prometheus config, **Grafana dashboard** (`deploy/grafana/limiter-dashboard.json`).
- **CI** pipeline (`.github/workflows/ci.yml`).

---

## 🎯 What's genuinely left

### 🔴 P0 — Correctness & trust before any real users
1. **Tests are the biggest gap.** Only `internal/middleware/ratelimit_test.go` exists.
   Add: unit tests for each Lua algorithm, JWT/crypto utils, `matchRoute`, CORS helpers; handler
   integration tests (httptest + sqlite/testcontainers); one end-to-end gateway flow. Wire into `ci.yml`.
2. **[DONE] Real email transport.** Integrated with Resend API using the configured `RESEND_API_KEY` and fallback mailer logger for local development. Full HTML OTP/link dispatch is verified.
3. **Analytics retention is defined but never enforced.** Plans carry `AnalyticsRetentionDays`
   (7/90/365) but no job deletes old `analytics_logs` → the table grows forever. Add a scheduled cleanup
   (cron goroutine or the consumer) that prunes per project's plan.

### 🟠 P1 — Security & product hardening
4. **Brute-force protection on auth endpoints.** `/auth/login`, `/register`, `/forgot-password` are
   unthrottled. Add IP-based rate limiting (reuse the limiter) + account lockout/backoff.
5. **Production migrations.** `AutoMigrate` is fine for dev but risky in prod — adopt `golang-migrate`
   (or Atlas) with versioned, reviewable migration files.
6. **Usage-based billing enforcement.** Metering exists (`GetUsage`), but nothing acts on it — add
   soft/hard quota enforcement + overage handling, and surface warnings in the dashboard.
7. **Email notifications.** Upgrade confirmations, approaching-limit warnings, key-created alerts
   (depends on #2).

### 🟡 P2 — DX & polish
8. **In-app API docs / quickstart page** (`/dashboard/docs`) — embed the Swagger link + a copy-paste
   SDK snippet pre-filled with the project's real key.
9. **Onboarding wizard** — first-run flow: create project → generate key → starter rule → fire a test
   request in the Playground. Turns the empty dashboard into an "aha" in 60 seconds.
10. **Landing interactive demo** — a "watch a request get throttled" widget on the landing page.
11. **Publish the SDKs** — npm (`@limiter/sdk`) and PyPI, with versioning + README badges.
12. **Swagger/OpenAPI completeness** — annotate the newer endpoints (usage, members, webhooks, simulate,
    reset-password) so `/swagger` is fully accurate.

### 🟢 P3 — Scale & operations
13. **Alerting & tracing** — Prometheus alert rules (blocked-ratio spike, Redis/Kafka down, error rate)
    and OpenTelemetry traces to complement the Grafana dashboard.
14. **Backups & DR** — automated Postgres backups, Redis persistence policy, documented restore.
15. **Horizontal scale notes** — Redis cluster/sharding, Kafka partitioning by project, stateless API
    replicas (k8s HPA).
16. **Multi-region edge** — the marketing site promises "global edge"; document/implement Redis
    replication + geo-routing to back it.
17. **API versioning & deprecation policy** — you're on `/api/v1`; define how v2 ships without breaking
    SDK users.

---

## Suggested order
**#1 tests → #2 email → #3 retention** (make it trustworthy) →
**#4 auth throttling → #5 migrations** (make it safe for prod) →
**#8 docs → #9 onboarding → #10 landing demo** (make it convert) →
then P3 as you scale.

> The core product is done and works end-to-end. Everything above is hardening, polish, and growth —
> not missing functionality.
