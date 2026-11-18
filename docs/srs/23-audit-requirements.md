## 23. Audit requirements

- **AUD-01 [M·C]** The following MUST be recorded: user sign-in, failed sign-in, sign-out, password and 2FA changes, user creation, removal and role change, provider creation, change, verification, rotation and deactivation (never the secret itself), template creation, edit and archive, recipient data import, attachment upload, campaign creation, edit, submission, approval, rejection, launch, pause, resume, cancel, **recipient data export or download**, data deletion and purge, suppression list changes, **uncertain-recipient resolution**, **support access grants and every access under them**, and **operator actions** (organization creation, suspension, limit changes, campaign pauses).
- **AUD-02 [M·C]** Each entry records: UTC timestamp, organization, actor and role, action, target type and identifier, outcome, source IP and client, and non-sensitive details. Entries MUST NOT contain credentials and SHOULD NOT contain recipient addresses (use counts and identifiers).
- **AUD-03 [M·C]** **Who can view:** Owner and Admin see their organization's log. The Platform Admin sees only operator-level logs and the log of access they themselves made.
- **AUD-04 [M·C]** **Integrity and retention:** append-only, not editable or deletable by tenant users, retained at least **1 year** (configurable).
- **AUD-05 [S·A]** Audit log export as CSV.
