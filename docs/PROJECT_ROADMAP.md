# Limiter.io — Roadmap & Next Steps

A prioritized, grounded list of what to build next — based on the actual code in this repo
(`internal/`, `sdk/`, `landing/`). Ordered by impact. Each item notes *why* and *where*.

---

## 🔴 P0 — Correctness & "make the demo actually work"

1. **Seed the `plans` table (free / pro / enterprise).**
   `subscription_service.go:63` calls `GetPlanByID` — upgrades and plan limits break if the table is
   empty. Add a migration/seed with the three plans and their `allowed_algorithms`, `max_projects`,
   `rate_limit_requests`, `analytics_retention_days`.

2. **Give every new user a `free` subscription row on register.**
   `GetSubscription` / `UpgradeSubscription` (`subscription_service.go:53`) error if the user has no
   subscription. Create one in the auth/register flow so `/subscription` and Billing don't 500.

3. **Fix the Lemon Squeezy variant check.**
   `billing_handler.go:83` does `string(rune(payload.Data.Attributes.VariantID))` — that converts an int
   to a Unicode code point, not the string "1899978". It only works today because of the hardcoded
   `|| == 1899978` fallback. Compare the int directly to a configured `LEMON_SQUEEZY_PRO_VARIANT_ID`.

4. **Wire real email for `forgot-password` / verification.**
   `POST /auth/forgot-password` exists but there's no mail transport. Add SMTP/Resend/SES and a reset-token
   flow, or clearly mark it as a stub. The frontend page is already built (`/forgot-password`).

5. **Enforce plan limits.**
   Today keys/projects/algorithms aren't capped by plan. Enforce `max_projects`, `max_keys_per_project`,
   and `allowed_algorithms` in the project/key/policy services so Free vs Pro actually differ.

---

## 🟠 P1 — Product features (high user value)

6. **Real log pagination + filtering on the Overview page.**
   The API already supports `limit`/`offset` (`analytics_handler.go:78`). The dashboard currently loads
   only the first page. Add prev/next wired to offset, plus filters by decision / route / date range.

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

12. **Tighten CORS.** `router.go:49` sets `Access-Control-Allow-Origin: *` with credentials — lock to the
    dashboard origin(s) via config before production.

13. **Rate-limit-headers on allowed gateway responses always.** They're set only when a rule matches
    (`ratelimit.go:122`). Consider returning them (or a `X-RateLimit-Policy: none`) even on pass-through so
    clients can introspect.

14. **Refresh-token rotation & revocation on logout.** Ensure `/auth/logout` revokes the refresh token
    server-side (the `RefreshToken.Revoked` field exists in `models.go:37`).

15. **Observability.** Prometheus metrics already exist (`/metrics`). Add a Grafana dashboard JSON to
    `deploy/`, and structured request logging with the existing `RequestID`.

16. **Tests.** Only `ratelimit_test.go` exists. Add unit tests for each algorithm's Lua script, auth/JWT,
    and handler integration tests (httptest). Add a `landing` typecheck/lint CI step.

17. **Fix pre-existing frontend type errors.** `npx tsc --noEmit` flags template components
    (`ui/carousel`, `ui/pagination`, `ui/alert-dialog`, `ui/calendar`) still using old shadcn button
    variants (`outline`/`ghost`) that the neobrutalism `button.tsx` renamed. Align or remove unused ones so
    `next build` passes.

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
