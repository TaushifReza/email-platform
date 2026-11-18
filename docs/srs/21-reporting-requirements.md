## 21. Reporting requirements

- **RPT-01 [M·C]** **Campaign report:** total recipients, Sent (Accepted), Failed, Skipped (by reason), Uncertain, Cancelled, Pending, start and end times, duration, provider used, approver, creator.
- **RPT-02 [M·C]** A **searchable, filterable per-recipient table** with status, reason and attempt history.
- **RPT-03 [M·C]** **Downloadable CSV** of results and a downloadable **error report** of skipped and failed rows with reasons. Exports MUST neutralize spreadsheet-formula injection (SEC-10).
- **RPT-04 [M·C]** **Organization history:** list of campaigns with filters by status, date, label and provider.
- **RPT-05 [S·A]** **Provider usage view:** emails sent today against the daily limit, per provider.
- **RPT-06 [M·C]** After the retention period only **aggregate totals** remain (RET-01).
- **RPT-07 [M·C]** Open rate, click rate and bounce rate are not reported in the MVP.
