# Test strategy (this is how the owner trusts agent-written code)

| Layer | What | Where |
|---|---|---|
| Unit | Pure logic: state machines, renderer, reply classification, IP classifier, crypto | per package |
| Database | Real PostgreSQL in CI: sqlc queries, guarded transitions, RLS | `*_db_test.go` |
| Catalog guard | Fails if a table with `organization_id` lacks forced RLS, a policy or composite FKs | `internal/platform/db` |
| Behavioral tenancy | User of org A attempts every read and write on org B resources, including files and exports (AC-SEC-01) | `internal/*/tenancy_test.go` |
| Role matrix | Every role x every capability from SRS 6.2 | `internal/identity/permissions_test.go` |
| Fake SMTP | `tools/fakesmtp`: scripted behaviors (accept, each 4xx/5xx class, drop after DATA, delay, TLS variations); API to list what it received | used by sending, providers, templates |
| Crash injection | Kill the worker at each failpoint; restart; check invariants (docs/technical-design/03) on a 5,000-recipient run | `tools/crashsuite` |
| Fuzz | Template renderer, CSV and XLSX parsers, ZIP extraction, email parsing | `*_fuzz_test.go` |
| Security fixtures | Header injection, formula injection, zip bomb, path traversal, SSRF (test resolver returns public then private, IPv6, mapped), planted secrets scanned in logs and responses | per module |
| Frontend | Typecheck, lint, component tests for forms, a few end-to-end flows against staging stack | `frontend/` |

## Make targets
`make ci` (everything), `make test`, `make lint`, `make test-module M=sending`, `make dev-up`, `make dev-down`, `make crash-suite`, `make gen` (sqlc and OpenAPI generation).

## Rule
Every acceptance criterion in `docs/srs/30-acceptance-criteria.md` must map to at least one automated test. Name tests with the criterion id, for example `TestAC_SND_01_NoDuplicateAcrossCrash`.
