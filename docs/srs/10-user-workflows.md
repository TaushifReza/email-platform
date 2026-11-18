## 10. User workflows

### 10.1 Primary workflow (the MVP success workflow)

The MVP is successful when a user completes this end to end **without writing code**: (C)

1. Operator creates the organization and invites the Owner. Owner signs in with two-factor authentication.
2. Owner or Admin **configures and verifies an email provider** (connection check plus test email).
3. Campaign Manager **creates a template** with static content and variables such as `{{name}}`, `{{ticket_number}}`.
4. Campaign Manager **creates a campaign** and **uploads an Excel or CSV** file of recipients.
5. User **maps columns to variables** (auto-matched by identical names).
6. System **validates** data and template and shows errors, warnings and skipped rows.
7. User **uploads attachments** (ZIP or files, and any shared attachment) and the system matches them to rows.
8. User **previews** emails for individual recipients and optionally sends a **test email**.
9. If approval is required, the campaign goes to a different user for **approval**.
10. User **launches now or schedules** it, confirming provider, time zone and the sender declaration.
11. System sends personalized emails through the organization's provider within limits.
12. System records the **outcome for every recipient**.
13. User views **campaign status and results**, downloads reports, and (if needed) **resends failed recipients**.

Headline criterion: *Excel/CSV recipient data + reusable template + personalization + optional per-recipient attachments leads to automated sending through the user's provider with recipient-level status and campaign reporting, reliably, for up to 5,000 recipients and beyond.*

### 10.2 Secondary workflows

| Workflow | Summary |
|---|---|
| **Approval** | Creator submits, a Reviewer, Admin or Owner (not the creator) approves or rejects with a comment. Rejection returns to Draft. Any later edit invalidates approval. |
| **Pause, resume, cancel** | Authorized user pauses a running campaign. Resume continues from where it left off. Cancel stops further sending. |
| **Resend failed** | After completion, user resends to Failed recipients (and resolves Uncertain recipients explicitly). Accepted recipients are never included. |
| **Duplicate campaign** | Copies template reference, settings and mapping into a new Draft. New data and attachments are required. |
| **Provider failure** | Campaign auto-pauses, Owner and creator are notified, an authorized user fixes the provider and resumes. |
| **Unsubscribe** | Recipient opens link, confirms, address enters the organization's suppression list. |
| **Support access** | Owner grants time-limited access, operator investigates, access expires, all access is logged. |
| **Data purge** | After retention expires or on request, recipient data and attachments are deleted, totals remain. |
