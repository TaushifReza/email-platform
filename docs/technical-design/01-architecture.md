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
- Metrics: Prometheus format from `api` and `worker`.
- Trusted proxy: the client IP header is trusted only for requests from the nginx container network.
