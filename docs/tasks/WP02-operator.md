# WP02-operator — Platform operator bootstrap and console

Follow-up to WP02 based on owner decisions resolving OQ-16 (2026-11-22).

## Scope delivered
- Operator identity: Cloudflare Access on `admin.<domain>`; no app passwords, TOTP seeds or sessions.
- Dev shim: `admin.localhost` host sets `Cf-Access-Authenticated-User-Email` from a trusted proxy.
- CLI: `go run ./backend/cmd/operator create <email>` maintains the allowlist (idempotent).
- ADM-01: operator API to create, suspend, reinstate and set quota/caps for organizations.
- AUD-03: operator-level audit view (operator rows only, null `organization_id`).
- Admin console at `/admin` on the operator host.

## Deferred (owner decision)
- ADM-02 abuse signals: WP11 (campaigns/analytics do not exist yet).
- Support grants, emergency access (OQ-08), organization offboarding (OQ-07): M1.
- Caps stay off by default (OQ-10); tenant lockout defaults unchanged (OQ-13).

## Local access
1. Start the dev stack (`make dev` / docker compose). Migrations create `operators`.
2. Add `127.0.0.1 admin.localhost` to `/etc/hosts` (uses the same loopback as `localhost`).
3. Create the operator with both tenant and system DSNs:
   `DATABASE_URL=postgres://email_app:app-dev@localhost:5432/email_platform_dev?sslmode=disable SYSTEM_DATABASE_URL=postgres://email_system:system-dev@localhost:5432/email_platform_dev?sslmode=disable go run ./backend/cmd/operator create dev-operator@example.com`
   (idempotent; re-running reports that the operator already exists).
4. Open `http://admin.localhost:8081/admin`. nginx on that host injects the identity header;
   the tenant host clears it, so the API cannot be reached as an operator there.

## Contracts
- DB: `backend/db/migrations/0012_operator_access.sql`; queries in `backend/db/queries/operator.sql`.
- API: `/operator/session`, `/operator/organizations` (GET/POST, `/:id/suspend`, `/:id/reinstate`,
  `/:id/limits`), `/operator/audit` in `api/openapi.yaml`.
- Go: `platformadmin.Repository`, `IdentityService`, `Service`; `api.OperatorIdentityService`,
  `api.PlatformAdminService`; `identity.PlatformAdminActor`; `httpx.IsTrustedPeer`.

## Security notes
- Header is trusted only when the immediate peer is in `TRUSTED_PROXY_CIDRs`; `X-Forwarded-For` is ignored.
- Lookup is a `SECURITY DEFINER` function; email normalized; unknown/suspended operators get an opaque 403.
- `WithSystem` uses a dedicated `SYSTEM_DATABASE_URL` connection as `email_system`; `email_app` is deliberately **not** a member of that role (membership would inherit its RLS policies). Tenant requests still use `WithOrg`.
- OQ-17: staging/production must restrict header forwarding to the authenticated CF tunnel or verify the JWT.
