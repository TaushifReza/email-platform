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
