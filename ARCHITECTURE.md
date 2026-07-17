# Architecture & Data Flow

## System Architecture

```mermaid
graph TB
    subgraph Browser["Browser (Next.js 16 App Router)"]
        A1["Landing Pages
/login /register /pricing"]
        A2["Dashboard Pages
/analytics /keys /rules /billing
/quotas /tenants /orgs /audit"]
        A3["API Client (lib/api.ts)
Auto-refresh tokens on 401"]
    end

    subgraph CDN["Cloudflare CDN"]
        CF["Edge Cache
Turnstile CAPTCHA"]
    end

    subgraph Gateway["API Layer (Gin Framework)"]
        MW["Middleware Pipeline
Logger → RequestID → CORS
SecurityHeaders → JWT Auth
RateLimit → CircuitBreaker"]
        RT["Router
60+ Routes → 20 Handlers"]
    end

    subgraph Biz["Business Layer"]
        SV["15 Services
auth / project / apikey / policy
analytics / billing / subscription
org / sso / passkey / quota
tenant / notification / approval"]
    end

    subgraph Data["Data Layer"]
        PG[("PostgreSQL 16
Users / Projects / Keys
Rules / Analytics / Billing
Orgs / Tenants / Audit")]
        RC[("Redis 7
Cache (plans, keys, subs)
Rate Limit Counters")]
    end

    subgraph Events["Event Pipeline"]
        KP["Kafka Producer"]
        KC["Kafka Consumer
Batch insert every 2s / 100 records
DLQ on failure"]
        KT[("Kafka 3.7
analytics_logs topic")]
    end

    subgraph Realtime["Real-time"]
        WS["WebSocket Hub
gorilla/websocket
per-project pub/sub"]
    end

    subgraph External["External"]
        GOOG["Google OAuth"]
        MAIL["SMTP Mailer"]
        LS["LemonSqueezy
Subscription billing"]
        WEBH["Webhook Events"]
    end

    Browser --> CF
    CF --> Gateway
    Gateway --> MW --> RT
    RT --> SV
    SV --> Data
    SV --> KP --> KT --> KC --> PG
    SV --> WS
    WS --> Browser
    SV --> External
```

## User Journey Flow

```mermaid
flowchart LR
    REG["Register
POST /auth/register
email + password → bcrypt hash"] --> LOGIN
    LOGIN["Login
POST /auth/login
validate → JWT access + refresh token"] --> DASH
    DASH["Dashboard
/projects → create / select project"] --> PROJ
    PROJ["Project Overview
Analytics, Keys, Rules, Billing"] --> KEYS
    PROJ --> RULES
    PROJ --> BILLING
    PROJ --> ORGS

    KEYS["Create API Key
rk_live_xxx → SHA256 hash
store hash, return plain key once"] --> USE
    RULES["Create Rate Limit Rule
choose algorithm (token_bucket, etc.)
set limit, period, burst"] --> USE
    USE["Use API
request → header X-API-Key
→ middleware matches rule
→ Redis Lua atomic check
→ allow/block → Kafka log"]

    BILLING["Billing
free plan = 3 projects, 100 req/s
upgrade → LemonSqueezy checkout
webhook → update subscription"]

    ORGS["Organization
create org → invite members
groups, approval workflows
SSO (SAML/OIDC), audit logs"]
```

## Invite Flow

```mermaid
sequenceDiagram
    actor Owner as Org Owner
    participant API as API
    participant DB as PostgreSQL
    participant Mail as SMTP
    actor User as Invited User

    Owner->>API: POST /orgs/:id/invites {email, role}
    API->>DB: Create invite (pending, 7 day expiry)
    API->>Mail: Send invite email with token link
    Mail-->>User: "You've been invited to join Org"
    
    User->>API: GET /invites/:token
    API->>DB: Lookup token, check expiry
    API-->>User: Show org name, role, accept/decline
    
    alt User has account
        User->>API: POST /invites/:token/accept
        API->>DB: Add org_member, mark invite accepted
        API-->>User: Redirect to org dashboard
    else User needs register
        User->>API: POST /auth/register {email, password}
        API->>DB: Create user + subscription
        User->>API: POST /invites/:token/accept
        API->>DB: Add org_member
        API-->>User: Redirect to org dashboard
    end

    API->>WS: Broadcast member_joined to org
    API->>Mail: Notify owner "User accepted invite"
```

## Auth Security Flow

```mermaid
flowchart TD
    subgraph Login["Primary Login"]
        L1["Email + Password
→ bcrypt verify"]
        L2["Google OAuth
→ verify ID token → auto-register"]
        L3["SSO (SAML/OIDC)
→ org-level IdP login"]
    end

    subgraph MFA["MFA Check"]
        M1{"MFA Enabled?"}
        M1 -->|"Yes"| M2["Require TOTP code
or WebAuthn passkey"]
        M1 -->|"No"| TOKEN
    end

    subgraph Session["Session"]
        TOKEN["Issue JWT (15m TTL)
+ Opaque Refresh Token (7d)"]
        REFRESH["Refresh on 401
old token revoked → new issued
reuse detection → revoke all"]
    end

    Login --> MFA --> Session
```

## Rate Limit Decision Flow

```mermaid
sequenceDiagram
    participant Client
    participant MW as Rate Limit Middleware
    participant Matcher as Rule Matcher
    participant Redis as Redis (Lua)
    participant Kafka as Kafka Producer

    Client->>MW: Request (API key in header)
    MW->>Matcher: Match route pattern → find rule
    Matcher-->>MW: Rule found (algorithm, limit, period)
    
    alt No matching rule
        MW-->>Client: Allow (pass through)
    else Rule found
        MW->>Redis: EVALSHA (algorithm Lua script)
        Redis->>Redis: Atomic token/water/check
        Redis-->>MW: {allowed: 1/0, remaining: N}
        
        alt Allowed
            MW-->>Client: Allow (forward to upstream)
        else Blocked
            MW-->>Client: 429 Too Many Requests
            MW->>Client: X-RateLimit-Remaining: 0
            MW->>Client: Retry-After: X
        end
        
        MW->>Kafka: Log decision (async)
    end
```

## Tech Stack

| Component | Technology |
|---|---|
| Frontend | Next.js 16, TypeScript, Tailwind, shadcn, Framer Motion |
| Backend | Go 1.25, Gin, GORM, gorilla/websocket |
| Database | PostgreSQL 16 |
| Cache | Redis 7 (rate limit Lua scripts) |
| Queue | Kafka 3.7 + DLQ |
| Auth | JWT + bcrypt + TOTP + WebAuthn |
| Monitoring | Prometheus + OpenTelemetry |
| Deployment | Docker Compose / Kubernetes |
| CDN | Cloudflare (Turnstile, caching) |
