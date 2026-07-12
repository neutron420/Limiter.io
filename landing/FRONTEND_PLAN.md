# Limiter.io — Frontend Build Plan

What the frontend needs to become, mapped **1:1 to the real Go backend** (`internal/delivery/http/router.go`).
Everything currently on the dashboard is **mock data** — this doc is the checklist to wire it to the real API,
page by page, in the existing **brutalist / neobrutalism** design language.

> Stack already in place: Next.js 16 (App Router) · React 19 · Tailwind **v4** · shadcn + neobrutalism UI ·
> Framer Motion · TanStack Table · lucide-react. Fonts: JetBrains Mono + Geist Pixel.

---

## 0. Current state (what exists today)

| Route | File | State |
|---|---|---|
| `/` | `app/page.tsx` | ✅ Landing (hero, bento, pricing, footer) — static, fine |
| `/login` | `app/login/page.tsx` | ⚠️ UI only — not wired to `/auth/login` |
| `/register` | `app/register/page.tsx` | ⚠️ UI only — not wired to `/auth/register` |
| `/dashboard` | `app/dashboard/page.tsx` | ⚠️ **Mock logs + fake stats**, no project concept |
| dashboard sidebar | `app/dashboard/sidebar.tsx` | ✅ Static nav — needs real projects + routing |

**Gap:** no API client, no auth/token storage, no project selection, no real data anywhere.

---

## 1. Backend contract (source of truth)

- **Base URL:** `http://localhost:8080/api/v1` → put in `landing/.env.local` as `NEXT_PUBLIC_API_URL`.
- **Auth:** JWT. `login`/`register` return `{ access_token, refresh_token, email, user_id }`.
  Send `Authorization: Bearer <access_token>` on every private call. On 401 → call `/auth/refresh`
  with `{ refresh_token }`, retry once, else redirect to `/login`.
- **Everything is scoped to a `projectId`** (UUID) — the dashboard MUST have a selected project before it
  can show keys / rules / analytics.

### Endpoint → screen map

| Method + Path | Purpose | Frontend screen |
|---|---|---|
| `POST /auth/register` | Sign up | `/register` |
| `POST /auth/login` | Sign in | `/login` |
| `POST /auth/refresh` | Rotate access token | API client (silent) |
| `POST /auth/forgot-password` | Reset request | `/forgot-password` (new) |
| `POST /auth/logout` | End session | sidebar user menu |
| `POST /auth/change-password` | Change pw | `/dashboard/settings` |
| `GET /projects` · `POST /projects` | List / create projects | Project switcher + `/dashboard/projects` |
| `GET /projects/:id` · `DELETE /projects/:id` | Detail / delete | `/dashboard/projects` |
| `POST/GET /projects/:id/keys` | Create / list API keys | `/dashboard/keys` |
| `POST .../keys/:keyId/rotate` · `/revoke` · `DELETE` | Manage key | `/dashboard/keys` (row actions) |
| `POST/GET/GET/PUT/DELETE /projects/:id/rules` | CRUD rate-limit rules | `/dashboard/policies` |
| `GET /projects/:id/analytics/stats?duration=24h` | Aggregate stats | `/dashboard` (stat cards + charts) |
| `GET /projects/:id/analytics/logs?limit=&offset=` | Historical logs | `/dashboard` (table, paginated) |
| `GET /projects/:id/ws` (WebSocket) | Live request stream | `/dashboard` (live feed) |
| `GET /subscription` · `POST /subscription/upgrade` | Plan / upgrade | `/dashboard/billing` |
| `POST /gateway/*path` (API-key auth) | Test the limiter | `/dashboard/playground` (new, optional) |

---

## 2. Shared infrastructure to build FIRST (blocks everything else)

Create these before touching pages:

1. **`lib/api.ts`** — fetch wrapper: base URL, injects `Bearer` token, auto-refresh on 401, typed helpers
   (`api.get/post/put/del`). Throws `ApiError { status, message }`.
2. **`lib/types.ts`** — TypeScript mirrors of the Go DTOs (see §4). Single source for all pages.
3. **`lib/auth.tsx`** — `AuthProvider` + `useAuth()`: stores tokens (httpOnly cookie preferred; localStorage
   acceptable for v1), exposes `user`, `login()`, `register()`, `logout()`. Guards `/dashboard/*`.
4. **`lib/project-context.tsx`** — `useProject()`: current `projectId`, project list, setter. The sidebar
   switcher writes here; every dashboard page reads it. Persist selection in a cookie.
5. **`hooks/use-analytics-ws.ts`** — opens `GET /projects/:id/ws`, parses each JSON message into an
   `AnalyticsLog`, returns a rolling buffer + connection status. Reconnect with backoff.

---

## 3. Page-by-page spec

### `/login` and `/register`  (wire existing UI)
- On submit → `POST /auth/login` | `/auth/register`. Store tokens via `useAuth`, redirect to `/dashboard`.
- Show inline field errors (email format, password `min=8` for register). Show API error banner.
- Add "Forgot password?" link → `/forgot-password`.

### `/forgot-password`  (new, small)
- Single email field → `POST /auth/forgot-password`. Always show neutral "if the email exists…" success.

### `/dashboard`  (Overview — the main screen, replace all mocks)
- **Guard:** redirect to `/login` if unauthenticated.
- **Requires a selected project.** If user has 0 projects → empty state with "Create your first project" CTA.
- **Stat cards** from `GET /analytics/stats?duration=<range>`: `total_requests`, `allowed_requests`,
  `blocked_requests`, `avg_latency_ms`. Add a duration switcher (`24h` / `7d` / `30d`).
- **Top blocked routes** (`stats.top_blocked[]`) → small brutalist bar list.
- **Live feed**: `use-analytics-ws` prepends rows in real time; show a green/red "● LIVE" indicator.
- **Historical table**: `GET /analytics/logs` with pagination (limit/offset) — reuse `components/data-table.tsx`.
  Map fields: `decision`→status badge, `route`→path, `latency_ms`→latency, plus `client_ip`, `status_code`,
  `blocked_reason`, `timestamp`.
- Keep the current card styling (2px border, hard offset shadow, zero radius).

### `/dashboard/projects`  (new)
- Grid/list from `GET /projects`. "New Project" dialog → `POST /projects` (`name` 3–100, `description` ≤500).
- Each card: open (sets current project + go to overview), delete (`DELETE /projects/:id`, confirm dialog).
- Wire the sidebar workspace switcher to this list (replace hard-coded `teams`).

### `/dashboard/keys`  (new)
- Table from `GET /projects/:id/keys`: name, `prefix`, `last_used_at`, `expires_at`, revoked state.
- "Create key" → `POST .../keys` → **show the full `key` once** in a copy-once modal (backend only returns
  `plain_key` on creation). Row actions: **Rotate**, **Revoke**, **Delete** (each a confirm dialog).

### `/dashboard/policies`  (new — the core feature)
- Table from `GET /projects/:id/rules`: name, `route_pattern`, `algorithm`, `limit / period`, `burst`, active toggle.
- Create/Edit form (`POST` / `PUT`): fields per DTO — `algorithm` is a select of the 5 values
  (`token_bucket`, `fixed_window`, `sliding_window_counter`, `sliding_window_log`, `leaky_bucket`),
  `limit`>0, `period`(sec)>0, `burst`≥0. Active toggle → `PUT { is_active }`. Delete → `DELETE`.
- Add short helper text per algorithm (from `PROJECT_CONTEXT.md` §Redis).

### `/dashboard/billing`  (new)
- Current plan from `GET /subscription` (`plan_id`, `status`, `starts_at`, `expires_at`).
- Plan cards (free / pro / enterprise). Upgrade CTA → Lemon Squeezy checkout (external) **or**
  `POST /subscription/upgrade { plan_id }` for the manual path. Reuse `components/pricing-section.tsx` styling.

### `/dashboard/settings`  (new)
- Change password → `POST /auth/change-password` (`old_password`, `new_password` min 8).
- Account info (email from `useAuth`). Logout button → `POST /auth/logout` + clear tokens.

### `/dashboard/playground`  (new, optional but great demo)
- Enter an API key + path, fire `POST /gateway/<path>` with `X-API-Key`, show allow/block + rate-limit
  headers. Lets a dev watch their own rule trigger live in the Overview feed.

---

## 4. Data contracts (mirror in `lib/types.ts`)

```ts
type AuthResponse = { access_token: string; refresh_token: string; email: string; user_id: string }
type Project = { id: string; user_id: string; name: string; description: string; created_at: string; updated_at: string }
type ApiKey = { id: string; project_id: string; name: string; prefix: string; key?: string; // only on create
               expires_at?: string; revoked_at?: string; last_used_at?: string; created_at: string }
type Rule = { id: string; project_id: string; name: string; route_pattern: string;
              algorithm: 'token_bucket'|'fixed_window'|'sliding_window_counter'|'sliding_window_log'|'leaky_bucket';
              limit: number; period: number; burst: number; is_active: boolean; created_at: string; updated_at: string }
type Subscription = { user_id: string; plan_id: 'free'|'pro'|'enterprise'; status: string; starts_at: string; expires_at?: string }
type Stats = { total_requests: number; allowed_requests: number; blocked_requests: number;
               avg_latency_ms: number; top_blocked: { route: string; count: number }[] }
type AnalyticsLog = { id: string; project_id: string; api_key_id: string; request_id: string;
                      client_ip: string; route: string; status_code: number; latency_ms: number;
                      decision: 'allowed'|'blocked'; blocked_reason?: string; timestamp: string }
```

---

## 5. Recommended build order

1. **Infra** (§2) — `api.ts`, `types.ts`, `auth.tsx`, `project-context.tsx`, `use-analytics-ws.ts`.
2. **Auth wiring** — `/login`, `/register`, route guard on `/dashboard/*`.
3. **Projects** — switcher + `/dashboard/projects` (nothing else works without a selected project).
4. **Overview** — replace dashboard mocks with real stats + logs + WS live feed.
5. **Policies** → **Keys** → **Billing** → **Settings**.
6. **Playground** (optional polish).

Keep every new screen in the current brutalist system: 2px `border-foreground`, hard offset box-shadows,
`rounded-none`, uppercase mono labels, orange (`#ea580c` / `main`) accents.
