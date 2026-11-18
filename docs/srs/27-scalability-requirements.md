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
