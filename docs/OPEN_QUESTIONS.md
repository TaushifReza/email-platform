# Open questions (owner decides; agents add new ones here and stop that sub-task)

| ID | Question | Status / default in use |
|---|---|---|
| OQ-01 | Second alert channel (Telegram, Slack, Discord, other) | Open. Email only until named |
| OQ-02 | Typical attachment size (sizing of storage) | Open. 20 GB per-organization quota default stands |
| OQ-03 | Cloudflare plan on the domain (affects only the upload cap; chunking makes it irrelevant) | Open |
| OQ-04 | Default per-provider rate limit and daily limit when the user does not know | Default 20 per minute, 500 per day |
| OQ-05 | Attachment required when an attachment cell is empty (EDGE-37) | Default: required in column mode |
| OQ-06 | Internationalized (non-ASCII) email addresses | Rejected in MVP |
| OQ-07 | Organization offboarding process (export format, grace period) | Deferred to M1 by owner (2026-11-22) |
| OQ-08 | Operator emergency-access path | Deferred to M1 by owner (2026-11-22). No emergency path in MVP. |
| OQ-09 | Legal documents (ToS, privacy notice, data processing statement) | Post-MVP, before commercialization |
| OQ-10 | Default caps for commercial tenants | Confirmed by owner (2026-11-22): caps off by default internally; placeholder 500/day and 1,000/campaign at commercialization. |
| OQ-11 | Local model throughput and rework rate | Measure after WP01 |
| OQ-12 | How should the pre-auth login lookup find a user by email when forced RLS needs `app.org_id` before authentication? | Resolved (WP02.3): `SECURITY DEFINER` function `auth_lookup_user` owned by `email_system`, executable by `email_app`, returning only login fields. |
| OQ-13 | Lockout threshold and duration (proposed 5 failures / 15 minutes) and Redis outage behavior (proposed fail closed) for login | Confirmed by owner (2026-11-22): keep the same defaults for tenant logins. Not applicable to operators; they authenticate through Cloudflare Access and are allowlisted. |
| OQ-15 | System sender SMTP may use explicit `SMTP_TLS_MODE=none` for Mailpit dev/staging; production requires starttls/implicit | Open (WP02.6) |

| OQ-16 | Operator authentication and operator org-management model | Resolved by owner (2026-11-22): Cloudflare Access is the operator identity provider; the app stores no operator password or TOTP seed. First operator is created with the `operator create` CLI. Include ADM-01 org create/suspend/reinstate/limits, AUD-03 operator audit and a separate admin console now. Defer ADM-02 abuse signals to WP11, support grants and emergency access (OQ-08) to M1, and keep org offboarding (OQ-07) at M1. Implemented in WP02-operator. |
| OQ-17 | Production Cloudflare Access verification: trust the `Cf-Access-Authenticated-User-Email` header only from the tunnel, or verify the CF JWT (team domain + audience) in the app | Open for staging/production. Dev uses the header from trusted proxy CIDRs only. Header trust must be restricted to the authenticated tunnel before any non-local deployment. |
