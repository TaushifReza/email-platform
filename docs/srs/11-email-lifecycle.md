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
