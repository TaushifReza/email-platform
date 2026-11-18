# AGENTS.md: rules for coding agents

Project: multi-tenant email automation platform (templates + Excel/CSV data + per-recipient attachments, sent through each tenant's own SMTP provider).
Stack: Go backend (Gin, pgx, sqlc, goose, zap), PostgreSQL with row-level security, Redis (rate limits and wake-ups only), RustFS via S3 API, Next.js (TypeScript, static export, Tailwind, shadcn/ui).
Source of truth: `docs/srs/` (requirements) and `docs/technical-design/` (design). If they conflict, `docs/srs/09-business-rules.md` and `docs/srs/30-acceptance-criteria.md` win.

## Context discipline (your context window is small)
1. Your task is described in ONE file: `docs/tasks/WPxx.md`. Read that file and only the files it lists. Never load the whole SRS (`docs/SRS_*.md`).
2. Work on ONE sub-task per session. At the start read the last 40 lines of `docs/PROGRESS.md`. At the end append a handoff entry (what is done, what is next, anything surprising).
3. Keep files under 400 lines and functions short. Prefer many small packages and files.
4. Find code with `rg` and read narrow line ranges. Do not paste large files into the conversation.
5. Contracts first: write the Go interface, the SQL queries (sqlc) and the OpenAPI definition before the implementation, and commit them. Later sessions then only need the contract.
6. Commit after every green test run, with the requirement IDs in the message (for example `PRV-04: block private IPs in dialer`).

## Non-negotiable invariants (never violate, never "simplify")
- **Tenant isolation:** every tenant table has `organization_id`, forced RLS and composite foreign keys. Every request and job runs inside `WithOrg(ctx, orgID, ...)`. The organization comes from the session, never from user input.
- **All SQL through sqlc.** No string-built SQL.
- **Secrets:** provider passwords and 2FA seeds are encrypted (AES-256-GCM, HKDF subkeys, row-bound AAD). They are never logged, never returned by any API, never shown in the UI.
- **No `InsecureSkipVerify`**, no plaintext SMTP, no private or internal SMTP hosts, SMTP dialer re-checks the resolved IP on every connect.
- **User templates are never executed by `text/template` or `html/template`.** Use the purpose-built `{{variable}}` renderer. Values are escaped and never re-expanded.
- **Never send twice:** recipient state changes only through guarded updates (`UPDATE ... WHERE state = expected`). Follow `docs/technical-design/03-sending-protocol.md` exactly.
- **GET never unsubscribes anyone.** Only POST does.
- **Never write "delivered", "opened" or "bounced"** anywhere in the MVP. Success is "Sent" (accepted by provider).
- **Audit:** every action listed in `docs/srs/23-audit-requirements.md` writes an audit entry. Audit rows are append-only.

## Process rules
- Do not invent business rules. If the SRS is silent, add the question to `docs/OPEN_QUESTIONS.md` and stop that sub-task.
- Tests first for critical modules (crypto, RLS, sending, SSRF dialer, sessions, unsubscribe). Never weaken, skip or delete a test to make it pass. If a test seems wrong, explain in `docs/PROGRESS.md` and ask.
- No new dependency without a note in the pull request (purpose, license, maintenance status). Verify the package exists and is maintained before using it.
- Migrations follow expand-then-contract. Never edit a merged migration.
- `make ci` must pass before a pull request. `make test-module M=<name>` runs one module's tests.
- Pull request description must list deviations from the design and the reason.
- Files under `CODEOWNERS` critical paths need the owner's review. Do not merge them yourself.

## Modes (OpenCode)
- plan (large context): break a work package into sub-tasks, write contracts, review diffs.
- build: implement one sub-task.
- explore: read-only code search and summaries.
