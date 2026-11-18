## 29. Edge cases

Each edge case has an expected behavior (C if derived from confirmed rules, otherwise A).

| ID | Situation | Expected behavior |
|---|---|---|
| EDGE-01 | Same email appears in several rows | Default key is email, first kept, others Skipped (DUPLICATE). User may select a different key, such as email plus ticket number, or disable dedupe. Warning lists dropped rows |
| EDGE-02 | Email has surrounding spaces or mixed case | Trimmed. Domain compared case-insensitively. Original value preserved in the record |
| EDGE-03 | Address with display name (`John <j@x.com>`) or multiple addresses in one cell | INVALID_EMAIL in the MVP (one plain address per row) |
| EDGE-04 | Non-ASCII (internationalized) email addresses | Rejected as INVALID_EMAIL in the MVP unless the provider path is proven to support it (A, see OD-11) |
| EDGE-05 | Excel number cells such as ticket `00123` lose leading zeros | Show a warning when a mapped column mixes text and numeric cells. Recommend text format |
| EDGE-06 | CSV with different delimiter, encoding or byte-order mark | Auto-detect delimiter, accept UTF-8 with or without BOM, reject other encodings with a clear message |
| EDGE-07 | Excel with several sheets, hidden sheets or macros | User picks the sheet. Hidden sheets are ignored unless picked. Macro-enabled files are refused |
| EDGE-08 | Duplicate or blank column headers | Block mapping and ask user to fix. Blank headers are ignored |
| EDGE-09 | A recipient unsubscribes (or is added to suppression) while the campaign is running | Suppression is re-checked right before each send. The recipient is Skipped (SUPPRESSED) |
| EDGE-10 | Attachment filename has different case, or extra spaces | Matches case-insensitively after trimming |
| EDGE-11 | Two files with the same name in different ZIP folders | Flagged at upload. Referencing recipients Skipped (AMBIGUOUS_ATTACHMENT) |
| EDGE-12 | ZIP is password protected, nested, corrupt, or contains path traversal | Upload rejected with clear reason |
| EDGE-13 | Zero-byte, oversized or disallowed attachment | Rejected at upload. Referencing recipients Skipped with a reason |
| EDGE-14 | Sum of attachments plus inline images exceeds 10 MB | Recipient Skipped (ATTACHMENT_TOO_LARGE) |
| EDGE-15 | Template uses a variable that exists in no column and has no default | Validation error, campaign cannot become Ready |
| EDGE-16 | Value contains HTML or script text | Inserted as escaped text. Never interpreted |
| EDGE-17 | Value contains a line break in a subject or name | Line breaks removed in subject and headers |
| EDGE-18 | Subject becomes very long or empty after substitution | Very long: warning. Empty: error |
| EDGE-19 | Row edited after launch | Only Failed or Pending rows. Audited |
| EDGE-20 | Creator is deactivated mid-campaign | Campaign continues. Creator reference preserved. Notifications go to the Owner as well |
| EDGE-21 | Approver's role changes after approving | Approval stands. It is invalidated only by campaign edits |
| EDGE-22 | Provider credentials rotated while a campaign is running | Running campaign finishes with the new credentials if authentication succeeds, and is paused if it fails. Rotation while **scheduled** moves the campaign to Needs Attention (SCH-08) |
| EDGE-23 | Provider deactivated while campaign running | Campaign pauses immediately |
| EDGE-24 | Daylight-saving change between scheduling and start time | The stored instant is honored. The schedule screen shows the resulting local time |
| EDGE-25 | Two campaigns launched on the same provider at once | Second is queued behind the first |
| EDGE-26 | Duplicate campaign created from an old one | New draft. New data and attachments required. Template reference is to the latest version unless the user picks one |
| EDGE-27 | Storage quota exceeded during upload | Upload refused with clear message. No partial attachment set remains for matching |
| EDGE-28 | Test email address differs from campaign data | Allowed. Counts against provider limits. Subject prefixed `[TEST]` |
| EDGE-29 | Marketing campaign without unsubscribe link | Blocked at validation |
| EDGE-30 | Transactional email to an address that unsubscribed through an unsubscribe link | Sent normally (the entry is `non_transactional`). Blocked only if an `all` entry exists for the address |
| EDGE-31 | Restore from backup after purge or deletion | Deletions and purges re-applied before service resumes (COM-07) |
| EDGE-32 | Recipient data purged while someone is viewing the campaign report | Report shows aggregate totals and a "data purged" notice |
| EDGE-33 | Organization suspended during scheduled period | Scheduled campaigns do not start. Reinstatement requires user re-confirmation (Needs Attention) |
| EDGE-34 | Daily limit smaller than campaign size | Spreads over days. Launch screen shows the estimate |
| EDGE-35 | User opens the unsubscribe link twice or after being removed from suppression | Idempotent. A repeat adds the address again only if clicked again after removal, and shows a confirmation either way |
| EDGE-36 | Spreadsheet includes more than 10,000 rows or exceeds file size | Rejected with clear limit message |
| EDGE-37 | Attachment column empty for some rows | Those recipients receive no attachment only if the campaign is configured as "attachment optional". Otherwise Skipped (MISSING_ATTACHMENT). Default: attachment required when the campaign uses column mode (A) |
