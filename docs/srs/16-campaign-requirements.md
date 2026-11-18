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
