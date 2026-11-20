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
  forced RLS is OQ-12. Full `make dev-up` was not run because host port 8080 is
  occupied by an unrelated container; migrate/api/worker were verified directly.
- 2026-11-20 WP02.1 handoff: schema migrations 0002-0006, forced RLS, composite
  FKs, role grants, DSN split and dev bootstrap are complete.
  `make test-module M=platform`, `go test -race ./backend/internal/platform/db/...`,
  `make lint` and `make ci` pass; migrate/api/worker start on `email_app` and RLS
  fails closed without `app.org_id`. Next: WP02.2 `WithOrg` helper and catalog
  guard test. Unfinished: OQ-12 pre-auth login lookup; full `make dev-up` blocked
  by host port 8080; no commit.
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
