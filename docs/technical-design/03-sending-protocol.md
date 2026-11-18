# Sending protocol (the highest-risk part; follow exactly)

## Per-email protocol
| Step | Action | Crash here is safe to retry? |
|---|---|---|
| 1 Claim (one transaction) | Pick a due recipient `FOR UPDATE SKIP LOCKED`. Check campaign running, organization not suspended, suppression (scope by campaign type), provider healthy. Set state Sending, create attempt `claimed`, reserve one unit of provider budget, set lease | Yes, nothing sent |
| 2 Prepare | Render from frozen template version. Fetch attachments. If an attachment is missing: Failed (ATTACHMENT_UNAVAILABLE) | Yes, nothing sent |
| 3 Mark handoff (tiny transaction) | Attempt becomes `handoff_started` | Not safe from here |
| 4 Send | SMTP conversation, read final reply | Ambiguous |
| 5 Record outcome (one transaction) | Accepted, Failed, or back to Queued with `next_attempt_at`. Guarded by attempt id. Adjust budget | Ambiguous until saved |

## Recovery sweeper (at worker start and periodically)
Recipient in Sending with expired lease: attempt `claimed` -> Queued (nothing sent); attempt `handoff_started` with no outcome -> **Uncertain**. Lease is longer than max SMTP time (for example 10 min vs 2 min timeout). Graceful shutdown: stop claiming, finish in-flight (up to about 60 s), exit.

**A timeout after the message body was sent is Uncertain, never a retry.** A timeout before that (connect, auth, MAIL FROM, RCPT) is a temporary failure.

## Reply classification (one table in code, tested)
| Reply | Class | Result |
|---|---|---|
| Success after full message | Accepted | Done |
| 421, 450, 451, 452, enhanced 4.x.x | Temporary | Retry with backoff |
| Connect refused, TLS failure, timeout before body | Temporary connection failure | Retry, counts toward 5 in a row |
| 550, 551, 553 recipient-specific | Permanent recipient | Failed |
| 552, 554 message-specific | Permanent message | Failed with reason |
| 530, 534, 535 | Provider-wide | Pause campaign, provider `needs_attention` |
| Sender blocked, reputation or policy (5.7.x) | Provider-wide | Pause, `needs_attention`, notify |
| Provider daily quota exceeded | Provider quota | Stop for the day, do not fail recipients |
Unknown replies default to the safe class: temporary for recipient-level, never a silent success.

## Retry policy
Up to 5 attempts over about 24 h (delays after first attempt: 5 min, 30 min, 3 h, 12 h; operator-configurable). Then Failed (RETRIES_EXHAUSTED). Retries respect rate limits, daily budget and sending windows.

## Limits
- One dispatcher per provider (PostgreSQL advisory lock). 2 to 3 parallel connections per provider.
- Token bucket for per-minute and per-hour limits, counters persisted in `provider_usage`. On "slow down" replies halve the rate and recover gradually.
- Daily budget reserved at claim, released if the attempt ends before handoff. Exhausted: campaign shows "waiting for tomorrow's allowance", creator notified once. UI advises entering a limit below the provider's real cap (rolling 24 h windows).

## Health, auto-pause, queueing
- Provider status `active|needs_attention|deactivated`. Auth failure, sender block or a tripped trigger sets `needs_attention`; nothing sends through it until an Owner or Admin re-verifies.
- Triggers: authentication failure (immediate), 5 consecutive connection failures, more than 20% failures in the last 50 attempts (evaluated once at least 20 attempts exist). Notify creator and Owner.
- Scheduler tick every 10 to 15 s: promote due Scheduled campaigns, apply the 1-hour grace rule (else Needs Attention), approval check, provider version check, start the next queued campaign when the provider is free. One active campaign per provider, launch order.
- Manual pause releases the provider. Resume returns to the original queue position and never interrupts a campaign already sending.
- Cancel: one transaction moves Queued recipients to Cancelled; in-flight sends finish with their real outcome. Suspension behaves like pause.

## Invariants the tests must prove (AC-SND-01 to AC-SND-07)
1. The fake SMTP server never receives the same recipient's message twice.
2. Every recipient has exactly one final state.
3. Accepted + Uncertain accounts for everything the fake server received; each Uncertain is explained by a killed attempt.
4. Nothing is sent after cancel, pause, suppression or suspension took effect.

## Failpoints (named, enabled only in test builds)
`after_claim`, `after_prepare`, `after_handoff_marked`, `after_body_before_reply`, `after_reply_before_record`.
