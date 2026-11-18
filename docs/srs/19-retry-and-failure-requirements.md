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
