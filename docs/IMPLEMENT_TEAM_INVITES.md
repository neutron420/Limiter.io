# IMPLEMENT: Team Invitations with Email + Role-Based Access (RBAC)

**Goal:** Replace the current "add member instantly" behavior with a real invitation flow:

1. Owner invites `teammate@company.com` with a role → **invitation email** is sent (Resend).
2. The invitee clicks **Accept Invite** in the email → logs in / registers → accepts.
3. **Only after accepting** do they get access — the shared project appears in *their* dashboard
   (switcher + Projects page) with keys, rules, and analytics.
4. Roles actually mean something:
   - **`admin` (read-write)** — can view everything AND create/edit/delete keys, rules; run playground.
   - **`member` (read-only)** — can view dashboards, analytics, keys list, rules; **cannot** mutate anything.
   - **`owner`** — the project creator; everything + manage members + delete project.

---

## What exists today (baseline — don't rebuild these)

| Piece | Where | State |
|---|---|---|
| `ProjectMember` model (`project_id`, `user_id`, `role`, `email`) | `internal/models/models.go` | ✅ exists |
| Member repo (Add/Remove/List/IsMember/ListProjectIDsByUser) | `internal/repository/postgres/postgres_repos.go` | ✅ exists |
| Endpoints `GET/POST/DELETE /projects/:projectId/members` | `router.go:146-148` | ✅ exists |
| `AddMember` — **adds instantly, no invite/email** | `project_service.go:128` | ⚠️ to change |
| Access check treats every member the same (no role check) | `checkProjectAccess` in services | ⚠️ to change |
| Shared projects listed for members | `project_service.go:102` (`ListProjects`) | ✅ exists |
| Resend mailer (`mailer.New`, used by password reset) | `internal/mailer/mailer.go` | ✅ reuse it |
| Settings UI with invite form + role dropdown | `landing/app/dashboard/settings/page.tsx` | ⚠️ extend |

---

## Phase 1 — Backend: invitation model + flow

### 1.1 New model: `ProjectInvite` (`internal/models/models.go`)

```go
type ProjectInvite struct {
    ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
    ProjectID  uuid.UUID  `gorm:"type:uuid;index;not null" json:"project_id"`
    Email      string     `gorm:"index;not null" json:"email"`          // invitee (may not be a user yet)
    Role       string     `gorm:"not null;default:member" json:"role"`  // admin | member
    TokenHash  string     `gorm:"uniqueIndex;not null" json:"-"`        // SHA-256 of the emailed token
    InvitedBy  uuid.UUID  `gorm:"type:uuid;not null" json:"invited_by"`
    Status     string     `gorm:"not null;default:pending" json:"status"` // pending | accepted | revoked | expired
    ExpiresAt  time.Time  `gorm:"not null" json:"expires_at"`             // 7 days
    AcceptedAt *time.Time `json:"accepted_at,omitempty"`
    CreatedAt  time.Time  `json:"created_at"`
}
```

Add to `AutoMigrate` in `internal/database/postgres.go`.
Token pattern = same as password reset: `utils.GenerateRandomToken(32)` emailed raw, store `utils.HashAPIKey(raw)`.

### 1.2 Repo (`interfaces.go` + `postgres_repos.go`)

```go
type ProjectInviteRepository interface {
    Create(ctx, inv *models.ProjectInvite) error
    GetByTokenHash(ctx, hash string) (*models.ProjectInvite, error)
    ListByProject(ctx, projectID uuid.UUID) ([]models.ProjectInvite, error)   // pending only, for UI
    ListPendingByEmail(ctx, email string) ([]models.ProjectInvite, error)     // "invites waiting for you"
    Update(ctx, inv *models.ProjectInvite) error                              // status transitions
}
```

### 1.3 Service changes (`project_service.go`)

**Replace `AddMember`** with `InviteMember(ctx, ownerID, projectID, req)`:
1. Assert caller is project **owner** (or `admin` once Phase 2 lands).
2. Validate role ∈ {`admin`, `member`}; reject inviting the owner's own email; reject an existing member;
   revoke any previous pending invite for the same email+project (delete or mark `revoked`).
3. Create invite (7-day expiry), then send email via the injected `mailer.Mailer`:

```go
inviteURL := fmt.Sprintf("%s/accept-invite?token=%s", strings.TrimRight(cfg.AppBaseURL, "/"), rawToken)
subject := fmt.Sprintf("You've been invited to %q on Limiter.io", project.Name)
// brutalist HTML template — copy the style from ForgotPassword's email in auth_service.go
```

> `NewProjectService` gains `mailer.Mailer` + `cfg *config.Config` params — mirror how
> `NewAuthService` receives them; update the call in `cmd/api/main.go`.

**New `AcceptInvite(ctx, userID, rawToken)`**:
1. Hash token → `GetByTokenHash`; reject if `Status != "pending"` or `ExpiresAt` passed.
2. **Email must match**: the logged-in user's email must equal `invite.Email` (case-insensitive) —
   otherwise someone who intercepts the link could join with the wrong account.
3. Create the `ProjectMember` row (`role` from the invite), set invite `Status = "accepted"`, `AcceptedAt = now`.
4. Return the project so the frontend can switch to it.

**Also add**: `ListInvites` (owner sees pending), `RevokeInvite` (owner cancels), `ListMyInvites`
(user sees invites addressed to their email → in-app accept without digging through email).

### 1.4 Endpoints (`router.go`)

```text
POST   /projects/:projectId/invites          owner/admin → create + send email
GET    /projects/:projectId/invites          owner/admin → pending list
DELETE /projects/:projectId/invites/:id      owner/admin → revoke
GET    /invites                              me → pending invites for my email
POST   /invites/accept        {token}        me → accept (JWT required — never public)
```

Keep the existing `/members` GET/DELETE (list + remove). **Delete the old instant-add
`POST /members`** handler wiring so the invite path is the only way in.

---

## Phase 2 — Backend: role enforcement (read-only vs write)

### 2.1 One helper, one rule (put it in each service or a shared spot)

```go
// roleForProject: "owner" | "admin" | "member" | "" (no access)
func roleForProject(ctx, projectRepo, memberRepo, userID, projectID) string

func canRead(role string) bool  { return role != "" }
func canWrite(role string) bool { return role == "owner" || role == "admin" }
```

### 2.2 Enforcement matrix — apply in the service layer (not handlers)

| Action | owner | admin | member (read-only) |
|---|---|---|---|
| View project / keys / rules / analytics / logs / WS stream | ✅ | ✅ | ✅ |
| Create/rotate/revoke/delete API keys | ✅ | ✅ | ❌ `403` |
| Create/update/delete/toggle rate rules | ✅ | ✅ | ❌ `403` |
| Rule simulate (dry-run) | ✅ | ✅ | ❌ |
| Invite / revoke / remove members | ✅ | ✅ (not owner) | ❌ |
| Delete project | ✅ only | ❌ | ❌ |

Where to patch (each currently checks only `proj.UserID == userID` or `IsMember`):
- `apikey_service.go` — Create/Rotate/Revoke/Delete → `canWrite`; List → `canRead`
- `policy_service.go` — Create/Update/Delete/Simulate → `canWrite`; List/Get → `canRead`
- `analytics_service.go` — `checkProjectAccess` → `canRead` (already effectively this)
- `project_service.go` — Delete → owner only (already); member mgmt → owner or admin
- `ws_handler.go` — currently owner-only! Loosen to `canRead` so members see the live stream.

Return a distinct error (`"insufficient role: read-only members cannot modify the project"`) so the
frontend can show a friendly message.

### 2.3 Include role in API responses

`GET /projects` and `GET /projects/:id` should include `"role": "owner|admin|member"` for the caller
(compute per project). The frontend uses this to hide write buttons.

---

## Phase 3 — Frontend (landing/)

### 3.1 Accept page — `app/accept-invite/page.tsx` (new)
- Reads `?token=`. If not logged in → redirect to `/login?next=/accept-invite?token=...`
  (add `next` redirect support to the login page) — register works the same way.
- Logged in → show brutalist card: *"You've been invited to **{project}** as **{role}**"* +
  **ACCEPT INVITE** button → `POST /invites/accept` → `refresh()` project context → `select(projectId)`
  → router.push `/dashboard`. Handle expired/revoked/wrong-email errors inline.

### 3.2 Settings page — upgrade the Team panel
- Invite form → `POST /projects/:id/invites` (roles: `MEMBER (READ-ONLY)`, `ADMIN (READ-WRITE)` — the
  dropdown already exists, wire the value through).
- Two lists: **Members** (accepted, with role badge + remove) and **Pending invites**
  (email + role + expiry + revoke button).
- Show a "PENDING" amber badge (brutalist: `border-yellow-500 text-yellow-600 bg-yellow-500/10`).

### 3.3 Role-aware UI (use `role` from `GET /projects`)
- Store `role` in `project-context.tsx` alongside `current`.
- `member` role → hide/disable: New Rule, Edit, Delete, toggle (policies); New Key, Rotate, Revoke,
  Delete (keys); invite form (settings). Show a small `READ-ONLY` badge next to the project name in
  the breadcrumb so it's obvious why.
- `admin` → everything except Delete Project + can't remove the owner.

### 3.4 Invite awareness
- On dashboard load, call `GET /invites`; if pending invites exist show a banner:
  *"📨 You have 1 pending project invite — View"* → accept inline.

---

## Phase 4 — Emails (reuse `internal/mailer`)

Two templates (match the password-reset HTML style — mono font, `#ea580c` button):
1. **Invitation** → to invitee: project name, inviter email, role, big ACCEPT button, 7-day note,
   plain URL fallback.
2. **Accepted notification** (nice-to-have) → to owner: "{email} joined {project} as {role}".

Dev mode without `RESEND_API_KEY` still works — the mailer logs the accept link to stdout.

---

## Security checklist (do not skip)
- [ ] Store only the **hash** of invite tokens; raw token appears once, in the email.
- [ ] Accept requires **JWT + matching email** — the link alone must never grant access.
- [ ] Expire invites (7d) and honor `revoked` status on accept.
- [ ] All role checks live in the **service layer** (UI hiding is cosmetic, not security).
- [ ] Owner cannot be removed or downgraded; admin cannot remove the owner or delete the project.
- [ ] Don't leak whether an email has an account in invite/accept error messages.

## Test checklist
- [ ] Invite → email sent → invite `pending`; second invite for same email replaces the first.
- [ ] Accept with the wrong logged-in email → 403 with clear message.
- [ ] Accept → member row created, project shows in invitee's dashboard, role correct.
- [ ] `member` gets `403` on every write endpoint (keys/rules/simulate) but `200` on reads + WS.
- [ ] `admin` can write but cannot delete the project or remove the owner.
- [ ] Expired/revoked token → friendly error on the accept page.
- [ ] Unit tests: role helper (`canRead`/`canWrite`), invite token round-trip, accept transitions.

## Suggested implementation order
1. Model + repo + migrate (1.1–1.2)
2. Invite/Accept service + endpoints + email (1.3–1.4, Phase 4)
3. Role enforcement in services + `role` in responses (Phase 2)
4. Frontend: accept page → settings panel → role-aware UI → invite banner (Phase 3)
5. Tests (checklist above)
