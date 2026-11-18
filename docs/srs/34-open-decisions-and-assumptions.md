## 34. Open decisions and assumptions

### 34.1 Open decisions

| ID | Decision | Recommendation or note | Needed by |
|---|---|---|---|
| OD-01 | Default per-organization storage quota | **RESOLVED 2026-11-17:** 20 GB default per organization, adjustable per organization by the operator. Typical attachment size is still useful for infrastructure sizing | Closed |
| OD-02 | Exact RPO and RTO | Target about 99% availability. Choose after infrastructure is known | Deployment planning |
| OD-03 | Which account and provider is the operator's system sender | Separate from tenant providers. Needs its own SPF, DKIM, DMARC | Before implementation |
| OD-04 | Different retry numbers or provider-failure thresholds | Current: 5 attempts over about 24 hours, thresholds in RTY-04 to RTY-08 | SRS review |
| OD-05 | Legal documents: Terms of Service, privacy notice, data processing statement | Needed before commercialization | Before commercial launch |
| OD-06 | Default daily and per-campaign caps for commercial tenants | Placeholder 500 per day and 1,000 per campaign | Before commercial launch |
| OD-07 | Should suppression be scoped by campaign type? | **RESOLVED 2026-11-17:** yes. Unsubscribes use scope `non_transactional`. A separate `all` scope blocks every type. See SUP-05 | Closed |
| OD-08 | Store suppression entries as plain addresses or hashed | Hashing reduces exposure but prevents operator or user review of the list | SRS review |
| OD-09 | Default provider rate limit values and invitation expiry | Suggested 20 per minute, 500 per day, 72 hours | SRS review |
| OD-10 | Whether campaign **attachment** is required or optional for rows with an empty attachment cell (EDGE-37) | Default is required in column mode | SRS review |
| OD-11 | Support for internationalized (non-ASCII) email addresses | Rejected in the MVP | Before commercial launch |
| OD-12 | Organization offboarding: export format, grace period, deletion confirmation | Operator procedure | Before first offboarding |
| OD-13 | Whether the operator emergency-access path (ADM-03) is wanted at all | Recommended only for security or abuse incidents, always notified | SRS review |

### 34.2 Assumptions (not explicitly confirmed)

Treat as defaults. Ask the client before changing them.

| ID | Assumption |
|---|---|
| A-01 | Approvers are Reviewer, Admin and Owner, never the creator |
| A-02 | Retry policy: 5 attempts over about 24 hours, increasing delays, permanent failures never retried |
| A-03 | Pause triggers: authentication failure, 5 consecutive connection failures, more than 20% failures in the last 50 attempts (minimum 20 attempts) |
| A-04 | Organization campaign history and a provider-usage view are included |
| A-05 | A system sender is a required deployment dependency |
| A-06 | The only public endpoint in the MVP is the unsubscribe handler |
| A-07 | Default retention is 90 days for recipient data and attachments, audit logs at least 1 year |
| A-08 | Pattern-based attachment matching is not in the MVP |
| A-09 | Campaign names, labels, saved column mappings, and duplicate-campaign semantics as stated |
| A-10 | Variable names are case-insensitive, letters digits underscore. `unsubscribe_url` is reserved |
| A-11 | Platform-supplied sender display-name override per campaign is allowed, From address never |
| A-12 | Day boundary for daily limits uses the organization's time zone |
| A-13 | Specific permission matrix cells marked (A) in section 6.2 |
| A-14 | Retention range 7 to 365 days |
| A-15 | Needs Attention can be resolved by Owner, Admin or the creator |

### 34.3 Confirmed requirements summary

Everything tagged **C** in this document was explicitly accepted by the client in the requirements interview. Highlights: internal-first multi-tenant use, operator-created organizations, single-organization accounts, the six roles, optional second-person approval, SMTP with app passwords and verification by test email, fixed From, TLS required, write-only credentials, HTML plus text templates with simple variables, escaped data, Excel and CSV with per-row error reports, ZIP attachments up to 5 per recipient with skip-on-missing, allow-list file types, no malware scanning in the MVP, honest "Accepted" status, no bounce handling in the MVP, unsubscribe in the MVP, never-send-twice with Uncertain handling, all 18 failure scenarios, 90-day retention, mandatory 2FA for privileged roles, support access only by Owner grant, and no public API in the MVP.

### 34.4 Resolved decisions log

| Date | Decision | Outcome |
|---|---|---|
| 2026-11-17 | OD-07 suppression scope | Scoped suppression adopted (SUP-05) |
| 2026-11-17 | OD-01 storage quota | 20 GB per organization default, operator-adjustable |
| 2026-11-17 | Analyst assumptions (A-01 to A-15 and inline **A** tags) | Accepted as working defaults, changeable on request |
| 2026-11-17 | Malware scanning | Post-MVP, risk RISK-04 accepted |
| 2026-11-17 | SRS draft 0.1 | Reviewed, approved without changes |
