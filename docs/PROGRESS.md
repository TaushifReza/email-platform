## Progress

### WP01: Foundations

- [x] Set up basic project structure
- [x] Implemented core database schema with RLS
- [x] Configured logging and error handling
- [x] Added initial migration for tenant tables
- [x] Set up SMTP provider configuration
- [x] Implemented email template system
- [x] Added recipient data processing
- [x] Set up basic authentication flow
- [x] Fixed linter errors in config package (err113, mnd, gci, godot, nlreturn, wsl)
- [x] 2026-11-18 WP01 3: Created dev/staging compose files (postgres, redis, mailpit, S3), nginx config, /healthz and /readyz endpoints. Makefile dev-up/down targets functional. Tests not run due to module conflicts. Next: WP01 4 - GitHub Actions workflow with Postgres service, go test, lint, govulncheck, gosec, frontend typecheck and build, secret scan.
- [x] 2026-11-18 WP01 4: Created .github/workflows/ci.yml with jobs for tests (Postgres service), linting, security checks (govulncheck, gosec, gitleaks), and conditional frontend typecheck/build. Workflow tests pass. Next: WP01 5 - pgx pool and migrate command; empty first migration.
- [x] 2026-11-19 WP01 5: Implemented pgx connection pool with proper settings (MaxConns=10, MinConns=1, MaxConnLifetime=30m), goose migrate command that runs migrations against PostgreSQL, and created first empty migration file. Added tests for pool connectivity and migration idempotency which pass with DATABASE_URL set. Next: WP01 6 - OpenAPI scaffold with Go server stubs and TypeScript client.
- [x] 2026-11-19 WP01 6: Created OpenAPI scaffold with Gin framework; implemented /healthz and /readyz endpoints; updated main.go to use Gin instead of ServeMux; added make gen target for code generation. Built successfully, tests pass. Next: WP01 7 - Frontend shell with Next.js static export, Tailwind, shadcn/ui.

## Current Work Package: WP01: Foundations

This work package covers:
- Initial project setup including directory structure and dependencies
- Core database tables with proper RLS and tenant isolation
- Basic application configuration handling
- SMTP provider integration for sending emails
- Email template management system
- Recipient data processing workflows
- Authentication mechanism implementation
- Linter compliance for config package

## Next Tasks

- Implement email sending protocol with rate limiting
- Add comprehensive audit logging
- Create dashboard UI components
- Set up notification system

- [x] 2026-11-19 WP01 7: Created frontend shell with Next.js static export, Tailwind, shadcn/ui components, API client placeholder and CSP generation spike. Directory structure established with app/layout.tsx, app/page.tsx, app/login/page.tsx, component libraries (button, input, card, label), UI utilities, and lib/health.ts for backend health checking.
- [x] 2026-11-19 WP01 7: Frontend shell implementation complete. Created Next.js 15 static export with Tailwind CSS, shadcn/ui components, placeholder login, API client health check and CSP spike framework. Verified frontend build structure works. All tests passed. Next: WP01 8 - tools/fakesmtp: Go SMTP test server with scripted behaviors and HTTP API.
- [x] 2026-11-19 WP01 8: Implemented fakesmtp tool package in tools/fakesmtp with core Server, Script, Behavior, and Message types; SMTP protocol handling; self-signed TLS support; HTTP API endpoints (list, get, delete); and basic HTTP server scaffolding. Package compiles successfully with no external dependencies. Tests pass including go vet, golangci-lint, and make ci.