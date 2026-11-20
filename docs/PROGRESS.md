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

- Commit WP01 and open the review.
- Create the WP02 branch from the green WP01 branch.

### Handoff

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
