## 20. Tracking requirements

- **TRK-01 [M·C]** MVP tracking means **what the platform actually knows**: the recipient states in section 11.1. "Accepted" is shown to users as **Sent**, and the UI MUST NEVER say "delivered".
- **TRK-02 [M·C]** **Bounces are not handled in the MVP.** Later bounce notices arrive in the From mailbox and the user checks that mailbox manually. The product SHOULD tell users this plainly (for example next to campaign results).
- **TRK-03 [P·C]** **Opens and clicks** (tracking pixels and link rewriting) are post-MVP. Reasons: unreliable open counts, privacy implications, and the need for a public tracking domain.
- **TRK-04 [P·C]** Future bounce handling options (reading a bounce mailbox, provider webhooks) must remain possible. The stored Message-ID (SND-08) and provider response text support this.
- **TRK-05 [M·C]** **Data stored per email:** recipient address, rendered subject, the data values used, attachment names and sizes, state history with UTC timestamps, provider response or error text, attempt count, Message-ID, and send timestamps. The full rendered body is **not** stored (PER-05).
- **TRK-06 [M·C]** Recipient-level records are removed at purge (RET-01). Aggregate totals remain.
