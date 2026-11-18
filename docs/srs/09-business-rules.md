## 9. Business rules

These rules MUST be enforced by the application regardless of interface.

- **BR-01** All data access is scoped to the caller's organization. Cross-organization access is impossible without a support grant (and then read-only).
- **BR-02** A recipient who has been **Accepted** by the provider for a campaign MUST NEVER be sent the same email again by any automatic process, restart or retry.
- **BR-03** If the outcome of a send attempt cannot be determined (crash or loss after handoff but before recording), the recipient MUST be marked **Uncertain** and MUST NOT be auto-retried.
- **BR-04** A campaign MUST NOT start sending unless: its provider is verified and active, its template and data have passed validation, required approval is recorded, and the sender declaration has been made.
- **BR-05** A recipient with a **missing required variable** (no default), a **missing or ambiguous attachment**, an **invalid address**, a **duplicate** (under the chosen key), or a **suppressed address** MUST NOT be sent an email. They are marked **Skipped** with a reason. A recipient is never sent a partial email (for example, with only some of their attachments).
- **BR-06** The UI and reports MUST NOT describe an email as **delivered**, **received** or **opened**. The success state is **Accepted by provider** (shown to users as "Sent").
- **BR-07** Any change to template, recipient data, attachments, provider or schedule after approval **invalidates the approval**. Approvals do not expire otherwise. The creator cannot approve their own campaign.
- **BR-08** At launch the platform freezes a **snapshot** of the template version, recipient data, attachments and chosen provider. Later edits to the originals do not affect that campaign.
- **BR-09** Campaigns on the **same provider are queued**, not run simultaneously. Campaigns on different providers may run in parallel.
- **BR-10** Sending never exceeds a provider's daily limit or rate limit. A campaign larger than the remaining daily allowance continues on following days.
- **BR-11** Provider credentials are never visible, retrievable or logged.
- **BR-12** Spreadsheet and template data is inserted into emails as **escaped text**. Data cannot inject HTML, headers or recipients.
- **BR-13** Suppression is absolute within its scope (`non_transactional` or `all`, see SUP-05). An unsubscribe takes effect immediately.
- **BR-14** Marketing-type campaigns cannot launch without an unsubscribe link. A missing postal address in the footer produces a warning only.
- **BR-15** Test sends use the real provider and count against the provider's daily and rate limits.
- **BR-16** Sending caps and storage quotas set by the Platform Admin are enforced at validation and at launch.
- **BR-17** A cancelled campaign stops all further sending. Messages already accepted by the provider cannot be recalled.
- **BR-18** Recipient data and attachments are deleted after the retention period (RET-01). Aggregate campaign totals survive.
- **BR-19** Every action listed in section 23 is audit-logged and audit logs cannot be edited or deleted by tenant users.
- **BR-20** When a rule produces a skipped or failed recipient, the campaign continues for others. A single bad row never aborts a campaign unless the user chose to block on errors.

---
