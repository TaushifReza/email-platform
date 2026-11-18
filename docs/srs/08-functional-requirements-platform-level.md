## 8. Functional requirements (platform-level)

Requirements specific to templates, data, attachments, campaigns, sending, tracking, reporting, notifications and audit are in sections 12 to 23. This section covers accounts, providers, suppression and system functions.

### 8.1 Accounts and authentication

- **AUTH-01 [M·C]** Users sign in with **email and password**.
- **AUTH-02 [M·C]** **Two-factor authentication** with an authenticator app (time-based codes) is **mandatory** for Platform Admin, Owner and Admin accounts. It MAY be optional for other roles. Recovery codes MUST be provided.
- **AUTH-03 [M·C]** Accounts are created by **invitation**: an email containing an expiring, single-use link. Suggested invitation expiry: 72 hours. (Expiry A.)
- **AUTH-04 [M·C]** Password rules, lockout after repeated failed sign-ins, rate limiting of sign-in attempts, and session expiry follow current industry best practice. Suggested baseline: minimum 12 characters, checked against known-breached passwords. (Values A.)
- **AUTH-05 [M·C]** Password reset by emailed expiring single-use link (sent by the system sender, PRV-12). Reset invalidates existing sessions.
- **AUTH-06 [P·C]** Single sign-on (Google, Microsoft) is post-MVP.

### 8.2 Users

- **USR-01 [M·C]** Owner and Admin can invite, deactivate and change roles of users in their organization, within the limits of the permission matrix. An organization MUST always have at least one active Owner. (Last-Owner rule A.)
- **USR-02 [M·C]** Deactivating a user does not delete or alter campaigns they created. The creator reference is preserved.
- **USR-03 [M·C]** A role change or deactivation takes effect immediately for new actions. It does not interrupt a running campaign.

### 8.3 Email providers

- **PRV-01 [M·C]** An organization can configure **several providers**. Exactly one is marked **default**. The campaign launch screen shows the provider and requires the user to **confirm** it.
- **PRV-02 [M·C]** MVP provider types: **generic SMTP** (host, port, username, password, TLS mode), including Gmail and Microsoft 365 through SMTP with **app passwords**. OAuth sign-in, provider-specific APIs (SES, SendGrid, Mailgun) and automatic failover are post-MVP **P**, but the provider layer MUST be an extensible abstraction.
- **PRV-03 [M·C]** **Encrypted connections are required.** The platform MUST use TLS or STARTTLS and MUST refuse plaintext SMTP.
- **PRV-04 [M·C]** **Server address safety:** hostnames that resolve to private, loopback, link-local or otherwise internal addresses (including the platform's own infrastructure) MUST be refused, checked at configuration time **and again at every connection**. Ports are restricted to 25, 465, 587 and 2525. No on-premise or internal mail-server exceptions. The check MUST be resistant to DNS rebinding.
- **PRV-05 [M·C]** **Credentials are write-only and encrypted.** They are entered once, stored encrypted with a key held outside the data store, never displayed, never returned by any screen or API, never logged, and never visible to any role including the Platform Admin. They are decrypted only at the moment of connection. The password can be **rotated** without recreating the provider.
- **PRV-06 [M·C]** **Verification before activation** requires both: (a) a successful connection and authentication check and (b) a **test email** delivered to an address the user specifies and accepted by the provider. A provider that has not been verified cannot be used for a campaign.
- **PRV-07 [M·C]** The **From address is fixed per provider** and cannot be typed per campaign. The **sender display name** defaults to the organization's default sender name. A per-campaign override of the display name only is allowed. (Override A.) **Reply-To** may differ and is set per campaign, defaulting to the organization's reply-to.
- **PRV-08 [S·C]** On provider verification and on request, the platform checks the From domain's **SPF, DKIM and DMARC** DNS records and **warns** (never blocks) when missing or misconfigured.
- **PRV-09 [M·C]** Each provider has a **daily sending limit** and a **rate limit** (messages per minute and per hour), entered by Owner or Admin. Defaults if unknown: 500 per day and 20 per minute. (Defaults A.) See SND-02 to SND-05.
- **PRV-10 [M·C]** A provider can be **deactivated**. Deactivating or changing credentials of a provider used by a scheduled campaign moves that campaign to **Needs Attention** (SCH-08).
- **PRV-11 [M·C]** If a provider's credentials fail during sending, the campaign is automatically **paused** and notified (RTY-08).
- **PRV-12 [M·C]** The platform has **one operator-owned system sender** used only for invitations, password resets, alerts and notifications. It is a deployment dependency, separate from tenant providers, because a failing tenant provider cannot report its own failure. Its provider is an open decision (OD-03).

### 8.4 Suppression list and unsubscribe

- **SUP-01 [M·C]** Each organization has its own **suppression list**. It applies **only** to that organization. There is **no platform-wide list** in the MVP.
- **SUP-02 [M·C]** Entries are added by: recipient unsubscribe (UNS-03), manual entry, and CSV upload. Each entry records the address, source, date and, for unsubscribes, the originating campaign. Automatic addition from bounces is post-MVP.
- **SUP-03 [M·C]** Suppression is **absolute**: no one-off override is possible. Only **Owner or Admin** can remove an entry, and removal is audited.
- **SUP-04 [M·C]** The suppression check applies at **import validation** and again **immediately before each send** (EDGE-09), using the scope rules in SUP-05. A suppressed recipient is marked **Skipped (SUPPRESSED)**.
- **SUP-05 [M·C]** Every suppression entry has a **scope** (resolved decision OD-07):
  - `non_transactional`: blocks **marketing and announcement** campaigns only. A recipient who unsubscribes through the unsubscribe link receives this scope, so they still receive their tickets and other transactional email.
  - `all`: blocks **every campaign type, including transactional**. Used for a recipient's explicit request to receive no email at all and for known-bad addresses. It can only be created by Owner or Admin (manually or by CSV upload).
  - Within its scope, suppression is absolute (SUP-03). When both scopes exist for an address, `all` applies.
- **SUP-06 [M·A]** Manual and CSV additions by users with permission choose the scope. Default when unspecified: `all`. The unsubscribe page offers only the `non_transactional` opt-out.
- **UNS-01 [M·C]** The platform provides a **public unsubscribe page** and the standard **one-click unsubscribe** mechanism (`List-Unsubscribe` and `List-Unsubscribe-Post` headers). This is the **only** public-facing function in the MVP.
- **UNS-02 [M·C]** Each email in a **marketing** campaign carries a per-recipient unsubscribe link and the one-click headers. For **announcement** campaigns this is optional per campaign. For **transactional** campaigns it is not required. (Announcement optionality A.)
- **UNS-03 [M·C]** Unsubscribing adds the address to that **organization's** suppression list with scope `non_transactional`, effective immediately. The operation is idempotent.
- **UNS-04 [M·C]** Unsubscribe links MUST use unguessable signed tokens that cannot be altered to unsubscribe other addresses, MUST be rate limited, MUST NOT reveal recipient or campaign details beyond what is needed, and MUST remain valid for at least as long as the suppression data is retained.
- **UNS-05 [M·C]** The template variable `{{unsubscribe_url}}` is system-provided and always available. (Variable name A.)

### 8.5 Platform administration

- **ADM-01 [M·C]** The Platform Admin can create, suspend and reinstate organizations, set caps and storage quotas, and pause an individual campaign for abuse reasons.
- **ADM-02 [M·C]** The Platform Admin sees **aggregate abuse signals** (volume spikes, failure rates, complaint counts where known) **without reading content**.
- **ADM-03 [M·C]** **Support access to tenant content** requires a **time-limited grant issued by the organization's Owner** (suggested default 24 hours, read-only, auto-expiring, revocable). Every access under a grant is audit-logged and the Owner is notified. An emergency path for security or abuse incidents MAY exist, but any use must be logged and the Owner notified afterward. (Emergency path and read-only scope A.)
- **ADM-04 [M·C]** Each organization must declare, per campaign, **"I have the right to send this email to these recipients"** before launch. (Declaration is stored with the launch audit entry.)
