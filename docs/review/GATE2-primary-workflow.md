# Gate 2: primary workflow on staging (end of week 3)

## Click-through (staging, Mailpit)
1. Create a template with `{{name}}` and `{{ticket_number}}`, preview, send a test.
2. New campaign, upload a 1,000-row Excel file with 20 deliberately bad rows. Check the error report and proceed with valid rows.
3. Upload a ZIP of per-recipient PDFs larger than 100 MB (chunked upload). Remove one file and add a duplicate name: those recipients are skipped with the right reasons.
4. Launch. Watch progress. Pause, resume, cancel a second run.
5. **Kill the worker during a run, restart it.** No recipient appears twice in Mailpit. Any Uncertain recipient is explained.
6. Use the fake provider that returns 4xx and 5xx replies: retries, Failed reasons, auto-pause.
7. Download the results CSV: no cell can execute as a formula.

## Read personally
- [ ] `internal/sending` guarded transitions and the claim, handoff, record sequence match `03-sending-protocol.md`
- [ ] Crash suite output: all invariants hold for the 5,000-recipient run
