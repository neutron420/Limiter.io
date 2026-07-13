# Limiter.io — Roadmap & Next Steps

A prioritized, grounded list of what to build next — based on the actual code in this repo
(`internal/`, `sdk/`, `landing/`). Ordered by impact. Each item notes *why* and *where*.

---

## 🔴 P0 — Correctness & "make the demo actually work"

1. ✅ **Already implemented** — the `plans` table is seeded (free / pro / enterprise with real limits) in
   `database/postgres.go:77` (`MigrateAndSeed`).

2. ✅ **Already implemented** — `Register` creates a default `free` subscription row
   (`auth_service.go:72`), so `/subscription` and Billing don't 500.

3. ✅ **DONE (this pass)** — fixed the Lemon Squeezy variant check. `billing_handler.go` now uses
   `strconv.Itoa(variantID)` compared to `LEMON_SQUEEZY_PRO_VARIANT_ID` (default `1899978` in
   `config.go`); the buggy `string(rune(id))` and hardcoded magic number are gone.

4. **Wire real email for `forgot-password` / verification.** ← still open
   `POST /auth/forgot-password` is a stub (`auth_service.go:209` returns nil, no mail). Add SMTP/Resend/SES
   and a reset-token flow. The frontend page is already built (`/forgot-password`).

5. ✅ **Already implemented** — plan limits are enforced: `max_projects` (`project_service.go:47`),
   `max_keys_per_project` (`apikey_service.go:68`), and `allowed_algorithms` (`policy_service.go:57` on
   create, `:161` on update).

---

## 🟠 P1 — Product features (high user value)

6. ✅ **DONE (this pass)** — real server-side log pagination on the Overview. Prev/Next are wired to
   `limit`/`offset` (`app/dashboard/page.tsx`), reset on project switch. Still open: filters by
   decision / route / date range.

7. **Analytics charts.**
   `recharts` is already a dependency. Add a requests-over-time line chart and an allowed-vs-blocked
   donut on the Overview, fed by a new time-bucketed stats endpoint (extend `GetAggregatedStats`).

8. **Per-key & per-IP rate limiting granularity.**
   The limiter key is `project:apikey:rule` (`ratelimit.go:107`). Add an option to key by client IP or a
   custom identifier header, so one noisy client can't be limited as the whole key.

9. **Rule tester / dry-run.** Extend the Playground: pick a rule, simulate N requests, and show exactly
   when it would trip — without spending real quota.

10. **API docs page in-app.** Swagger is already served at `/swagger`. Link it from the dashboard and the
    SDK guide, and embed a copy-paste quickstart (`sdk/client.go`) per project with its real key.

11. **Webhook delivery log.** Show incoming Lemon Squeezy events + verification status in Billing so users
    can debug why an upgrade didn't land.

---

## 🟡 P2 — Hardening & operations

12. ✅ **DONE (this pass)** — CORS is now config-driven via `CORS_ALLOWED_ORIGINS` (`router.go`,
    `config.go`). Defaults to `*` for dev; set a comma-separated allowlist in production, and credentials
    are only sent for explicit origins (never with `*`).

13. **Rate-limit-headers on allowed gateway responses always.** They're set only when a rule matches
    (`ratelimit.go:122`). Consider returning them (or a `X-RateLimit-Policy: none`) even on pass-through so
    clients can introspect.

14. **Refresh-token rotation & revocation on logout.** Ensure `/auth/logout` revokes the refresh token
    server-side (the `RefreshToken.Revoked` field exists in `models.go:37`).

15. **Observability.** Prometheus metrics already exist (`/metrics`). Add a Grafana dashboard JSON to
    `deploy/`, and structured request logging with the existing `RequestID`.

16. **Tests.** Only `ratelimit_test.go` exists. Add unit tests for each algorithm's Lua script, auth/JWT,
    and handler integration tests (httptest). Add a `landing` typecheck/lint CI step.

17. ✅ **DONE (this pass)** — `npx tsc --noEmit` is now **0 errors**. Added `outline`/`ghost` compat
    variants + exported `ButtonProps` in `ui/button.tsx`, and fixed the `darkMode` type. `next build`
    passes type-checking.

---

## 🟢 P3 — Growth & polish

18. **Onboarding wizard** — first-run flow: create project → generate key → create a starter rule → fire a
    test request in the Playground. Turns a blank dashboard into a "aha" in 60 seconds.
19. **Team / multi-user workspaces** — invite members to a project (schema already isolates by `UserID`;
    generalize to org membership).
20. **SDKs beyond Go** — the Go SDK (`sdk/`) is solid; add JS/Python thin clients hitting `/gateway`.
21. **Usage-based billing** — meter gateway requests per plan and surface overage in Billing.
22. **Landing polish** — finish aligning copy to the real product (see the About/Pricing edits), add an
    interactive "watch a request get throttled" demo, and real OG images.

---

## Suggested order

P0 (1→5) to make the core loop trustworthy → P1 (6, 7, 10) for the features users see →
P2 (12, 16, 17) before any public launch → P3 as growth work.
