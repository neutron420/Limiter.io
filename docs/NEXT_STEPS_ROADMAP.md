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
- Project-based organization.
- Analytics, request logs, time-series data, and live WebSocket events.
- JWT authentication, refresh tokens, logout, password reset, OTP protection, and Turnstile support.
- Subscription plans and usage limits.
- Team invitations with email delivery.
- Owner, admin, and read-only member roles.
- Dedicated read-only member workspace at \`/member-dashboard\`.
- Admin/owner workspace at \`/dashboard\`.
- Docker deployment files, Swagger documentation, health checks, and Prometheus metrics.

## Recommended order

### 1. Finish production hardening

Highest priority before adding many new features.

- Add integration tests using PostgreSQL, Redis, and Kafka containers.
- Add end-to-end tests for:
  - Register/login
  - Create project
  - Create key and send a gateway request
  - Invite, accept, and remove a member
  - Read-only member access
  - Admin permissions
- Add database migrations instead of relying only on auto-migration.
- Add request timeouts and graceful shutdown to every service.
- Add retry and dead-letter handling for Kafka events.
- Add structured audit logs for security-sensitive actions.
- Add rate limits to login, invite, password reset, and public auth endpoints.
- Add backups and restore documentation for PostgreSQL.

### 2. Improve team collaboration

- Add a team activity timeline:
  - Member invited
  - Invite accepted
  - Key created or revoked
  - Rule changed
  - Project deleted
- Allow owners to change a member role.
- Add invite resend.
- Add invite expiration cleanup.
- Add project-level notification preferences.
- Add a member profile page with joined projects and activity.
- Add organization/workspace support above individual projects.
- Add groups such as Engineering, Support, and Operations.
- Add approval workflows for sensitive actions such as key rotation.

### 3. Make rate limiting more powerful

- Add custom response headers:
  - \`RateLimit-Limit\`
  - \`RateLimit-Remaining\`
  - \`RateLimit-Reset\`
- Add configurable 429 response bodies.
- Support quotas:
  - Per minute
  - Per hour
  - Daily
  - Monthly
- Support multiple rules per route with priority ordering.
- Add rule previews before publishing.
- Add dry-run mode that reports what would be blocked without blocking.
- Add maintenance mode and emergency global blocking.
- Add allowlists and denylists for IPs, CIDRs, countries, and headers.
- Add weighted requests where expensive endpoints consume more quota.
- Add per-customer or per-tenant identifiers.
- Add rule versioning and rollback.

### 4. Build better analytics

- Add a dedicated analytics page with charts for:
  - Requests over time
  - Allowed versus blocked
  - P95 and P99 latency
  - Top routes
  - Top API keys
  - Top clients
- Add date range filters and timezone selection.
- Add CSV and JSON exports.
- Add saved analytics views.
- Add anomaly detection for traffic spikes.
- Add alerts for:
  - Sudden block-rate increases
  - High latency
  - Traffic spikes
  - Key abuse
- Add email, Slack, and webhook alert channels.

### 5. Improve the developer experience

- Add SDKs for:
  - JavaScript/TypeScript
  - Python
  - Java
  - Ruby
- Add an official OpenAPI client.
- Add copy-ready integration snippets in the dashboard.
- Add a local development CLI:
  - Create project
  - Create key
  - Push rules
  - Test a route
- Add a public API status page.
- Add a request inspector showing the exact matched rule and decision reason.
- Add a sandbox project for testing without production traffic.
- Add API key scopes such as read-only, gateway-only, and admin.

### 6. Improve security

- Add optional MFA using TOTP or passkeys.
- Add session/device management.
- Add API key hashing with key metadata and last-used location.
- Add IP allowlisting for admin actions.
- Add CSRF protection where cookie authentication is used.
- Add security headers and a strict Content Security Policy.
- Add secret scanning in CI.
- Add dependency and container vulnerability scanning.
- Add automatic key rotation reminders.
- Add immutable audit-log storage for enterprise customers.

### 7. Add billing and enterprise features

- Add plan upgrade and downgrade flows.
- Add usage-based billing.
- Add invoices and payment history.
- Add plan enforcement for retention, projects, keys, and requests.
- Add enterprise SSO with SAML or OIDC.
- Add SCIM user provisioning.
- Add custom retention policies.
- Add dedicated regions and data residency options.
- Add SLA and support-contact settings.
- Add white-label email templates.

### 8. Improve operations and scale

- Add OpenTelemetry traces across API, Redis, Kafka, and PostgreSQL.
- Add Grafana dashboards and alert rules.
- Add Redis cluster and Kafka partitioning guidance.
- Add horizontal autoscaling based on request rate and Kafka lag.
- Add regional gateway deployment.
- Add edge rate limiting close to users.
- Add circuit breakers for PostgreSQL and Kafka.
- Add load tests for:
  - Single project
  - Many projects
  - High-cardinality API keys
  - Kafka lag
  - Redis failure
- Add chaos tests for Redis, Kafka, and database outages.

## Strong next milestone

The best next milestone is a **production-ready team and observability release**:

1. Add integration and end-to-end tests.
2. Add audit logs.
3. Add role management and invite resend.
4. Add analytics charts and CSV export.
5. Add Slack/webhook alerts.
6. Add MFA and session management.
7. Add OpenTelemetry and Grafana dashboards.
8. Add database migrations and backup verification.

This gives Limiter.io a strong foundation for real teams before expanding into advanced enterprise billing and multi-region infrastructure.

## Definition of done for a production release

- All critical flows have automated tests.
- A failed Redis or Kafka dependency does not silently lose important data.
- Every admin action is auditable.
- Members cannot mutate project configuration.
- API keys are never returned after creation.
- Database backups have been restored successfully in a test environment.
- Metrics, logs, traces, and alerts are available.
- Deployment and rollback steps are documented.
- Security scanning passes in CI.
- Load testing has been completed against the expected traffic target.
