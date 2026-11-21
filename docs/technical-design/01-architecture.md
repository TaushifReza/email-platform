# Architecture

## Processes (same Go codebase and image)
| Process | Responsibility |
|---|---|
| `api` | JSON API (Gin), sessions, uploads, the public unsubscribe endpoint |
| `worker` | Imports, attachment processing, scheduler tick, per-provider sender, notifications, purge (M1) |
| `migrate` | Runs goose migrations then exits |

## Backend layout
```
backend/
  cmd/api  cmd/worker  cmd/migrate
  internal/
    identity  tenancy  providers  templates  imports  attachments
    campaigns sending  suppression reporting notifications audit platformadmin
    platform/  (config, logging, db, crypto, s3, httpx, ratelimit, clock)
  db/migrations  db/queries   (sqlc input)
  api/openapi.yaml
```
Rules: handlers are thin and call services. Services take an `Actor` (user id, organization id, role). Modules talk through interfaces and never read each other's tables. Time comes from an injectable clock so tests control it.

## Backend layering
Every module under `internal/<module>` uses the same layers; dependencies point inward only.

| Layer | Location | Responsibility |
|---|---|---|
| Handler | `backend/api` | Implement the generated OpenAPI interface. Decode and validate transport input, call one service method, map errors to HTTP. No SQL and no business rules. |
| Service | `internal/<module>/service.go` | Use cases, authorization through the central permission function, transaction boundaries through `db.WithOrg`, and orchestration of repositories and other services. |
| Repository | `internal/<module>/repository.go` | All module SQL through sqlc, inside the transaction the service opened. Implements the repository interfaces declared by the service. |

- The service declares the repository interface it consumes; the repository is the only layer that imports `sqlc` or knows table names.
- Cross-module access goes through the other module's service interface, never its repository or tables.
- Handlers and services never build SQL; repositories never make authorization decisions.
- Service tests use fake repositories, repository tests use the real schema and RLS, handler tests use fake services.
- Domain types and errors live in `internal/<module>`; handlers translate them to OpenAPI responses.

## Request and job context
Every request and job opens a transaction and runs `SET LOCAL app.org_id`. The helper is `WithOrg(ctx, orgID, func(tx) error)`. System jobs that must span organizations (scheduler tick, recovery sweeper, purge) use a separate narrow database role with explicit policies.

## Queues
- Send queue: the `campaign_recipients` table itself, claimed with `FOR UPDATE SKIP LOCKED`.
- Other jobs (import, ZIP processing, system emails, notifications, purge): a Postgres job library.
- Redis: login and upload rate limits, wake-up signals. If Redis is down nothing is lost or duplicated.

## Files
RustFS through S3 API. Keys are generated IDs: `org/{org_id}/campaign/{campaign_id}/{uuid}`. Never user-supplied names. No public URLs. Downloads are streamed by the API with `Content-Disposition: attachment` and `nosniff`.

## Frontend
Next.js static export, client-side only, talking to `/api`. Typed client generated from OpenAPI. Session cookie set by Go. The unsubscribe page is rendered by Go, not by Next.js.

Frontend layering mirrors the backend: route -> feature -> service -> generated client.
```
frontend/
  app/                  route segments and layouts only
  features/<feature>/   feature components, hooks and service modules
  components/           shared presentational components
  components/ui/        shadcn/ui primitives
  lib/api.ts            fetch wrapper: credentials, CSRF, error mapping
  lib/types/openapi.ts  generated types (never edit)
```
- Pages in `app/` only compose feature components and read route params; no fetching or business rules.
- A feature owns its UI, hooks (`use<Feature>`) and service (`<feature>Service`) that calls the generated client.
- Components are presentational and receive data and callbacks.
- All HTTP goes through the typed client or `lib/api.ts`; no ad-hoc `fetch` in pages or components.
- Server state lives in feature hooks (TanStack Query once introduced per D-01); form state stays local.
- Request and response types come from `lib/types/openapi.ts`; do not hand-write them.

## Network
```
Internet -> Cloudflare (WAF, rate limits) -> dedicated Tunnel -> nginx
   /api/*  /u/*  -> api container
   /*            -> web container (static files)
admin.<domain> behind Cloudflare Access
postgres, redis: internal docker network only
worker egress: internet (SMTP) + postgres, redis, rustfs only
```

## Cross-cutting
- Logging: zap with a redaction layer (passwords, secrets, tokens). Recipient emails masked. Request bodies never logged.
- Config: environment variables, and `*_FILE` variables pointing to Docker secrets.
- Health: `/healthz` (process) and `/readyz` (database reachable).
- Dev-only API docs: when `DEV_MODE=true`, the API serves Swagger UI at `/docs` and the spec at `/openapi.yaml`.
- Metrics: Prometheus format from `api` and `worker`.
- Trusted proxy: the client IP header is trusted only for requests from the nginx container network.
