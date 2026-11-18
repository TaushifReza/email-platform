# Implementation plan (4 weeks; week 4 is buffer and pilot preparation)

## Working with local models
- Planning model (131k) breaks each work package into sub-tasks and writes contracts. Build models (64k) implement one sub-task per session. Each sub-task should be about 400 lines of new code plus tests at most.
- After WP01 measure real throughput (sub-tasks finished per day, rework rate) and adjust the schedule at Gate 1.
- Local 24B to 30B models are usually good at CRUD, UI and glue code and weaker at concurrency, cryptography and security-sensitive code. For the critical modules (crypto, RLS, sessions and 2FA, SSRF dialer, sending protocol, unsubscribe) the owner should either have a stronger model write or at least review the code, and must read it personally. The test suite is the guardrail.

## Work packages and order
| Week | WP | Scope |
|---|---|---|
| 1 | WP01 | Repo, tooling, CI, Compose, config, redaction logging, OpenAPI scaffold, Next.js shell, fake SMTP |
| 1 | WP02 | Tenancy, identity, RLS framework, 2FA, roles, audit writer, system-email outbox |
| 1 to 2 | WP03 | Crypto, providers, SSRF dialer, verification |
| **Gate 1** | | Security-critical review |
| 2 | WP04 | Templates |
| 2 | WP05 | Imports (with chunked upload service) |
| 2 to 3 | WP06 | Attachments |
| 2 to 3 | WP07 | Campaigns |
| 2 to 3 | WP08 | Sending pipeline (starts when WP02 schema exists) |
| **Gate 2** | | Click-through of primary workflow on staging |
| 3 | WP09 | Suppression and unsubscribe |
| 3 | WP10 | Reports, notifications, audit viewer |
| 3 to 4 | WP11 | Operations: deploy, backups, alerts |
| 4 | | Buffer, fixes, Gate 3, pilot |
| **Gate 3** | | Pilot go/no-go |

## Go or no-go for the pilot
All `AC-SEC-*` and `AC-SND-*` tests green in CI; restore test passed; staging forced-crash test passed; a real provider verified end to end. If anything is missing the date slips. Safety items are never cut.

## Cut order if behind
1 audit viewer screen (events still recorded); 2 saved mappings and MX check; 3 inline images; 4 scheduled sending (immediate only); 5 organization history views and CSV polish.
Never cut: tenant isolation, credential protection, never-send-twice, SSRF protection, unsubscribe correctness, backups.

## M1 (after pilot)
Approval workflow, support-access grants, retention purge job, labels, sending windows, URL-per-row attachments, SPF/DKIM/DMARC warnings, provider usage view, spam-wording check, audit export, breached-password lookup, full notification set.
