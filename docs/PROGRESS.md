## Progress

### WP01: Foundations

Status: complete. Verified locally with `make ci`, `make dev-up`, `make gen` and
`bash tools/redflags.sh`.

| Sub-task | Status | Notes |
|---|---|---|
| 1. Repo skeleton, Makefile, lint, sqlc, goose | Done | Single root Go module, real Make targets, sqlc v1.31.1 |
| 2. Config loader and redacting zap logger | Done | Env and `*_FILE` secrets; redaction tests pass |
| 3. Dev/staging Compose, nginx, health endpoints | Done | Dockerfiles, RustFS bucket init, web service, health checks |
| 4. GitHub Actions | Done | Go 1.26, Postgres service, contracts, lint, frontend, security |
| 5. pgx pool and migrate command | Done | Embedded goose migrations; idempotency test passes |
| 6. OpenAPI scaffold and TS client | Done | oapi-codegen v2.8.2 and openapi-typescript v7 |
| 7. Frontend shell | Done | Static export, Tailwind, shadcn/ui, API client, CSP spike |
| 8. fakesmtp | Done | Scripted SMTP, TLS modes, messages API, self-tests |

### Corrected scope note

WP01 contains no business features. RLS schema, SMTP provider configuration,
templates, recipient processing, authentication, audit logging and dashboards
belong to later work packages and were not implemented here.

### Verification

- `make test` and `go test -race ./backend/... ./tools/...` pass.
- `make lint` passes with the project `.golangci.yml`.
- `make gen-check` reproduces generated files; CI also verifies they are committed.
- `make dev-up` starts Postgres, Redis, Mailpit, RustFS, migrations, API, worker,
  web and nginx; `/healthz`, `/readyz`, `/api/healthz`, `/` and `/login` respond.
- `make security` passes `govulncheck` with Go 1.26.11 and `gosec`.
- `make ci` passes.
- `bash tools/redflags.sh` reports no red flags.

### Deviations

- The generated TypeScript contract is used with a small typed `fetch` wrapper in
  `frontend/lib/api.ts`. `openapi-fetch` rejected the generated `paths` interface
  as a client generic, so the runtime client was dropped instead of duplicating
  casts.

### Known residual risk

- `npm audit` reports a high-severity advisory in Next.js' nested PostCSS build
  dependency. Fixing it requires a Next.js major upgrade, which is out of WP01.

### Next

- WP02 sub-task 2: `WithOrg` helper plus the catalog guard test.
- Then WP02 sub-task 3: password hashing, login, lockout and rate limits.

### Handoff

- 2026-11-20 WP02.1 identity/tenancy schema: added goose migrations 0002-0006
  for `email_owner`/`email_app`/`email_system`, organizations, operators, users,
  sessions, operator_sessions, invitations, auth_tokens and a partitioned
  append-only audit_log; forced RLS, composite FKs, app/system grants and the
  `app.org_id` GUC. Split runtime and migration DSNs (`DATABASE_URL` vs
  `MIGRATION_DATABASE_URL`), added an idempotent dev postgres bootstrap service
  so existing volumes get LOGIN roles, and mapped sqlc UUID/inet types to stdlib
  to avoid new dependencies. Tests cover AC-SEC-01, AC-AUD-01, ORG-05/06/09/10
  and schema/role checks. `make test-module M=platform` and `make ci` pass.
  Deviation: `operator_sessions` is separate from `sessions` so every tenant
  table keeps `organization_id NOT NULL`. Open: pre-auth login lookup under
  forced RLS is OQ-12. Full `make dev-up` was not run because host port 8081 is
  occupied by an unrelated container; migrate/api/worker were verified directly.
- 2026-11-20 WP02.1 handoff: schema migrations 0002-0006, forced RLS, composite
  FKs, role grants, DSN split and dev bootstrap are complete.
  `make test-module M=platform`, `go test -race ./backend/internal/platform/db/...`,
  `make lint` and `make ci` pass; migrate/api/worker start on `email_app` and RLS
  fails closed without `app.org_id`. Next: WP02.2 `WithOrg` helper and catalog
  guard test. Unfinished: OQ-12 pre-auth login lookup; full `make dev-up` blocked
  by host port 8081; no commit.
- 2026-11-19 WP01 remediation: started on `wp01-foundations`. Consolidated the Go
  backend and `tools/fakesmtp` into the root module, removed the duplicate root
  `internal/` scaffold, replaced the Makefile with real `test`, `lint`, `gen`,
  `gen-check`, frontend and security targets, and corrected this progress log.
- 2026-11-20 WP01 completion: implemented config and redaction tests, pgx/goose
  migrations, health service and OpenAPI-generated Gin handlers, Dockerfiles,
  Compose stack with a real `web` service and RustFS bucket creation, typed
  frontend health client, CSP hash spike, and a scripted fakesmtp with SMTP/TLS
  and HTTP API self-tests. Moved to Go 1.26.11 to clear standard-library
  vulnerabilities. No commit was created in this session.
- 2026-11-21 WP02.2 (`WithOrg` + catalog guard): added `db.WithOrg`, which runs
  each tenant transaction as `email_app` with `app.org_id` set and fails closed
  on an empty organization; `db.DB` hides raw query methods and uses
  `database/sql` (pgx stdlib) to match sqlc. The catalog guard requires forced
  RLS plus an `email_app`/PUBLIC policy on every tenant table and composite
  `organization_id` FKs between them; `audit_log` partitions are skipped and the
  parent checked. `make test-module M=platform/db`,
  `go test -race ./backend/internal/platform/db/...` and `make ci` pass.
  Next: WP02.3 password hashing, login, lockout and rate limits. Committed;
  handler/service/repository and frontend layering documented in
  `docs/technical-design/01-architecture.md`.

- 2026-11-21 WP02.3: argon2id hasher and policy, common-password list, SECURITY DEFINER pre-auth lookup, login service, lockout, Redis rate limiter, uniform errors, and POST /auth/login contract. Tests cover AUTH-01/02/04, SEC-12/16, BR-01, AC-SEC-01. `make test-module M=identity`, `make test-module M=platform/ratelimit`, `make test-module M=platform/db`, `make test-module M=api`, `make ci` pass. Next: WP02.4 sessions, cookies, CSRF and trusted-proxy IP. Open: OQ-13 lockout/rate-limit defaults; per-IP limiting, login audit events, and session issuance deferred to WP02.4/WP02.8; operator login not yet covered.

- 2026-11-21 WP02.3: argon2id hasher/policy, common-password list, SECURITY DEFINER pre-auth lookup, login service, lockout, Redis rate limiter, uniform errors, POST /auth/login. Tests: identity, ratelimit, db, api modules and REDIS_URL make ci plus redflags all passed. Next: WP02.4 sessions, cookies, CSRF, trusted-proxy IP. Unfinished: per-IP limiting, login audit, session issuance, operator login, OQ-13.
- 2026-11-22 WP02.4 (sessions, cookies, CSRF, trusted-proxy IP): added 0008_session_lookup function, sqlc session queries, identity session repository/service (SignIn/Authenticate/SignOut), httpx client IP, CSRF/auth middleware, cookie helpers, OpenAPI endpoints, config + compose updates. Tests: identity/api/httpx non-DB pass; db tests need Postgres. `make ci` passes. Next: WP02.5 TOTP enrollment/verification. Open: per-IP login limit added; audit of sign-in deferred to WP02.8.

- 2026-11-22 WP02.5 (TOTP/MFA): added crypto keyring (AES-256-GCM/HKDF), TOTP+recovery codes, MFA service (enroll/verify/step-up), migration 0009, OpenAPI endpoints, pending-session middleware. Tests: crypto/identity/api unit tests pass; DB tests need Postgres. Next: WP02.6 invitations/password reset. Unfinished: operator TOTP wiring, audit events (WP02.8), step-up TTL OQ-14.

- 2026-11-22 WP02.6 (invitations + password reset): added migration 0010 (system_emails outbox + invitation/reset SECURITY DEFINER lookups), sqlc queries, platform/outbox store, db.WithSystem, identity access service, public accept/reset endpoints, notifications SMTP sender + worker processor, config SMTP_TLS_MODE/APP_BASE_URL, OpenAPI endpoints. `make ci` passes. Unfinished: invitation creation HTTP endpoint and audit events deferred to WP02.8; frontend pages WP02.9; invitation user pre-creation (WP02.8); no tests added for new service yet.

- 2026-11-22 WP02.6: system_emails outbox migration + SECURITY DEFINER lookups, sqlc queries, platform/outbox store, db.WithSystem, identity access service, public accept/reset endpoints, SMTP sender + worker, config SMTP_TLS_MODE/APP_BASE_URL, OpenAPI. make ci passes; redflags clean. Next: WP02.7 central permission matrix. Unfinished: invite HTTP route and audit events deferred to WP02.8; frontend WP02.9.
- 2026-11-22 WP02.7 (central permission matrix): added backend/internal/identity/permissions.go with Capability constants, role matrix, Request and Authorize (fail closed; scope rules for own/approve-never-own/admin-not-owners/PA grant/pause); replaced ad-hoc role check in access_service.go; role constants moved here; validRole reused. tests/permissions_test.go covers SRS 6.2 role matrix plus scope rules. make ci passes. Next: WP02.8 organization and user management services/endpoints.

- 2026-11-22 WP02.8 (partial): audit writer + tenant org settings + user management (invite/role/deactivate) + audit events + OpenAPI endpoints. Operator org-management and operator auth deferred to OQ-16. make ci passes. Next: operator authentication sub-task, then org-management endpoints.

- 2026-11-22 WP02.8: added audit writer (append-only, AUD-01/02/04), tenant org settings (profile/policy) and user management (invite/role-change/deactivate with last-owner guard, step-up, session revocation) plus OpenAPI endpoints and cmd/api wiring. Deferred operator auth + operator org-management to OQ-16. Tests: go build, make lint, make gen-check, go test ./backend/..., make ci all pass. Next: operator authentication sub-task. Unfinished: retrofit audit events into login/MFA flows; frontend WP02.9.

- 2026-11-22 WP02.9 (frontend): replaced static login stub with real client flows for login, MFA entry and enrollment (secret/recovery codes), invitation accept (`/invitations/accept#token=`), protected shell with session guard, organization settings and user list/invite (step-up before role/deactivation). Extended `lib/api.ts` with credentials, CSRF header, problem mapping and 204 handling; added vitest unit tests. Backend enablers: `LoginResponse` now reports `mfa_enrollment_required` (also added to `SessionResult`), invitation emails link to the static-friendly fragment URL. Tests: `npm test`, `npm run typecheck`, `npm run lint`, `npm run build`, `go test ./backend/api/...`, `make test-module M=identity` (Postgres started), `make ci` all pass. No commit. Unfinished: operator auth/org management (OQ-16), audit retrofit into login/MFA flows, password-reset page, org logo (not in API), QR rendering; `MFAService.StepUp` still ignores the password argument (owner decision needed).

- 2026-11-22 CSRF fix: web UI logins were rejected because nginx rewrote `Host` with `$host` (drops the port) while browser `Origin` includes it, so the backend Origin check failed. Changed all proxy locations to `proxy_set_header Host $http_host;` and reloaded nginx; verified browser-style request with `Origin: http://localhost:8081` now reaches the login handler. Added `backend/api/middleware_test.go` covering same-origin port matching and the nginx header contract. `go test ./backend/api/...` and `make ci` pass. A login still needs a real user: no `dev-owner@example.com` row exists yet (bootstrap SQL was not applied), so the API now returns `invalid_credentials`, not `csrf_rejected`.

- 2026-11-22 WP02-operator (OQ-16, ADM-01, AUD-03, SEC-02): implemented Cloudflare-Access-based operator identity (no app passwords/TOTP/sessions), CLI `backend/cmd/operator create <email>` allowlist bootstrap, migration 0012 + SECURITY DEFINER lookup, sqlc queries, OpenAPI `/operator/*` contracts, `platformadmin` service/repository, trusted-peer-only header auth middleware, operator organization create/suspend/reinstate/limits, operator audit view, `/admin` console, dev `admin.localhost` nginx shim, config/env/docs. Audit writer now generates row IDs (latent bug affecting all prior audit inserts) and audit SQL accepts nullable tenant/actor fields. Tests: `make test-module M=platformadmin` (with dev Postgres), `go test ./backend/...`, `npm test`, `npm run typecheck`, `npm run lint`, `npm run build`, `make lint`, `make gen-check`, `make ci` all pass. Next: staging/production CF Access verification (OQ-17; header trust restricted to tunnel or JWT verification), ADM-02 abuse signals in WP11, M1 support grants/emergency/offboarding. No commit created; prior WP02.9 changes remain uncommitted too.

- 2026-11-22 operator seeding follow-up: dev DB had only test-generated operators, so seeded `dev-operator@example.com` via the CLI. Found and fixed a real isolation bug: granting `email_app` membership in `email_system` (to make `SET LOCAL ROLE` work) also makes it inherit `email_system` RLS policies, breaking `WithOrg` tenant isolation. Reverted that approach and added a dedicated `SYSTEM_DATABASE_URL` connection used by `WithSystem`; `email_app` is explicitly not a member of `email_system` (new schema test). Updated config, compose, `.env.example`, CLI and docs. Also reloaded nginx so the bind-mounted admin host config took effect. Verified via curl: `/api/operator/session`, `/api/operator/organizations`, `/api/operator/audit` and `/admin` all return 200 on `admin.localhost`. `make test-module M=platform/db` and `make ci` pass. No commit created.
