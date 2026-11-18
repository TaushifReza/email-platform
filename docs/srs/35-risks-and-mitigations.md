## 35. Risks and mitigations

| ID | Risk | Impact | Mitigation |
|---|---|---|---|
| RISK-01 | **Duplicate sends** after crashes or retries | Duplicate tickets or invitations, embarrassment, trust loss | Unique send record before handoff (SND-06), Uncertain state (BR-03), acceptance tests with crash injection (AC-SND-01, AC-SND-02) |
| RISK-02 | Provider daily caps (for example Google Workspace or Microsoft 365 accounts) make large campaigns take days | Late invitations, user frustration | Per-provider daily limits, multi-day spread, launch-time estimate (SND-03), recommend API-type providers post-MVP |
| RISK-03 | Provider policy volatility: app-password availability and SMTP authentication can be restricted by Google or Microsoft administrators | Providers stop working unexpectedly | Verification before use, clear error messages, OAuth and API providers post-MVP |
| RISK-04 | **No malware scanning** in the MVP. Tenants' files go out under their identity | Reputation damage, abuse | Allow-list, content-type checks, size limits, suspension tools, scanning is the first security addition (**accepted risk**) |
| RISK-05 | SSRF through user-supplied SMTP hosts | Internal network scanning or access | PRV-04 enforced at configuration and every connection |
| RISK-06 | Tenant isolation defect in a shared database | Cross-tenant data leak | Central scoping, automated cross-tenant tests, per-tenant file separation |
| RISK-07 | Emails sent with missing personalization or missing attachments | Reputational harm | Required flags, skip rules, preview, validation, no partial emails |
| RISK-08 | Users misread "Sent" as "delivered" | False confidence. Bounces unnoticed | BR-06, TRK-02 messaging in the UI |
| RISK-09 | Marketing mail without opt-out, or sent to unsuitable lists | Legal and reputation risk | Unsubscribe hard block, declaration, tenant responsibility, suspension, ToS before commercialization |
| RISK-10 | Suppression blocks legitimate transactional mail | Missed tickets | Scoped suppression (SUP-05): unsubscribes affect only marketing and announcement mail |
| RISK-11 | Personal data retained too long or surviving backups | Privacy exposure | Automatic purge, deletion on request, re-apply purges after restore |
| RISK-12 | Attachment storage growth | Disk exhaustion | Per-organization quota, retention purge, alerts at 80% |
| RISK-13 | System sender is a single point of failure for notifications and invitations | Silent failures | Monitoring, in-app notifications in addition to email |
| RISK-14 | Privileged account takeover | Mass abuse of tenant providers | Mandatory 2FA, audit, rate limits, suspension |
| RISK-15 | Public unsubscribe endpoint abused | Mass unsubscribing or enumeration | Signed tokens, rate limits, no data disclosure |
| RISK-16 | Spreadsheet quirks (leading zeros, mixed types, encodings) | Wrong personalization | Warnings, previews, validation (EDGE-05 to EDGE-08) |
| RISK-17 | Scope growth toward a marketing platform | Delays | Non-goals (section 4), post-MVP list (section 32) |
| RISK-18 | Missed scheduled sends after downtime for time-sensitive mail | Late tickets | 1-hour grace, Needs Attention state, notifications |
