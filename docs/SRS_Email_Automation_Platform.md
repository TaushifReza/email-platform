# Software Requirements Specification
## Multi-Tenant Email Automation and Personalized Bulk Email Platform

| | |
|---|---|
| **Document status** | Draft 0.2, reviewed and approved by the client on 2026-11-17 (decisions OD-01 and OD-07 resolved; analyst assumptions accepted as defaults) |
| **Date** | 2026-11-17 |
| **Phase** | Requirements complete for MVP; technology, architecture, schema and API design intentionally NOT decided |
| **Audience** | Product owner (client), and AI coding agents (Claude Code, Codex, OpenCode, others) |

---

## How to read this document (instructions for implementing agents)

1. **This SRS defines behavior, not technology.** No language, framework, database, queue, cloud or library is chosen. Choose the stack in a separate design step and record it separately. Do not infer a stack from this document.
2. **Do not invent business rules.** If a behavior is not specified here, ask the product owner or record it as an open question. Never silently assume an important rule.
3. **Requirement IDs.** Every requirement has an ID such as `TPL-03`. Use IDs in code comments, tests, commit messages and task breakdowns.
4. **Tags on each requirement** use the form `[Priority·Basis]`:
   - Priority: **M** = must have for MVP, **S** = should have (build if capacity allows; the MVP may ship without it), **P** = post-MVP (do not build now, but do not block it architecturally).
   - Basis: **C** = explicitly confirmed by the client, **A** = assumption proposed by the analyst and not yet explicitly confirmed. Treat **A** items as defaults that the client may change. They are listed again in section 34. On 2026-11-17 the client reviewed the SRS, approved it without changes and accepted all **A** items as the working defaults (changeable on request).
5. **Normative words.** MUST, MUST NOT, SHOULD and MAY are used in their usual requirement sense.
6. **Precedence.** Section 9 (Business Rules) and section 30 (Acceptance Criteria) win over any looser wording elsewhere.
7. **Architectural constraints that are requirements, not design choices:** a reusable service layer for all business operations (API-01), tenant scoping enforced centrally (SEC-03), and a "never send the same email twice" guarantee (SND-06).

---

## Table of contents

1. Product overview
2. Problem statement
3. Goals
4. Non-goals
5. Target users
6. User roles
7. Tenant / organization model
8. Functional requirements (platform-level)
9. Business rules
10. User workflows
11. Email lifecycle
12. Template requirements
13. Personalization requirements
14. Recipient-data requirements
15. Attachment requirements
16. Campaign requirements
17. Scheduling requirements
18. Sending requirements
19. Retry and failure requirements
20. Tracking requirements
21. Reporting requirements
22. Notification requirements
23. Audit requirements
24. Security requirements
25. Compliance considerations
26. API requirements
27. Scalability requirements
28. Data-retention requirements
29. Edge cases
30. Acceptance criteria
31. MVP scope
32. Post-MVP scope
33. Future considerations
34. Open decisions and assumptions
35. Risks and mitigations
36. Glossary

---

## 1. Product overview

The product is a **multi-tenant email automation platform** for preparing, personalizing, sending and tracking large numbers of one-to-one emails. A user combines a reusable **template** with **recipient data** from an Excel or CSV file, optionally maps **recipient-specific attachments**, validates everything, previews individual emails, and sends or schedules the **campaign** through the organization's **own external email provider** (SMTP).

Key characteristics:

- **Multi-tenant:** multiple organizations share one deployment with strict data isolation.
- **Provider-independent:** the platform does not run SMTP or mail infrastructure. Each organization connects its own provider. The provider layer must be replaceable and extensible.
- **Template-driven and data-driven:** no custom script per project.
- **Observable:** every recipient has a recorded outcome; every campaign has a report.
- **Internal first:** initial tenants are the owner's own client organizations (colleges, a marketing company, a tech company). Commercialization is possible later, so commercial concerns (billing, plans, legal documents) are designed for but not built.

## 2. Problem statement

The product owner repeatedly writes custom scripts to send the same email body with per-recipient personalization (name, ticket number) and often a different attachment per recipient (for example a ticket PDF), to batches of 150 to 5,000 recipients. Each project needs a new script, new validation, new logging and new status tracking. Ordinary email clients cannot do this at scale, and generic mail-merge or marketing tools handle per-recipient attachments, reuse and per-recipient result tracking poorly.

## 3. Goals

- **G1.** A user completes the full workflow, from data upload to a campaign report, **without writing code**.
- **G2.** Reliable large-volume sending (up to 5,000 recipients per campaign in the MVP, with headroom to 10,000 and no architectural block at 100,000).
- **G3.** **Recipient-specific attachments** matched automatically from imported data.
- **G4.** **Safety against costly mistakes:** no duplicate sends, no emails with missing personalization or missing attachments, no silent failures.
- **G5.** **Reusability:** templates, provider settings, and campaign duplication.
- **G6.** **Tenant isolation, credential protection and auditability** strong enough for commercial use later.
- **G7.** **Honest tracking:** report only what is actually known (see section 20).

## 4. Non-goals

The following are explicitly out of scope for the MVP (client-confirmed):

1. Newsletter or marketing-automation platform: no subscriber lifecycle, drip campaigns or journeys.
2. CRM, lead or pipeline management.
3. Permanent contact database or address book. Recipient data belongs to a campaign.
4. Drag-and-drop visual email or page designer. (A template editor is required.)
5. Inbound email handling.
6. SMS, WhatsApp, push or any non-email channel.
7. The platform's own SMTP or mail-delivery infrastructure. (A single operator-owned **system sender** account for platform notifications is a required dependency; see PRV-12.)
8. Becoming an API-first transactional email service.
9. Advanced marketing automation: behavioral triggers, scoring, funnels.
10. Delivery, bounce, open or click tracking (post-MVP).
11. Public API, API keys, webhooks, billing, plans and subscriptions (post-MVP).

## 5. Target users

| User type | Description | Technical level |
|---|---|---|
| **Platform operator** (the product owner) | Creates and suspends organizations, sets limits, handles abuse and support | Technical |
| **Organization staff** | Staff at colleges, marketing companies and tech companies who send event tickets, graduation invitations, announcements, credentials and occasionally marketing email | Mostly non-technical |
| **Approvers** | Colleagues who review campaigns before launch | Non-technical |
| **Email recipients** | People who receive the emails; interact only with the public unsubscribe page | n/a |

Typical use today: event tickets and graduation invitations for colleges with batches of 150, 500, 1,500 and 5,000, one personalized PDF or image per recipient, with a common body per college.

## 6. User roles

Each user account belongs to **exactly one organization** and has **exactly one role** within it. The Platform Admin is an operator account outside any tenant. (C)

### 6.1 Role definitions

| Role | Purpose |
|---|---|
| **Platform Admin** | Operates the platform: creates, suspends and limits organizations, handles abuse. Has **no default access** to tenant content. |
| **Org Owner** | Full control of one organization, including users, providers, approvals and support-access grants. |
| **Org Admin** | Manages users (except Owners), providers, templates, campaigns and suppression within the organization. |
| **Campaign Manager** | Creates templates and campaigns, imports data, launches (subject to approval) and controls campaigns. |
| **Reviewer / Approver** | Reviews and approves or rejects campaigns. Cannot create or edit them. |
| **Viewer** | Read-only access to campaigns, reports, templates and recipient data. (May be restricted later; C) |

### 6.2 Permission matrix

`✔` allowed, `✘` not allowed, `own` only for items the user created, `grant` only with an active Owner-issued support grant.

| Capability | Platform Admin | Owner | Admin | Campaign Mgr | Reviewer | Viewer |
|---|---|---|---|---|---|---|
| Create, suspend, limit organizations | ✔ | ✘ | ✘ | ✘ | ✘ | ✘ |
| Set org sending caps and storage quota | ✔ | ✘ | ✘ | ✘ | ✘ | ✘ |
| View operator-level audit logs and aggregate abuse signals | ✔ | ✘ | ✘ | ✘ | ✘ | ✘ |
| View tenant content (recipients, emails, attachments) | grant | ✔ | ✔ | ✔ | ✔ | ✔ |
| View provider credentials in plaintext | ✘ never | ✘ never | ✘ never | ✘ never | ✘ never | ✘ never |
| Org settings (name, logo, sender name, reply-to, time zone, approval requirement, retention) | ✘ | ✔ | ✔ (except approval and retention) | ✘ | ✘ | ✘ |
| Manage users and roles | ✘ | ✔ | ✔ (not Owners) | ✘ | ✘ | ✘ |
| Grant time-limited support access | ✘ | ✔ | ✘ | ✘ | ✘ | ✘ |
| Create, edit, verify, rotate, deactivate providers | ✘ | ✔ | ✔ | ✘ | ✘ | ✘ |
| Create and edit templates | ✘ | ✔ | ✔ | ✔ | ✘ | ✘ |
| View templates | ✘ | ✔ | ✔ | ✔ | ✔ | ✔ |
| Create campaigns, import data, upload attachments | ✘ | ✔ | ✔ | ✔ | ✘ | ✘ |
| Launch or schedule a campaign | ✘ | ✔ | ✔ | ✔ | ✘ | ✘ |
| Approve or reject a campaign (never one's own) | ✘ | ✔ | ✔ | ✘ | ✔ | ✘ |
| Pause, resume, cancel campaigns | ✘ (suspend only) | ✔ | ✔ | own | ✘ | ✘ |
| View campaign status and reports | grant | ✔ | ✔ | ✔ | ✔ | ✔ |
| Export reports and error files | grant | ✔ | ✔ | ✔ | ✘ | ✘ |
| Delete a campaign's recipient data | ✘ | ✔ | ✔ | ✘ | ✘ | ✘ |
| Add to suppression list | ✘ | ✔ | ✔ | ✔ | ✘ | ✘ |
| Remove from suppression list | ✘ | ✔ | ✔ | ✘ | ✘ | ✘ |
| View organization audit log | ✘ | ✔ | ✔ | ✘ | ✘ | ✘ |

The matrix is (C) except these items, which are (A): Admin limits on approval and retention settings, Campaign Manager "own" scope for pause and cancel, Reviewer being unable to export, Campaign Manager adding to suppression, Admin being unable to manage Owners, deleting recipient data restricted to Owner and Admin.

## 7. Tenant / organization model

- **ORG-01 [M·C]** An **organization** (tenant) is the unit of data isolation. All data belongs to exactly one organization except operator-level records.
- **ORG-02 [M·C]** Organizations are **created only by the Platform Admin**. There is no self-signup.
- **ORG-03 [M·C]** Organization settings: name, logo, default sender name, default reply-to address, default time zone.
- **ORG-04 [M·C]** Organization settings also include: the approval requirement switch (ORG-06), retention period (RET-02) and daily or campaign caps set by the operator.
- **ORG-05 [M·C]** A user belongs to **one organization only**. The same email address cannot be a user in two organizations. A person who needs access to several organizations uses separate accounts with different email addresses.
- **ORG-06 [M·C]** Each organization can enable **required second-person approval** of campaigns (see CMP-08 to CMP-12). Default: off. (Default value is A.)
- **ORG-07 [M·C]** The Platform Admin can **suspend** an organization. Suspension immediately pauses all of its running campaigns, blocks launching or scheduling new ones, and blocks nothing else (users can still sign in and view data). Reinstating is possible.
- **ORG-08 [S·C]** **Departments, teams or projects** exist as a single flat level of **labels used to organize and filter** templates and campaigns and for reporting. They do **not** control access. No nesting. (C for purpose; flat level A.)
- **ORG-09 [M·C]** **Storage quota** per organization (default 20 GB, adjustable per organization by the Platform Admin; C) covering uploaded attachments and inline images. Uploads beyond the quota are refused with a clear message (EDGE-27).
- **ORG-10 [M·C]** **Sending caps** (emails per day, recipients per campaign) are supported per organization and set by the Platform Admin. **Caps are off by default** while the platform is internal-only. At commercialization they become on by default (suggested starting values: 500 emails per day, 1,000 recipients per campaign) and the operator raises them per organization. (Values A.)
- **ORG-11 [M·A]** Organization offboarding (export and deletion of all organization data) is an operator procedure that must be possible. Details are an open decision (OD-12).

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

## 11. Email lifecycle

### 11.1 Recipient-level states

| State | Meaning | Terminal for this campaign? |
|---|---|---|
| **Pending** | Row imported, not yet queued | No |
| **Queued** | Included in a launched campaign, waiting to send | No |
| **Sending** | Handed to the sending process, outcome not yet known | No |
| **Accepted** | Provider accepted the message (shown as "Sent") | Yes |
| **Failed** | Provider rejected it permanently, or retries exhausted. Reason stored | Yes (may be manually retried) |
| **Skipped** | Removed before sending, with a reason code (see 11.3) | Yes |
| **Uncertain** | Outcome unknown after a crash or loss. Needs user decision | Yes until resolved |
| **Cancelled** | Campaign cancelled before this recipient was sent | Yes |

**Delivered, Bounced, Opened and Clicked are post-MVP states** and MUST NOT appear in the MVP. (C)

### 11.2 Transitions

- Pending to Queued on launch.
- Queued to Sending when a send attempt begins. The attempt MUST be durably recorded **before** the provider handoff (SND-06).
- Sending to Accepted on provider acceptance.
- Sending to Queued (scheduled retry) on a temporary failure while attempts remain.
- Sending to Failed on a permanent failure or exhausted retries.
- Sending to Uncertain when recovery finds an attempt with no recorded outcome (BR-03).
- Queued to Cancelled on campaign cancellation.
- Failed to Queued only by an explicit user "retry" or "resend failed" action.
- Uncertain to Accepted ("mark as sent") or to Queued ("resend") only by explicit user action with a confirmation that warns of duplicate risk. Both are audited. (C for explicit decision; exact UI A.)
- Pending to Skipped during validation or just before send.

### 11.3 Skip and failure reason codes (minimum set)

`INVALID_EMAIL`, `DOMAIN_NOT_FOUND`, `DUPLICATE`, `SUPPRESSED`, `MISSING_REQUIRED_VALUE`, `MISSING_ATTACHMENT`, `AMBIGUOUS_ATTACHMENT`, `ATTACHMENT_TOO_LARGE`, `ATTACHMENT_TYPE_NOT_ALLOWED`, `ATTACHMENT_UNAVAILABLE` (failed at send time), `RECIPIENT_REJECTED` (permanent), `RETRIES_EXHAUSTED`, `PROVIDER_ERROR`, `CAP_EXCEEDED`. (Codes A.)

### 11.4 Campaign-level states (C)

`Draft` then `Validating` then `Ready` then `Pending Approval` (only if required) then `Scheduled` or `Queued` then `Sending` which can be `Paused`, ending in `Completed`, `Completed with Errors`, `Failed` or `Cancelled`.

An additional state, **Needs Attention**, holds a campaign that cannot proceed automatically and needs a human decision (SCH-08, SCH-09, SCH-10). (Needs Attention is C via the failure-scenario table.)

- **Completed:** every recipient is Accepted or Skipped.
- **Completed with Errors:** at least one recipient is Failed or Uncertain.
- **Failed:** the campaign could not run at all, for example validation became invalid at start, or the provider was permanently unusable with no accepted sends. (Definition A.)

## 12. Template requirements

- **TPL-01 [M·C]** Templates support **HTML and plain-text** bodies. The plain-text version is **generated automatically from the HTML**. (An editable override of the plain-text part is **S·A**.)
- **TPL-02 [M·C]** A **rich-text editor** (bold, italics, links, lists, images, basic layout) plus a **raw HTML mode** for advanced users. Complex layouts may require pasting HTML. A drag-and-drop designer is out of scope.
- **TPL-03 [M·C]** Variables use `{{variable_name}}` syntax. **Simple substitution only.** No conditionals, loops, filters or expressions in the MVP (**P**).
- **TPL-04 [M·C]** The **subject** supports variables.
- **TPL-05 [M·C]** Each template has a **variable settings** table: for each variable, an optional **default value** and a **required or optional** flag. Syntax stays `{{name}}`. Defaults live in the settings, not in the placeholder.
- **TPL-06 [M·C]** **Versioning:** every save creates a new version. A campaign records the exact version used and freezes a snapshot at launch (BR-08).
- **TPL-07 [M·C]** Templates are shared across the **whole organization** (not per department). Labels per ORG-08 may be applied. Templates can be **duplicated** and **archived** (archived templates cannot be chosen for new campaigns but remain readable, and do not affect existing campaigns).
- **TPL-08 [M·C]** **Preview** renders the real template with the data of any chosen recipient row (first, any number, random) in a sandboxed view that does not execute scripts.
- **TPL-09 [M·C]** **Test send:** sends one rendered email using a chosen real row to an address the user types, through the real provider. It counts against provider limits (BR-15). The subject is prefixed `[TEST]`. (Prefix A.)
- **TPL-10 [M·C]** Inline images (embedded in the message body) are supported. They count toward the total email size limit and are stored within the organization's storage quota.
- **TPL-11 [M·C]** **Template validation** checks before launch (each is a validation result, with severity):
  - variables used in the template but not mapped to data and with no default: **error**,
  - mapped columns not used by the template: **warning**,
  - malformed or unsafe links (`javascript:`, `data:` and similar blocked): **error**,
  - missing unsubscribe link in a marketing campaign: **error (block)**,
  - missing physical postal address in a marketing campaign: **warning**,
  - **spam-risk wording**: **advisory warning only, never blocking**,
  - subject empty after substitution: **error**, very long subject: **warning**.
- **TPL-12 [M·A]** Variable names: letters, digits and underscore, matched case-insensitively. Reserved system variables: `unsubscribe_url`.
- **TPL-13 [M·A]** Template HTML is sanitized: scripts, event handlers, external active content and forms are removed or rejected.
- **TPL-14 [M·C]** Email addresses in the data column mapped to **recipient address** are not template variables unless the user also maps them as variables (for example `{{email}}`).

## 13. Personalization requirements

- **PER-01 [M·C]** Any mapped column can be used as a variable in subject and body.
- **PER-02 [M·C]** **Missing value rule:** if a value is empty, use the template default if defined. If there is none and the variable is **required**, the recipient is **Skipped (MISSING_REQUIRED_VALUE)** or the whole campaign is **blocked**, depending on a **per-campaign strictness setting**. An empty **optional** variable with no default is rendered as an empty string. (Strictness default: block. Default A.)
- **PER-03 [M·C]** Values are **HTML-escaped** in HTML bodies. Raw HTML from data is not supported. In the plain-text part values are inserted literally, with control characters stripped.
- **PER-04 [M·C]** Values inserted into the **subject** or any header MUST have line breaks and control characters removed (header-injection protection).
- **PER-05 [M·C]** The rendered result per recipient can be **re-created on demand** from the frozen snapshot. Full rendered bodies are **not stored** per recipient.
- **PER-06 [M·A]** Leading and trailing whitespace in values is trimmed. Numeric-looking cells such as ticket numbers keep their displayed text. A warning is shown if leading zeros appear lost (EDGE-05).

## 14. Recipient-data requirements

- **DAT-01 [M·C]** Accepted formats: **Excel (.xlsx) and CSV**. Not in MVP: .xls, TSV, Google Sheets links, API upload, database connections, webhook lists (**P**).
- **DAT-02 [M·C]** Limits: **10,000 rows and about 10 MB per file**. Design target for committed use is 5,000 recipients per campaign. Larger files are refused with a clear message.
- **DAT-03 [M·C]** The **only mandatory column is the recipient email address**. Any other column can be mapped to any variable. There is no fixed schema.
- **DAT-04 [M·C]** **Column mapping:** columns whose headers equal a variable name are **auto-matched**. The user can adjust. Mappings may be saved and reused for later campaigns with the same layout. (Saved mappings A.)
- **DAT-05 [M·C]** **Validation at import** (all required):
  1. email syntax check,
  2. duplicates within the file (see DAT-06),
  3. missing required values for the template,
  4. domain check (the address's domain must exist and be able to receive mail, to catch typos like `gmial.com`; a DNS timeout produces a warning, not an error),
  5. suppression list check,
  6. attachment checks (section 15),
  7. cap checks (ORG-10).
- **DAT-06 [M·C]** **Duplicates:** detection is on by default using the **email address** as the key. The user can choose a different **uniqueness key** (for example email plus ticket number) or switch it off for the campaign. By default the **first occurrence is kept** and later ones are **Skipped (DUPLICATE)**. The warning lists exactly which rows were dropped. Email comparison is case-insensitive.
- **DAT-07 [M·C]** **Invalid rows:** a downloadable **per-row error report** is produced. The user chooses **proceed with valid rows** or fix and re-upload.
- **DAT-08 [M·C]** **Editing:** a user can **edit a single row** inside the platform before launch. After launch only Failed or Pending rows can be edited, with an audit entry. **Bulk editing** is not supported.
- **DAT-09 [M·C]** Recipient data **belongs to the campaign**. There is no permanent contact database. A campaign's data can be deleted on request (RET-03).
- **DAT-10 [M·C]** One email per row (**one-to-one**). No CC, BCC, or several To addresses in the MVP.
- **DAT-11 [M·A]** Excel files with several sheets: the user chooses the sheet. Formula cells use their stored value. Files with macros are refused. CSV encoding is UTF-8 with automatic delimiter detection and byte-order-mark tolerance.
- **DAT-12 [M·A]** Import of large files runs **asynchronously** with visible progress, and the UI paginates data of 10,000 rows. Row values are never trusted as markup (see section 24).

## 15. Attachment requirements

- **ATT-01 [M·C]** Sources: **(a) bulk upload** of a ZIP or multiple files matched to rows (**M**), **(d) one shared attachment** attached to every email (**M**), **(b) URL or storage path per row**, downloaded at send time or import time (**S**). Generated attachments (c) are **P**.
- **ATT-02 [M·C]** **Matching mode (MVP): column mode.** The sheet contains filenames in a column such as `attachment`. Several per row are supported using a `;` separator or numbered columns (`attachment1`, `attachment2`). Pattern-based matching (filename built from variables) is **P**.
- **ATT-03 [M·C]** **Maximum 5 attachments per email** (shared attachments count toward the 5).
- **ATT-04 [M·C]** Filename matching is **case-insensitive** after trimming whitespace. Matching is by file name within the upload.
- **ATT-05 [M·C]** **Ambiguity:** if an upload contains two files that match the same name (for example in different ZIP folders), the ambiguity is flagged **at upload time** and every recipient referencing that name is **Skipped (AMBIGUOUS_ATTACHMENT)**.
- **ATT-06 [M·C]** **Missing file:** a recipient whose referenced file is not found is **Skipped (MISSING_ATTACHMENT)** and reported. If **any** of a recipient's files is missing or ambiguous, the **whole recipient** is skipped. A partial email is never sent. Never send silently without the attachment.
- **ATT-07 [M·C]** **Unused files** (uploaded but referenced by no row) produce a **warning** only.
- **ATT-08 [M·C]** **Size limits:** 5 MB per attachment, **10 MB total per email** (attachments plus inline images), 500 MB per ZIP upload (configurable). Limits are enforced on the **server**. A recipient exceeding the per-email total is Skipped (ATTACHMENT_TOO_LARGE).
- **ATT-09 [M·C]** **Allow-list of file types** (default: PDF, PNG, JPG/JPEG, GIF, DOCX, XLSX, PPTX, TXT, CSV). Everything else is rejected. The Platform Admin can extend the list. The type is verified from **file content**, not only the extension, and double extensions are rejected. (Exact default list A.)
- **ATT-10 [M·C]** **No malware scanning in the MVP** (P). This is an **accepted risk** (RISK-04). The allow-list is the only technical safeguard.
- **ATT-11 [M·C]** **ZIP safety:** reject password-protected archives, nested archives, path traversal (`../`), absolute paths, excessive file counts, extreme compression ratios ("zip bombs") and zero-byte files. Suggested maximum 10,000 files per ZIP. (Numbers A.)
- **ATT-12 [M·C]** **Storage:** files are **copied into platform storage at upload** so that scheduled and retried sends do not depend on external availability. Storage is encrypted at rest and **separated per organization**.
- **ATT-13 [M·C]** **Access control:** attachments are visible only to users of the owning organization (and to the operator only under a support grant). There is **no public URL**. Every access is logged.
- **ATT-14 [M·C]** **Lifecycle:** attachments are tied to **one campaign** (no reuse across campaigns) and deleted a fixed period after campaign completion (RET-01).
- **ATT-15 [M·A]** Attachment file names shown in the email are sanitized (control characters and path separators removed) but otherwise preserved.
- **ATT-16 [M·C]** If an attachment becomes unavailable between upload and send, the recipient is marked **Failed (ATTACHMENT_UNAVAILABLE)** and the campaign continues. The email is never sent without it.
- **ATT-17 [P·C]** The **presigned-link-in-body** pattern (uploading large images to object storage and linking them) is not part of the MVP. Tenants may paste their own URLs into the template as ordinary links.

## 16. Campaign requirements

- **CMP-01 [M·C]** A **campaign** is one send job: a template version, a recipient dataset, optional attachments, a provider, schedule, and settings.
- **CMP-02 [M·C]** Each campaign has a **type**: `transactional`, `announcement` or `marketing`. The type drives validation (TPL-11) and unsubscribe behavior (UNS-02).
- **CMP-03 [M·C]** Campaign states follow section 11.4.
- **CMP-04 [M·C]** **Validation** runs when the user moves the campaign from Draft toward Ready, and again at launch and at start of sending. A campaign is **Ready** only with no blocking errors.
- **CMP-05 [M·C]** **Freeze at launch:** template version, recipient data, attachments and provider are snapshotted (BR-08).
- **CMP-06 [M·C]** **Duplicate:** a finished or draft campaign can be duplicated (settings and template reference, with new data and attachments required). A **failed-only resend** is available (BR-02 applies). (What duplicate copies A.)
- **CMP-07 [M·C]** A campaign can be **paused, resumed and cancelled** by authorized users (permission matrix). Pause survives restarts.
- **CMP-08 [M·C]** **Approval (optional per organization):** launching moves the campaign to Pending Approval. Approvers are **Reviewer, Admin and Owner (A)**, never the creator (C).
- **CMP-09 [M·C]** An approver may **approve** or **reject with a comment**. Rejection returns the campaign to Draft.
- **CMP-10 [M·C]** **Any edit** to template, data, attachments, provider or schedule after approval invalidates approval (BR-07). Approvals **do not expire** otherwise.
- **CMP-11 [M·C]** A scheduled campaign still unapproved at its start time does **not send**. It moves to Needs Attention and approvers are notified.
- **CMP-12 [M·C]** **Launch confirmation** screen shows: provider (default confirmed), recipient count, skipped count, estimated duration including multi-day spread, schedule and time zone, and the **sender declaration** (ADM-04).
- **CMP-13 [M·C]** A **Suspended** organization or a campaign paused by the Platform Admin cannot send until reinstated.
- **CMP-14 [M·C]** Suppression and caps are re-checked at launch.
- **CMP-15 [M·A]** A campaign has a name, optional labels (ORG-08), a creator, and created and modified timestamps.

## 17. Scheduling requirements

- **SCH-01 [M·C]** Campaigns can be sent **immediately** or **scheduled** for a future time.
- **SCH-02 [M·C]** The **time zone is chosen explicitly** on the schedule screen, defaulting to the organization's. Times are **stored in UTC**. Daylight-saving transitions are handled by storing the intended instant, not a wall-clock offset.
- **SCH-03 [M·C]** A scheduled campaign can be **rescheduled or cancelled until sending starts**.
- **SCH-04 [S·C]** **Sending windows** (for example 09:00 to 17:00 on weekdays in a chosen time zone). Outside a window sending pauses automatically and resumes in the next window.
- **SCH-05 [P·C]** **Blackout periods** and **recurring campaigns** are post-MVP (recurring conflicts with the no-automation non-goal).
- **SCH-06 [M·C]** **Provider daily limits** may spread a campaign over several days (SND-03).
- **SCH-07 [M·C]** **Missed start (downtime):** if the platform was unavailable at the scheduled time, it sends late if within a **grace period of 1 hour (configurable, A)**. Otherwise the campaign moves to **Needs Attention** and the creator and Owner are notified, and it does NOT send automatically.
- **SCH-08 [M·C]** If provider credentials are **changed** or the provider is **deactivated** while a campaign is scheduled, the campaign moves to **Needs Attention** and the Owner is notified. It never sends with credentials different from those it was approved with.
- **SCH-09 [M·C]** Needs Attention is resolved by an authorized user (Owner, Admin, or the creator if permitted) by re-confirming provider and schedule, which re-triggers validation and, if required, approval.
- **SCH-10 [M·C]** A scheduled campaign is never silently dropped. Every transition is audited and notified.

## 18. Sending requirements

- **SND-01 [M·C]** The platform generates each message, connects to the organization's provider, and hands off one email per recipient. Delivery itself is the provider's responsibility.
- **SND-02 [M·C]** **Rate limiting:** messages per minute and per hour per provider are configurable with safe defaults (PRV-09). The platform slows down automatically on provider "slow down" or throttle replies and continues. This is never treated as a permanent failure.
- **SND-03 [M·C]** **Daily limit:** the tenant enters a daily limit per provider. A campaign that exceeds the remaining allowance **continues automatically on following days**, within sending windows if configured. Before launch the user sees a warning such as "this campaign will take 3 days at your limit". Retries and test sends consume the same allowance. The day boundary uses the organization's time zone. (Day boundary A.)
- **SND-04 [M·C]** **Concurrency:** campaigns on the same provider are **queued** in launch order. Different providers run in parallel. Within a campaign the number of simultaneous connections is bounded to respect provider limits. (Order A.)
- **SND-05 [M·C]** Large campaigns taking **hours** are acceptable.
- **SND-06 [M·C]** **Never send twice (idempotency):** each recipient in a campaign has exactly one unique send record. Before any provider handoff the platform durably records an attempt in state Sending. A second concurrent or repeated attempt for the same recipient MUST be refused. On recovery after a crash, any recipient left in Sending without a recorded outcome becomes **Uncertain** (BR-03).
- **SND-07 [M·C]** **Restart recovery:** after an application restart, running campaigns **resume automatically**, except for Uncertain recipients. Paused campaigns stay paused.
- **SND-08 [M·C]** Each email gets a **unique Message-ID** generated by the platform and stored, so bounce matching can be added later without rework.
- **SND-09 [M·C]** **Cancellation while sending** stops further sends. Emails already accepted remain sent. Remaining recipients become Cancelled.
- **SND-10 [M·C]** The suppression list and organization suspension are checked **immediately before each send**.
- **SND-11 [M·C]** **Header and body safety:** the platform must make header injection impossible (PER-04) and must use the confirmed fixed From address.
- **SND-12 [M·C]** **Accepted** means the provider returned a success reply after receiving the full message. Nothing more is claimed.
- **SND-13 [M·A]** Automatic **failover to another provider** is not done in the MVP, because the From identity is fixed per provider.

## 19. Retry and failure requirements

### 19.1 Classification

- **RTY-01 [M·C]** **Temporary failures** (timeouts, connection errors, provider "try again later" replies, rate-limit replies) are **retried automatically**.
- **RTY-02 [M·C]** **Permanent recipient failures** (mailbox does not exist, recipient rejected, content blocked for that recipient) are **not retried**. The recipient is Failed with the provider's reason.
- **RTY-03 [M·C]** **Provider-wide failures** (authentication failure, sender rejected, account suspended) are not charged to the recipient. They **pause the campaign** (RTY-08).

### 19.2 Retry policy

- **RTY-04 [M·C]** Up to **5 attempts** per recipient over roughly **24 hours** with increasing delays (suggested delays after the first attempt: 5 minutes, 30 minutes, 3 hours, 12 hours; configurable by the operator). After exhaustion the recipient is **Failed (RETRIES_EXHAUSTED)**. The user can retry manually. (Numbers A.)
- **RTY-05 [M·C]** Retries never violate BR-02, rate limits or the daily limit, and respect sending windows.

### 19.3 Automatic pause triggers

- **RTY-06 [M·C]** **Authentication failure:** pause immediately.
- **RTY-07 [M·C]** **Repeated connection failures:** pause after **5 consecutive** connection failures. (Number A.)
- **RTY-08 [M·C]** **High failure rate:** pause when more than **20%** of the **last 50** attempts failed, evaluated once at least 20 attempts exist. (Thresholds A.)
- **RTY-09 [M·C]** On any automatic pause the **creator and Owner are notified by email** and in the app. Resuming requires a human action after fixing the cause.

### 19.4 Failure scenarios (all confirmed)

| # | Scenario | Required behavior |
|---|---|---|
| 1 | Provider credentials invalid | Pause campaign and notify |
| 2 | Provider unavailable or network failure | Retry with backoff. If failures continue, pause and notify |
| 3 | Provider rate limit | Slow down and continue. Not a permanent failure |
| 4 | Application crashes or restarts during sending | Running campaigns resume automatically. Recipients in doubt become Uncertain |
| 5 | Worker crashes after provider accepted, before recording | Recipient becomes Uncertain, never auto-resent |
| 6 | Same recipient processed twice by mistake | Refused by the unique send record |
| 7 | Attachment missing at send time | Recipient Failed (ATTACHMENT_UNAVAILABLE). Campaign continues. Never sent without attachment |
| 8 | Corrupted or unreadable Excel or CSV | Upload rejected with clear message. No campaign data created from it |
| 9 | Invalid rows | Error report. Proceed with valid rows if user chooses |
| 10 | Partial failure | Completed with Errors. Failed-only resend available |
| 11 | Cancel during sending | Stop further sends. Accepted emails stay sent |
| 12 | Scheduled time passes during downtime | Send late within 1 hour, otherwise Needs Attention |
| 13 | Scheduled campaign still awaiting approval at start time | Do not send. Needs Attention. Notify approvers |
| 14 | User, organization or campaign suspended while running | Sending pauses immediately. In-flight messages finish |
| 15 | Template archived or edited after launch | No effect. Snapshot is frozen |
| 16 | User edits a recipient row after launch | Not allowed for already-sent rows. Allowed for Failed or Pending rows with audit entry |
| 17 | Storage full | New uploads refused with clear message. Running campaigns continue |
| 18 | Provider daily limit reached mid-campaign | Continue the next day within windows |

---

## 20. Tracking requirements

- **TRK-01 [M·C]** MVP tracking means **what the platform actually knows**: the recipient states in section 11.1. "Accepted" is shown to users as **Sent**, and the UI MUST NEVER say "delivered".
- **TRK-02 [M·C]** **Bounces are not handled in the MVP.** Later bounce notices arrive in the From mailbox and the user checks that mailbox manually. The product SHOULD tell users this plainly (for example next to campaign results).
- **TRK-03 [P·C]** **Opens and clicks** (tracking pixels and link rewriting) are post-MVP. Reasons: unreliable open counts, privacy implications, and the need for a public tracking domain.
- **TRK-04 [P·C]** Future bounce handling options (reading a bounce mailbox, provider webhooks) must remain possible. The stored Message-ID (SND-08) and provider response text support this.
- **TRK-05 [M·C]** **Data stored per email:** recipient address, rendered subject, the data values used, attachment names and sizes, state history with UTC timestamps, provider response or error text, attempt count, Message-ID, and send timestamps. The full rendered body is **not** stored (PER-05).
- **TRK-06 [M·C]** Recipient-level records are removed at purge (RET-01). Aggregate totals remain.

## 21. Reporting requirements

- **RPT-01 [M·C]** **Campaign report:** total recipients, Sent (Accepted), Failed, Skipped (by reason), Uncertain, Cancelled, Pending, start and end times, duration, provider used, approver, creator.
- **RPT-02 [M·C]** A **searchable, filterable per-recipient table** with status, reason and attempt history.
- **RPT-03 [M·C]** **Downloadable CSV** of results and a downloadable **error report** of skipped and failed rows with reasons. Exports MUST neutralize spreadsheet-formula injection (SEC-10).
- **RPT-04 [M·C]** **Organization history:** list of campaigns with filters by status, date, label and provider.
- **RPT-05 [S·A]** **Provider usage view:** emails sent today against the daily limit, per provider.
- **RPT-06 [M·C]** After the retention period only **aggregate totals** remain (RET-01).
- **RPT-07 [M·C]** Open rate, click rate and bounce rate are not reported in the MVP.

## 22. Notification requirements

- **NTF-01 [M·C]** Notifications are sent by **email through the system sender** and also **shown in the app**.
- **NTF-02 [M·C]** Events and recipients (C unless noted):

| Event | Recipients |
|---|---|
| Campaign completed | Creator |
| Campaign completed with errors, or failed | Creator and Owner |
| Campaign paused automatically (authentication failure, connection failures, high failure rate) | Creator and Owner |
| Campaign needs attention (missed schedule, credentials changed, approval not given in time) | Creator and Owner |
| Campaign awaiting approval | Approvers |
| Campaign approved or rejected | Creator |
| Provider daily limit reached, campaign continues tomorrow | Creator |
| Scheduled campaign starting (**S**) | Creator |
| Support access granted, used or expired | Owner |
| Storage quota nearly full, suggested at 80% (A) | Owner and Admin |

- **NTF-03 [P·C]** Per-user notification preferences are post-MVP.
- **NTF-04 [M·C]** Notification emails MUST NOT contain recipient personal data beyond counts, and MUST NOT contain secrets.

## 23. Audit requirements

- **AUD-01 [M·C]** The following MUST be recorded: user sign-in, failed sign-in, sign-out, password and 2FA changes, user creation, removal and role change, provider creation, change, verification, rotation and deactivation (never the secret itself), template creation, edit and archive, recipient data import, attachment upload, campaign creation, edit, submission, approval, rejection, launch, pause, resume, cancel, **recipient data export or download**, data deletion and purge, suppression list changes, **uncertain-recipient resolution**, **support access grants and every access under them**, and **operator actions** (organization creation, suspension, limit changes, campaign pauses).
- **AUD-02 [M·C]** Each entry records: UTC timestamp, organization, actor and role, action, target type and identifier, outcome, source IP and client, and non-sensitive details. Entries MUST NOT contain credentials and SHOULD NOT contain recipient addresses (use counts and identifiers).
- **AUD-03 [M·C]** **Who can view:** Owner and Admin see their organization's log. The Platform Admin sees only operator-level logs and the log of access they themselves made.
- **AUD-04 [M·C]** **Integrity and retention:** append-only, not editable or deletable by tenant users, retained at least **1 year** (configurable).
- **AUD-05 [S·A]** Audit log export as CSV.

## 24. Security requirements

- **SEC-01 [M·C]** **Authentication and authorization** per AUTH-01 to AUTH-05 and the permission matrix, enforced on the server for every operation.
- **SEC-02 [M·C]** Mandatory two-factor authentication for Platform Admin, Owner and Admin (AUTH-02).
- **SEC-03 [M·C]** **Tenant isolation:** every record belongs to an organization. Every query and file access is scoped to the caller's organization, **enforced centrally** and not left to individual endpoints. Files are separated per organization. **Automated tests that attempt cross-tenant access are a required acceptance criterion** (AC-SEC-01).
- **SEC-04 [M·C]** **Credential protection** per PRV-05: encrypted, write-only, key outside the data store, rotatable, never logged.
- **SEC-05 [M·C]** **Transport:** HTTPS for all web traffic, TLS or STARTTLS to providers (PRV-03).
- **SEC-06 [M·C]** **Encryption at rest** for recipient data, attachments, credentials and backups.
- **SEC-07 [M·C]** **SSRF protection** for user-supplied provider addresses (PRV-04).
- **SEC-08 [M·C]** **Header injection** protection (PER-04, SND-11).
- **SEC-09 [M·C]** **Template safety:** templates substitute variables only. They cannot execute code, read files or make network requests. `javascript:` and `data:` links are blocked. HTML is sanitized (TPL-13). Preview is sandboxed.
- **SEC-10 [M·C]** **Formula injection:** cell values beginning with `=`, `+`, `-` or `@` are neutralized in every exported CSV or Excel file.
- **SEC-11 [M·C]** **File and ZIP safety** per ATT-09 to ATT-11, with limits enforced server-side.
- **SEC-12 [M·C]** **Rate limiting** on sign-in, uploads, test sends, password reset and the public unsubscribe endpoint.
- **SEC-13 [M·C]** **Abuse prevention:** organization and campaign suspension, sending caps (switchable), sender declaration, aggregate abuse signals, no first-campaign review in the MVP (**P**).
- **SEC-14 [M·C]** **Sender identity:** From address belongs to a verified provider and cannot be freely typed (PRV-06, PRV-07). SPF, DKIM, DMARC warnings (PRV-08, **S**).
- **SEC-15 [M·C]** **Support access** controls (ADM-03). The Platform Admin MUST NOT freely browse tenant data.
- **SEC-16 [M·C]** **Logging hygiene:** logs and error reports MUST NOT contain credentials or attachment contents, and SHOULD minimize personal data.
- **SEC-17 [M·C]** **Public unsubscribe endpoint** controls (UNS-04) because it is the only public surface.
- **SEC-18 [M·A]** Dependency and secret management follow current best practice. Backups are encrypted and access is restricted.
- **SEC-19 [M·A]** Session tokens are invalidated on password change, role change and deactivation.

## 25. Compliance considerations

The product is **general-purpose** and does not restrict content by recipient relationship (existing contacts, cold or purchased lists). **The tenant is responsible** for the legality of what it sends and for its provider's policies. The platform provides tools, not legal assurance. This section is design guidance, not legal advice. (C)

- **COM-01 [M·C]** Design to the **strictest common baseline** of privacy and email-marketing rules. Current tenants are in Nepal. Applicable local law and any foreign law relevant to recipients must be reviewed by counsel before commercialization.
- **COM-02 [M·C]** **Unsubscribe and consent tools:** marketing campaigns require an unsubscribe link (hard block) and should include the sender's postal address (warning). Opt-outs take effect immediately. Sender declaration at launch.
- **COM-03 [M·C]** **Privacy-by-design** principles are required: data minimization (no permanent contact database), purpose limitation (data tied to a campaign), storage limitation (retention purge), deletion on request, security, and access logging.
- **COM-04 [M·C]** **Roles under privacy law:** tenant is the data controller, platform is the processor. A **Terms of Service** (acceptable use), **privacy notice** and **data processing statement** are required before commercialization and are **post-MVP** (OD-05).
- **COM-05 [M·C]** **Provider and mailbox-provider policies:** tenants must respect their provider's terms. Large senders are increasingly required by major mailbox providers to authenticate with SPF, DKIM and DMARC and to offer one-click unsubscribe. The platform assists through PRV-08 and UNS-01 but does not guarantee deliverability.
- **COM-06 [M·C]** **Tenant-supplied content** and data are the tenant's responsibility. The platform provides abuse reporting, suspension and investigation tools.
- **COM-07 [M·C]** **Backups:** deleted data may persist in encrypted backups for up to about **30 days**. After any restore from backup, previous deletions and purges MUST be re-applied before the system returns to service. (Re-apply rule A.)
- **COM-08 [M·C]** **Transactional email** is not required to carry an unsubscribe link, but organizations should not use the transactional type to avoid marketing rules. Type selection is the sender's declaration.

## 26. API requirements

- **API-01 [M·C]** **No public API in the MVP.** All business operations (create campaign, import data, validate, launch, pause, cancel, read status and results, manage providers and templates) MUST be implemented as a **reusable service layer** so that the web interface and a future API call **the same logic and obey the same rules** (approvals, quotas, caps, suppression, tenant scoping, audit).
- **API-02 [P·C]** Future public API: **organization-scoped API keys with limited permissions** (for example create and launch campaigns but not manage users), **idempotency keys** on launch, **status and result retrieval**, **webhooks** for completion events, and rate limits.
- **API-03 [P·C]** Likely first consumers are the owner's other applications that will trigger campaigns, check status and retrieve results. No specific integration is required now.

## 27. Scalability requirements

### 27.1 Design targets (C)

| Dimension | Target |
|---|---|
| Organizations | ~10 initially (headroom for 100) |
| Users per organization | 3 to 5 |
| Campaigns per organization per month | ~1 (peaks possible) |
| Recipients per campaign | typical 100 to 1,000, maximum 5,000 committed, 10,000 headroom, 100,000 must not be architecturally blocked |
| Emails per day, platform-wide | up to ~3,000 |
| Attachment size | at most 5 MB each, 10 MB per email |
| Campaign duration | hours acceptable |
| Commercial growth | further in the future |

### 27.2 Requirements

- **SCL-01 [M·C]** Large imports, validation, attachment matching and sending run **asynchronously** with visible progress, and do not block the interface.
- **SCL-02 [M·C]** Design MUST NOT assume a whole dataset fits in memory or on a single page, so volumes up to 100,000 recipients are not architecturally blocked.
- **SCL-03 [M·C]** Different organizations' campaigns run **in parallel**. Same-provider campaigns queue (BR-09).
- **SCL-04 [M·C]** **Per-organization storage quota** (ORG-09). Worst case reminder: 5,000 recipients at 5 MB each is 25 GB for one campaign, so typical attachment size is needed (OD-01).
- **SCL-05 [M·C]** **Availability:** about 99% for the MVP. Planned maintenance with prior notice is acceptable. Occasional unplanned downtime is acceptable.
- **SCL-06 [M·C]** **Recoverability:** minimal data loss. Campaign, recipient and sending-status data MUST be recoverable from backups. Exact RPO and RTO are an open decision (OD-02).
- **SCL-07 [M·A]** The system SHOULD expose operational metrics (queue depth, send rate, failures per provider) to the operator.

## 28. Data-retention requirements

- **RET-01 [M·C]** **Recipient data and attachments** are automatically purged **90 days after campaign completion** by default. After purge, only **aggregate campaign totals** and the audit trail remain.
- **RET-02 [M·C]** The retention period is **configurable per organization** (Owner). Suggested allowed range: 7 to 365 days. (Range A.)
- **RET-03 [M·C]** **Deletion on request:** Owner or Admin can delete a campaign's recipient data and attachments at any time. This is audited without recording the deleted personal data.
- **RET-04 [M·C]** **Audit logs:** at least 1 year, configurable (AUD-04).
- **RET-05 [M·A]** **Suppression list** entries are kept until removed by Owner or Admin, because they are needed to honor opt-outs. Whether to store them hashed is OD-08.
- **RET-06 [M·C]** **Backups** per COM-07.
- **RET-07 [M·A]** A campaign still sending or scheduled is never purged. The retention clock starts at completion, cancellation or failure.
- **RET-08 [M·A]** Retention purge failures are alerted to the operator.

## 29. Edge cases

Each edge case has an expected behavior (C if derived from confirmed rules, otherwise A).

| ID | Situation | Expected behavior |
|---|---|---|
| EDGE-01 | Same email appears in several rows | Default key is email, first kept, others Skipped (DUPLICATE). User may select a different key, such as email plus ticket number, or disable dedupe. Warning lists dropped rows |
| EDGE-02 | Email has surrounding spaces or mixed case | Trimmed. Domain compared case-insensitively. Original value preserved in the record |
| EDGE-03 | Address with display name (`John <j@x.com>`) or multiple addresses in one cell | INVALID_EMAIL in the MVP (one plain address per row) |
| EDGE-04 | Non-ASCII (internationalized) email addresses | Rejected as INVALID_EMAIL in the MVP unless the provider path is proven to support it (A, see OD-11) |
| EDGE-05 | Excel number cells such as ticket `00123` lose leading zeros | Show a warning when a mapped column mixes text and numeric cells. Recommend text format |
| EDGE-06 | CSV with different delimiter, encoding or byte-order mark | Auto-detect delimiter, accept UTF-8 with or without BOM, reject other encodings with a clear message |
| EDGE-07 | Excel with several sheets, hidden sheets or macros | User picks the sheet. Hidden sheets are ignored unless picked. Macro-enabled files are refused |
| EDGE-08 | Duplicate or blank column headers | Block mapping and ask user to fix. Blank headers are ignored |
| EDGE-09 | A recipient unsubscribes (or is added to suppression) while the campaign is running | Suppression is re-checked right before each send. The recipient is Skipped (SUPPRESSED) |
| EDGE-10 | Attachment filename has different case, or extra spaces | Matches case-insensitively after trimming |
| EDGE-11 | Two files with the same name in different ZIP folders | Flagged at upload. Referencing recipients Skipped (AMBIGUOUS_ATTACHMENT) |
| EDGE-12 | ZIP is password protected, nested, corrupt, or contains path traversal | Upload rejected with clear reason |
| EDGE-13 | Zero-byte, oversized or disallowed attachment | Rejected at upload. Referencing recipients Skipped with a reason |
| EDGE-14 | Sum of attachments plus inline images exceeds 10 MB | Recipient Skipped (ATTACHMENT_TOO_LARGE) |
| EDGE-15 | Template uses a variable that exists in no column and has no default | Validation error, campaign cannot become Ready |
| EDGE-16 | Value contains HTML or script text | Inserted as escaped text. Never interpreted |
| EDGE-17 | Value contains a line break in a subject or name | Line breaks removed in subject and headers |
| EDGE-18 | Subject becomes very long or empty after substitution | Very long: warning. Empty: error |
| EDGE-19 | Row edited after launch | Only Failed or Pending rows. Audited |
| EDGE-20 | Creator is deactivated mid-campaign | Campaign continues. Creator reference preserved. Notifications go to the Owner as well |
| EDGE-21 | Approver's role changes after approving | Approval stands. It is invalidated only by campaign edits |
| EDGE-22 | Provider credentials rotated while a campaign is running | Running campaign finishes with the new credentials if authentication succeeds, and is paused if it fails. Rotation while **scheduled** moves the campaign to Needs Attention (SCH-08) |
| EDGE-23 | Provider deactivated while campaign running | Campaign pauses immediately |
| EDGE-24 | Daylight-saving change between scheduling and start time | The stored instant is honored. The schedule screen shows the resulting local time |
| EDGE-25 | Two campaigns launched on the same provider at once | Second is queued behind the first |
| EDGE-26 | Duplicate campaign created from an old one | New draft. New data and attachments required. Template reference is to the latest version unless the user picks one |
| EDGE-27 | Storage quota exceeded during upload | Upload refused with clear message. No partial attachment set remains for matching |
| EDGE-28 | Test email address differs from campaign data | Allowed. Counts against provider limits. Subject prefixed `[TEST]` |
| EDGE-29 | Marketing campaign without unsubscribe link | Blocked at validation |
| EDGE-30 | Transactional email to an address that unsubscribed through an unsubscribe link | Sent normally (the entry is `non_transactional`). Blocked only if an `all` entry exists for the address |
| EDGE-31 | Restore from backup after purge or deletion | Deletions and purges re-applied before service resumes (COM-07) |
| EDGE-32 | Recipient data purged while someone is viewing the campaign report | Report shows aggregate totals and a "data purged" notice |
| EDGE-33 | Organization suspended during scheduled period | Scheduled campaigns do not start. Reinstatement requires user re-confirmation (Needs Attention) |
| EDGE-34 | Daily limit smaller than campaign size | Spreads over days. Launch screen shows the estimate |
| EDGE-35 | User opens the unsubscribe link twice or after being removed from suppression | Idempotent. A repeat adds the address again only if clicked again after removal, and shows a confirmation either way |
| EDGE-36 | Spreadsheet includes more than 10,000 rows or exceeds file size | Rejected with clear limit message |
| EDGE-37 | Attachment column empty for some rows | Those recipients receive no attachment only if the campaign is configured as "attachment optional". Otherwise Skipped (MISSING_ATTACHMENT). Default: attachment required when the campaign uses column mode (A) |

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

## 31. MVP scope

Legend: **M** must, **S** should, **P** post-MVP. This is the client-approved scope table (C).

### Organizations, users, access
| Feature | Priority | IDs |
|---|---|---|
| Operator creates and suspends organizations, org settings | M | ORG-01 to ORG-07 |
| Invitation-only users, one organization per account, roles | M | AUTH-03, ORG-05, section 6 |
| Email and password login with mandatory 2FA for privileged roles | M | AUTH-01, AUTH-02 |
| Departments, teams or projects as labels | S | ORG-08 |
| Single sign-on | P | AUTH-06 |

### Providers
| Feature | Priority | IDs |
|---|---|---|
| Generic SMTP (TLS required), Gmail and Microsoft 365 via app passwords | M | PRV-02, PRV-03 |
| Several providers per org, default confirmed at send | M | PRV-01 |
| Verification by connection and test email, fixed From, Reply-To | M | PRV-06, PRV-07 |
| Encrypted write-only credentials, rotation, SSRF protection | M | PRV-04, PRV-05 |
| Daily and rate limits per provider | M | PRV-09, SND-02, SND-03 |
| SPF, DKIM, DMARC warnings | S | PRV-08 |
| OAuth, provider APIs, automatic failover | P | PRV-02, SND-13 |

### Templates and data
| Feature | Priority | IDs |
|---|---|---|
| HTML and plain text, rich editor plus raw HTML, variables, subject variables | M | TPL-01 to TPL-04 |
| Defaults and required flags, escaped values | M | TPL-05, PER-03 |
| Versioning, snapshot, duplicate, archive | M | TPL-06, TPL-07 |
| Preview with real rows, test send | M | TPL-08, TPL-09 |
| Template validation (advisory spam check is S) | M / S | TPL-11 |
| Excel and CSV import, column mapping, validation, error report, single-row edit | M | DAT-01 to DAT-08 |
| Conditionals, loops, other data sources | P | TPL-03, DAT-01 |

### Attachments
| Feature | Priority | IDs |
|---|---|---|
| ZIP or multi-file upload, column-mode matching, up to 5 per email | M | ATT-01 to ATT-08 |
| Shared attachment, inline images | M | ATT-01, TPL-10 |
| Allow-list, limits, copy on upload, per-org quota | M | ATT-09, ATT-12, ORG-09 |
| URL per row | S | ATT-01 |
| Malware scanning, presigned links, generated files, pattern matching | P | ATT-02, ATT-10, ATT-17 |

### Campaigns and sending
| Feature | Priority | IDs |
|---|---|---|
| Campaign types, full state model, freeze at launch | M | CMP-01 to CMP-05 |
| Optional approval workflow | M | CMP-08 to CMP-11 |
| Immediate and scheduled sending, explicit time zone | M | SCH-01 to SCH-03 |
| Pause, resume, cancel, duplicate, resend failed | M | CMP-06, CMP-07 |
| Retries, Uncertain handling, never-send-twice | M | SND-06, RTY-01 to RTY-05 |
| Auto-pause, per-provider queue, auto-resume after restart | M | RTY-06 to RTY-09, SND-04, SND-07 |
| Sending windows | S | SCH-04 |
| Recurring campaigns, blackout periods | P | SCH-05 |

### Tracking, reporting, notifications, audit
| Feature | Priority | IDs |
|---|---|---|
| Status through Accepted, failure reasons | M | TRK-01, TRK-05 |
| Campaign report, per-recipient table, CSV and error exports | M | RPT-01 to RPT-04 |
| Organization suppression list, unsubscribe page and one-click header | M | SUP-01 to SUP-05, UNS-01 to UNS-05 |
| Email and in-app notifications | M | NTF-01, NTF-02 |
| Audit log and support-access grants | M | AUD-01 to AUD-04, ADM-03 |
| Provider usage view | S | RPT-05 |
| Bounce handling, opens, clicks, notification preferences | P | TRK-03, TRK-04, NTF-03 |

### Platform and non-functional
| Feature | Priority | IDs |
|---|---|---|
| Tenant isolation with automated cross-tenant tests | M | SEC-03, AC-SEC-01 |
| Retention purge and deletion on request | M | RET-01 to RET-03 |
| Abuse controls (suspension, caps switchable, declaration, rate limits) | M | ADM-01 to ADM-04, ORG-10 |
| System sender for notifications | M | PRV-12 |
| Reusable service layer | M | API-01 |
| Public API, keys, webhooks, billing, plans, legal documents | P | API-02, COM-04 |

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

## 33. Future considerations

Features that should not be built now but may influence design:

- **Provider abstraction** must allow API-based providers, OAuth-based providers and webhook-based events.
- **Per-email Message-ID and provider response** are stored so bounces and events can be matched later.
- **Recipient states** must be extensible with Delivered, Bounced, Opened and Clicked without breaking reports.
- **Suppression** may later gain a platform layer and automatic population from bounces and complaints (scope by campaign type is already in the MVP, SUP-05).
- **Commercialization:** plans, quotas, metering, tenant self-signup, providing a managed sender (for example SES) to tenants who have none. Providing a managed sender makes the operator the account owner with all the reputation and abuse consequences, so it must be a conscious later decision.
- **Contact lists** are a possible future feature but conflict with the current non-goal.
- **Multi-organization user membership** was declined for now. A later change affects the identity model.
- **Localization** of the interface and templates, multi-language recipients.
- **100,000-recipient campaigns** and multi-day sending.
- **Data residency** options per organization.

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

## 36. Glossary

| Term | Meaning |
|---|---|
| **Accepted** | The provider returned a success reply after receiving the complete message. Shown to users as "Sent". Not the same as delivered |
| **Approval** | Optional second-person review of a campaign before launch |
| **Audit log** | Append-only record of significant actions |
| **Campaign** | One send job: template version, recipient data, attachments, provider, schedule and settings |
| **Campaign type** | Transactional, announcement or marketing. Drives validations and unsubscribe behavior |
| **Column mode** | Attachment matching where a sheet column holds the file names |
| **Default (variable)** | A fallback value defined in a template's variable settings |
| **Freeze / snapshot** | Copy of template version, data, attachments and provider taken at launch |
| **Needs Attention** | Campaign state requiring a human decision before it can proceed |
| **One-to-one** | One email per row, addressed to one recipient |
| **Organization / tenant** | A customer unit with isolated data |
| **Platform Admin / operator** | The product owner's administrative account type outside any tenant |
| **Provider** | The external email service or SMTP server an organization connects |
| **Rate limit** | Maximum messages per minute or hour to a provider |
| **Recipient row** | One imported row, representing one email to be sent |
| **Required variable** | A template variable that must have a value or default |
| **Skipped** | A recipient removed before sending for a specific reason, such as duplicate or missing attachment |
| **SMTP** | The standard protocol for handing email to a mail server |
| **SPF / DKIM / DMARC** | DNS-based email authentication standards that help mailbox providers trust a sender's domain |
| **SSRF** | Server-side request forgery: tricking the server into connecting to an unintended internal address |
| **Suppression list** | Per-organization list of addresses that must not be emailed. Each entry has a scope: `non_transactional` (marketing and announcement) or `all` |
| **System sender** | Operator-owned email account used only for platform notifications |
| **Support grant** | Time-limited, Owner-issued permission for the operator to view tenant content |
| **Template** | Reusable email subject and body with variables |
| **Uncertain** | Recipient whose send outcome is unknown after a failure and needs a human decision |
| **Unsubscribe (one-click)** | Standard mechanism allowing a recipient to opt out directly from the mail client |
| **Variable** | A `{{placeholder}}` replaced per recipient |

---

*End of document. Draft 0.2. Next steps after client review: resolve open decisions (section 34.1), then proceed to technology stack, architecture, database design, API specification, background-job design, infrastructure and deployment planning.*
