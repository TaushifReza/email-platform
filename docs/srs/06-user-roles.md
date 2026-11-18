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
