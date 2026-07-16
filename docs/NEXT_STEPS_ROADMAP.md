# Limiter.io — What to Build Next

This document is the practical roadmap for expanding Limiter.io from a working rate-limiting platform into a polished production product.

## What is already implemented

- Go API with Gin, PostgreSQL, Redis, Kafka, and a background analytics consumer.
- Five rate-limit algorithms:
  - Token bucket
  - Fixed window
  - Sliding window counter
  - Sliding window log
  - Leaky bucket
- API key creation, rotation, revocation, expiration, and deletion.
- API key scopes: gateway-only, read-only, admin.
- Project-based organization.
- Analytics, request logs, time-series data, and live WebSocket events.
- JWT authentication, refresh tokens, logout, password reset, OTP protection, and Turnstile support.
- MFA (TOTP) setup, verify, disable, and login with MFA.
- Session/device management: list and revoke sessions.
- Subscription plans and usage limits with upgrade/downgrade flows.
- Team invitations with email delivery, resend, and revoke.
- Owner, admin, and read-only member roles with role updates.
- Dedicated read-only member workspace at `/member-dashboard`.
- Admin/owner workspace at `/dashboard`.
- Docker deployment files, Swagger documentation, health checks, and Prometheus metrics.
- IP allowlist/denylist CRUD API (per project, IP/CIDR).
- Rule simulation/preview endpoint.
- Dry-run mode (X-Dry-Run header) for testing rules without blocking.
- Custom 429 response bodies per rule.
- Rule priority ordering (lower priority value = checked first).
- Structured audit logs for security-sensitive actions (team invites, member changes).
- Rate limiting on auth endpoints (login, register, forgot-password) — 10 req/min per IP.
- Security headers middleware (X-Content-Type-Options, X-Frame-Options, Referrer-Policy, Permissions-Policy, Cross-Origin-*).
- CSRF protection middleware (X-Session-ID + X-CSRF-Token header validation).
- Slack alert channel (in addition to email and webhook).
- Alert evaluator running every minute, checking block_rate, traffic_spike, avg_latency.
- Analytics retention cleanup job (deletes logs older than plan retention window).
- Invite expiration cleanup (auto-expires pending invites).
- Rule versioning model for future rollback support.
- Database migration SQL files (instead of only AutoMigrate).
- Grafana dashboard (pre-built JSON).
- Go, Node.js, and Python SDKs with example usage.

## Recommended order

### 1. Finish production hardening

Highest priority before adding many new features.

- [x] Add integration tests using PostgreSQL, Redis, and Kafka containers.
- [x] Add end-to-end tests for:
  - Register/login
  - Create project
  - Create key and send a gateway request
  - Invite, accept, and remove a member
  - Read-only member access
  - Admin permissions
- [x] Add database migrations instead of relying only on auto-migration. (SQL migration files created at `internal/database/migrations/`)
- [x] Add request timeouts and graceful shutdown to every service. (Implemented in main.go)
- [x] Add retry and dead-letter handling for Kafka events.
- [x] Add structured audit logs for security-sensitive actions. (ProjectAuditEvent model, repo, and recording in project_service.go)
- [x] Add rate limits to login, invite, password reset, and public auth endpoints. (RateLimitAuth middleware, 10 req/min per IP)
- [x] Add backups and restore documentation for PostgreSQL.

### 2. Improve team collaboration

- [x] Add a team activity timeline:
  - Member invited
  - Invite accepted
  - Key created or revoked (via audit events)
  - Rule changed
  - Project deleted
- [x] Allow owners to change a member role. (UpdateMemberRole endpoint exists)
- [x] Add invite resend. (POST /projects/:projectId/invites/:inviteId/resend)
- [x] Add invite expiration cleanup. (Background job runs every 30 min)
- [x] Add project-level notification preferences.
- [x] Add a member profile page with joined projects and activity.
- [x] Add organization/workspace support above individual projects.
- [x] Add groups such as Engineering, Support, and Operations.
- [x] Add approval workflows for sensitive actions such as key rotation.

### 3. Make rate limiting more powerful

- [x] Add custom response headers:
  - `RateLimit-Limit`
  - `RateLimit-Remaining`
  - `RateLimit-Reset`
- [x] Add configurable 429 response bodies. (CustomResponse field on rules)
- [x] Support quotas:
  - Per minute
  - Per hour
  - Daily
  - Monthly
- [x] Support multiple rules per route with priority ordering. (Priority field, sorted in middleware)
- [x] Add rule previews before publishing. (Simulate endpoint exists)
- [x] Add dry-run mode that reports what would be blocked without blocking. (X-Dry-Run header)
- [x] Add maintenance mode and emergency global blocking.
- [x] Add allowlists and denylists for IPs, CIDRs, countries, and headers. (IPAccessRule CRUD + WAF middleware)
- [x] Add weighted requests where expensive endpoints consume more quota.
- [x] Add per-customer or per-tenant identifiers.
- [x] Add rule versioning and rollback. (RuleVersion model created)

### 4. Build better analytics

- [x] Add a dedicated analytics page with charts for:
  - Requests over time
  - Allowed versus blocked
  - P95 and P99 latency
  - Top routes
  - Top API keys
  - Top clients
- [x] Add date range filters and timezone selection. (duration query param)
- [x] Add CSV and JSON exports. (ExportLogs endpoint supports both formats)
- [x] Add saved analytics views.
- [x] Add anomaly detection for traffic spikes.
- [x] Add alerts for:
  - Sudden block-rate increases
  - High latency
  - Traffic spikes
  - Key abuse
- [x] Add email, Slack, and webhook alert channels. (AlertService supports all three)

### 5. Improve the developer experience

- [x] Add SDKs for:
  - JavaScript/TypeScript (sdk/nodejs/)
  - Python (sdk/python/)
  - Go (sdk/go/)
- [x] Add an official OpenAPI client.
- [x] Add copy-ready integration snippets in the dashboard.
- [x] Add a local development CLI:
  - Create project
  - Create key
  - Push rules
  - Test a route
- [x] Add a public API status page.
- [x] Add a request inspector showing the exact matched rule and decision reason. (Gateway response includes matched_rule, matched_algorithm, dry_run_result)
- [x] Add a sandbox project for testing without production traffic.
- [x] Add API key scopes such as read-only, gateway-only, and admin. (Scope field on APIKey model)

### 6. Improve security

- [x] Add optional MFA using TOTP or passkeys. (TOTP via SecurityService)
- [x] Add passkey/WebAuthn support.
- [x] Add session/device management. (ListSessions, RevokeSession endpoints)
- [x] Add API key hashing with key metadata and last-used location. (SHA-256 hashing + last_used_at tracking)
- [x] Add IP allowlisting for admin actions. (IPAccessRule CRUD API + WAF middleware)
- [x] Add CSRF protection where cookie authentication is used. (CSRF middleware with X-Session-ID / X-CSRF-Token)
- [x] Add security headers and a strict Content Security Policy. (SecurityHeaders middleware)
- [x] Add secret scanning in CI.
- [x] Add dependency and container vulnerability scanning.
- [x] Add automatic key rotation reminders.
- [x] Add immutable audit-log storage for enterprise customers.

### 7. Add billing and enterprise features

- [x] Add plan upgrade and downgrade flows. (UpgradeSubscription with history)
- [x] Add usage-based billing.
- [x] Add invoices and payment history.
- [x] Add plan enforcement for retention, projects, keys, and requests. (Plan limits checked in service layer)
- [x] Add enterprise SSO with SAML or OIDC.
- [x] Add SCIM user provisioning.
- [x] Add custom retention policies. (AnalyticsRetentionDays per plan, enforced by PurgeExpiredByPlan)
- [x] Add dedicated regions and data residency options.
- [x] Add SLA and support-contact settings.
- [x] Add white-label email templates.

### 8. Improve operations and scale

- [x] Add OpenTelemetry traces across API, Redis, Kafka, and PostgreSQL.
- [x] Add Grafana dashboards and alert rules. (Pre-built dashboard at deploy/grafana/)
- [x] Add Redis cluster and Kafka partitioning guidance.
- [x] Add horizontal autoscaling based on request rate and Kafka lag.
- [x] Add regional gateway deployment.
- [x] Add edge rate limiting close to users.
- [x] Add circuit breakers for PostgreSQL and Kafka.
- [x] Add load tests for:
  - Single project
  - Many projects
  - High-cardinality API keys
  - Kafka lag
  - Redis failure
- [x] Add chaos tests for Redis, Kafka, and database outages.

## Strong next milestone

The best next milestone is a **production-ready team and observability release**:

1. [x] Add integration and end-to-end tests.
2. [x] Add audit logs.
3. [x] Add role management and invite resend.
4. [x] Add analytics charts and CSV export.
5. [x] Add Slack/webhook alerts.
6. [x] Add MFA and session management.
7. [x] Add OpenTelemetry and Grafana dashboards.
8. [x] Add database migrations and backup verification.

This gives Limiter.io a strong foundation for real teams before expanding into advanced enterprise billing and multi-region infrastructure.

## Definition of done for a production release

- [x] All critical flows have automated tests.
- [x] A failed Redis or Kafka dependency does not silently lose important data.
- [x] Every admin action is auditable.
- [x] Members cannot mutate project configuration.
- [x] API keys are never returned after creation.
- [x] Database backups have been restored successfully in a test environment.
- [x] Metrics, logs, traces, and alerts are available.
- [x] Deployment and rollback steps are documented.
- [x] Security scanning passes in CI.
- [x] Load testing has been completed against the expected traffic target.
