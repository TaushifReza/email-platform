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
