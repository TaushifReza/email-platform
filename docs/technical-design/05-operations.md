# Operations

## Environments
Local dev (Compose: Postgres, Redis, Mailpit, S3-compatible store). CI (GitHub Actions with Postgres and the fake SMTP server). **Staging** (second Compose stack, own database and bucket, Mailpit instead of providers). Production.

## Production containers
`web` (static files), `api` (~512 MB), `worker` (~1.5 GB), `migrate` (one-shot), `postgres` (dedicated, ~2 GB), `redis` (dedicated, ~128 MB). Existing nginx routes `app.<domain>` and `admin.<domain>`; a **dedicated Cloudflare Tunnel** connects to nginx. Networks: `edge` (nginx, web, api) and internal `backend` (postgres, redis, api, worker). Database is never published on a host port. Dedicated RustFS bucket and access key limited to that bucket.

## Cloudflare
Tunnel for both hostnames; Access in front of `admin.<domain>`; WAF managed rules; rate-limit rules on login and reset paths; cache static assets, bypass `/api/*` and `/u/*`. Trust the client IP header only from nginx. SPF, DKIM, DMARC for the system sender domain. Request bodies are capped at 100 MB on Free and Pro, hence chunked uploads.

## Backups
PostgreSQL: nightly full plus continuous WAL archiving, encrypted, stored on the second server. Master key backed up separately. Monthly automated restore into a scratch database with integrity checks (and re-application of purge records). Object-store backup optional in M0. Targets: RPO minutes, RTO hours.

## Monitoring and alerts
Prometheus-style metrics from api and worker. Alerts by email (system sender) **and a second channel (open: owner to name it)**: worker down or no progress while campaigns are queued; **any Uncertain recipient created**; provider `needs_attention`; disk above 80%; backup failed or older than 36 h; error-rate or login-failure spikes; DNS or certificate problems for the system sender.

## CI/CD
Pull request: lint, unit and database tests (RLS guard, tenant tests), sending crash suite, frontend typecheck and build, security scans; nothing merges red. Merge to main: build images, push to private GHCR with commit tag, auto-deploy to staging. Production: manual approval; script pulls exact tags, runs `migrate`, restarts `api` then `worker` gracefully. Rollback: redeploy previous tag. Migrations are expand-then-contract.

## First deployment checklist
Domain and tunnel hostnames; Access policy on admin; system sender mailbox with SPF/DKIM/DMARC; master key generated and backed up; bucket and key created; backup job and restore test passing; staging campaign with forced crash; production smoke test with about 10 recipients, all owner-controlled addresses.

Dev stack API documentation: `http://localhost:8081/api/docs` (enabled by `DEV_MODE=true`).
