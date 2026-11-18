## 28. Data-retention requirements

- **RET-01 [M·C]** **Recipient data and attachments** are automatically purged **90 days after campaign completion** by default. After purge, only **aggregate campaign totals** and the audit trail remain.
- **RET-02 [M·C]** The retention period is **configurable per organization** (Owner). Suggested allowed range: 7 to 365 days. (Range A.)
- **RET-03 [M·C]** **Deletion on request:** Owner or Admin can delete a campaign's recipient data and attachments at any time. This is audited without recording the deleted personal data.
- **RET-04 [M·C]** **Audit logs:** at least 1 year, configurable (AUD-04).
- **RET-05 [M·A]** **Suppression list** entries are kept until removed by Owner or Admin, because they are needed to honor opt-outs. Whether to store them hashed is OD-08.
- **RET-06 [M·C]** **Backups** per COM-07.
- **RET-07 [M·A]** A campaign still sending or scheduled is never purged. The retention clock starts at completion, cancellation or failure.
- **RET-08 [M·A]** Retention purge failures are alerted to the operator.
