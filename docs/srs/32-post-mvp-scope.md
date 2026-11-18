## 32. Post-MVP scope

Ordered roughly by expected value. Do not build in the MVP.

1. Bounce handling (reading a bounce mailbox or provider webhooks) with automatic suppression of hard bounces, and delivered or bounced states.
2. Public API: organization-scoped keys, idempotency, status retrieval, webhooks.
3. Open and click tracking (with a platform tracking domain).
4. OAuth for Gmail and Microsoft 365, provider-specific APIs (SES, SendGrid, Mailgun), automatic provider failover.
5. Malware scanning for uploads.
6. Pattern-based attachment matching, generated per-recipient documents.
7. Presigned-link delivery of large files and images.
8. Template conditionals and loops, other data sources (Google Sheets, database, API), .xls and TSV.
9. Recurring campaigns and blackout periods.
10. Single sign-on, per-user notification preferences.
11. Billing, plans, usage metering, Terms of Service, privacy notice, data processing statement.
12. First-campaign review, tiered tenant trust levels.
