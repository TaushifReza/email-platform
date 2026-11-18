## 26. API requirements

- **API-01 [M·C]** **No public API in the MVP.** All business operations (create campaign, import data, validate, launch, pause, cancel, read status and results, manage providers and templates) MUST be implemented as a **reusable service layer** so that the web interface and a future API call **the same logic and obey the same rules** (approvals, quotas, caps, suppression, tenant scoping, audit).
- **API-02 [P·C]** Future public API: **organization-scoped API keys with limited permissions** (for example create and launch campaigns but not manage users), **idempotency keys** on launch, **status and result retrieval**, **webhooks** for completion events, and rate limits.
- **API-03 [P·C]** Likely first consumers are the owner's other applications that will trigger campaigns, check status and retrieve results. No specific integration is required now.
