## 30. Acceptance criteria

Each criterion is testable. `AC-` identifiers group by area.

### Tenancy and security
- **AC-SEC-01** Given two organizations, when a user of A attempts to read or modify any resource of B by direct identifier or search, then access is denied and nothing of B is revealed. This is verified by automated tests across every resource type, including files and exports.
- **AC-SEC-02** Provider credentials never appear in any API response, UI, log or audit entry, for any role including Platform Admin.
- **AC-SEC-03** A provider hostname resolving to a private or loopback address, or using a port outside the allowed set, is refused. The check is repeated at each connection.
- **AC-SEC-04** A plaintext SMTP connection to any provider is refused.
- **AC-SEC-05** Platform Admin, Owner and Admin accounts cannot complete sign-in without a second factor.
- **AC-SEC-06** A cell value containing `\r\nBcc: x@y.com` produces no extra header or recipient.
- **AC-SEC-07** An exported CSV containing a cell starting with `=` is neutralized so spreadsheet applications do not execute it.
- **AC-SEC-08** A ZIP with path traversal, a nested archive, a password, or an extreme compression ratio is rejected.
- **AC-SEC-09** A template attempting script execution, file access or a `javascript:` link is rejected or neutralized.
- **AC-SEC-10** Operator access to tenant content without an active Owner-issued grant is impossible. Access under a grant is logged and expires on time.

### Providers
- **AC-PRV-01** A provider cannot be used until both connection and test email succeed.
- **AC-PRV-02** Rotating a password updates the credential without recreating the provider or losing history, and the old secret is no longer usable.
- **AC-PRV-03** Missing SPF, DKIM or DMARC records produce a warning but do not block.

### Templates and data
- **AC-TPL-01** `{{name}}` is replaced per recipient in subject and body. An unknown variable with no default blocks validation.
- **AC-TPL-02** A value containing `<b>x</b>` is displayed as literal text, not bold.
- **AC-TPL-03** Editing a template after launch does not change a running or finished campaign.
- **AC-DAT-01** A 10,000-row file imports with progress and without blocking the interface. A 10,001-row file is refused.
- **AC-DAT-02** Duplicates are detected on the chosen key, only the first is kept, and the warning lists the dropped rows.
- **AC-DAT-03** A file with 50 invalid rows out of 10,000 yields a downloadable error report and lets the user proceed with the valid 9,950.
- **AC-DAT-04** Columns with headers equal to variable names are auto-mapped.
- **AC-DAT-05** A suppressed address is never sent within the entry's scope, including one added to suppression after launch.
- **AC-SUP-01** An address with a `non_transactional` entry is skipped in marketing and announcement campaigns but receives transactional campaigns. An address with an `all` entry is skipped in every campaign type. An unsubscribe link creates only a `non_transactional` entry.

### Attachments
- **AC-ATT-01** Given a ZIP and an attachment column, each recipient gets exactly the files named for them. A recipient with one missing file receives nothing and is Skipped (MISSING_ATTACHMENT).
- **AC-ATT-02** Same-name files in different ZIP folders are flagged at upload and affected recipients are skipped.
- **AC-ATT-03** An attachment over 5 MB, an email over 10 MB total, a disallowed type, or a sixth attachment is refused or skipped as specified.
- **AC-ATT-04** Attachments are not retrievable by public URL and each access is logged.

### Campaign lifecycle
- **AC-CMP-01** A marketing campaign without an unsubscribe link cannot launch. A missing postal address only warns.
- **AC-CMP-02** With approval required, the creator cannot approve their own campaign. Editing after approval invalidates the approval.
- **AC-CMP-03** Launch records the snapshot. Later changes to template, data, attachments or provider do not alter the running campaign.
- **AC-CMP-04** Pause survives an application restart. Resume continues without re-sending accepted recipients.
- **AC-CMP-05** Cancel stops all further sends. Already accepted recipients remain Accepted. Remaining recipients become Cancelled.
- **AC-CMP-06** A scheduled campaign still unapproved at its start time does not send and shows Needs Attention.
- **AC-CMP-07** Changing credentials of the provider of a scheduled campaign moves it to Needs Attention and notifies the Owner.

### Sending and reliability
- **AC-SND-01** For a 5,000-recipient campaign, each recipient has exactly one outcome and no recipient is sent twice, including across a forced application crash and restart during sending.
- **AC-SND-02** Killing the worker after provider acceptance but before recording leaves that recipient Uncertain, never automatically resent.
- **AC-SND-03** Sending never exceeds the configured rate or daily limit. A campaign larger than the daily limit continues the next day automatically.
- **AC-SND-04** An authentication failure pauses the campaign immediately and notifies the creator and Owner by email.
- **AC-SND-05** Temporary failures are retried according to the policy. Permanent failures are not retried.
- **AC-SND-06** Two campaigns on one provider run one after the other. Campaigns on different providers run concurrently.
- **AC-SND-07** A missed scheduled start within 1 hour sends late. Beyond 1 hour it becomes Needs Attention.

### Tracking, reporting, notifications, audit
- **AC-TRK-01** The UI never displays the words "delivered", "opened" or "bounced" for MVP results.
- **AC-RPT-01** The campaign report totals equal the sum of the per-recipient statuses. CSV and error exports are available and formula-safe.
- **AC-RPT-02** After purge the report shows aggregate totals and no recipient-level rows.
- **AC-NTF-01** Each event in NTF-02 sends an email via the system sender to the specified recipients.
- **AC-AUD-01** Every action in AUD-01 appears in the audit log with actor, time and target. Tenant users cannot edit or delete entries.
- **AC-UNS-01** Following a valid unsubscribe link adds the address to that organization's suppression list immediately. The one-click request does the same. A tampered link does nothing.

### Retention and operations
- **AC-RET-01** Recipient data and attachments are purged 90 days after completion by default, and by the organization's configured period otherwise. Aggregates and audit entries remain.
- **AC-RET-02** After a backup restore, previously purged data is not exposed.
- **AC-OPS-01** Storage quota is enforced. Exceeding it refuses uploads but does not stop running campaigns.
- **AC-API-01** Every operation reachable from the UI is invoked through the same service layer, with the same validation, approvals and quotas.
