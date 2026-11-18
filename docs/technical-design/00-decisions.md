# Decision log (all approved by the owner)

| ID | Decision |
|---|---|
| D-01 | Backend in **Go** (Gin, zap, config from env and Docker secret files). Frontend **Next.js in TypeScript as a static export** (no Node server in production), **Tailwind CSS**, shadcn/ui, TanStack Query, React Hook Form with Zod, TipTap editor |
| D-02 | **PostgreSQL** (new stable version, dedicated database and roles for this app) with **pgx and sqlc**, migrations with **goose**. Not Gorm |
| D-03 | **Row-level security** on every tenant table, plus **composite foreign keys** (organization_id, id) and a CI catalog guard test |
| D-04 | IDs are **app-generated time-ordered UUIDs** |
| D-05 | **Modular monolith**: one codebase, one image, processes `api`, `worker`, `migrate` |
| D-06 | **PostgreSQL is the send queue and the job queue.** Redis is used only for rate limits and wake-up signals. Never for send state |
| D-07 | Attachments in **RustFS** (S3 API) on the second server, per-organization key prefixes, generated object keys, TLS |
| D-08 | **OpenAPI-first**: Go server stubs and the TypeScript client are generated |
| D-09 | Sessions: opaque random token stored hashed, `__Host-` cookie, HttpOnly, Secure, SameSite, CSRF header plus Origin check |
| D-10 | Secrets: one master key as Docker secret, **HKDF purpose subkeys**, AES-256-GCM with row-bound associated data, key IDs for rotation |
| D-11 | **No option to skip TLS certificate verification.** STARTTLS or implicit TLS required, TLS 1.2 minimum |
| D-12 | **Chunked, resumable uploads** (Cloudflare limits a request body to 100 MB on Free and Pro) |
| D-13 | Exposure: **dedicated Cloudflare Tunnel** to the existing **nginx**. Operator console on its own hostname behind **Cloudflare Access** |
| D-14 | **Dedicated PostgreSQL and Redis containers** for this app. **Staging** environment with **Mailpit**, no real providers |
| D-15 | Backups: nightly full plus WAL archiving to the second server, encrypted, **monthly automated restore test**. Proposed RPO minutes, RTO hours |
| D-16 | **GitHub** repository, **GitHub Actions** CI, private **GitHub Container Registry**, manual approval for production deploy |
| D-17 | Release split: **M0** internal pilot, **M1** SRS-complete. Plan is **4 weeks**, week 4 is buffer and pilot preparation |
| D-18 | Coding agent: **OpenCode with local Ollama models** (planning model 131k context, build models 64k context). Documents are therefore split into small files |
| D-19 | Sending protocol: two-phase claim, timeout after message body is Uncertain, provider health flag, pause releases provider, provider quota reply stops the day, conservative daily limits advised |
| D-20 | Suppression: plain normalized emails, per-organization, scopes `non_transactional` and `all`. Unsubscribe token is encrypted and self-contained |
| D-21 | Snapshot at launch by immutability (template versions immutable, rows locked, provider `config_version` check) |
| D-22 | System sender: operator-owned Google Workspace SMTP account (dedicated mailbox, app password) |
| D-23 | Library candidates (verify maintenance before use): SMTP client with explicit TLS control, an XLSX reader, an HTML sanitizer, a TOTP library, argon2, a Postgres job library such as River, oapi-codegen |
