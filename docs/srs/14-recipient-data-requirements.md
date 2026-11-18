## 14. Recipient-data requirements

- **DAT-01 [M·C]** Accepted formats: **Excel (.xlsx) and CSV**. Not in MVP: .xls, TSV, Google Sheets links, API upload, database connections, webhook lists (**P**).
- **DAT-02 [M·C]** Limits: **10,000 rows and about 10 MB per file**. Design target for committed use is 5,000 recipients per campaign. Larger files are refused with a clear message.
- **DAT-03 [M·C]** The **only mandatory column is the recipient email address**. Any other column can be mapped to any variable. There is no fixed schema.
- **DAT-04 [M·C]** **Column mapping:** columns whose headers equal a variable name are **auto-matched**. The user can adjust. Mappings may be saved and reused for later campaigns with the same layout. (Saved mappings A.)
- **DAT-05 [M·C]** **Validation at import** (all required):
  1. email syntax check,
  2. duplicates within the file (see DAT-06),
  3. missing required values for the template,
  4. domain check (the address's domain must exist and be able to receive mail, to catch typos like `gmial.com`; a DNS timeout produces a warning, not an error),
  5. suppression list check,
  6. attachment checks (section 15),
  7. cap checks (ORG-10).
- **DAT-06 [M·C]** **Duplicates:** detection is on by default using the **email address** as the key. The user can choose a different **uniqueness key** (for example email plus ticket number) or switch it off for the campaign. By default the **first occurrence is kept** and later ones are **Skipped (DUPLICATE)**. The warning lists exactly which rows were dropped. Email comparison is case-insensitive.
- **DAT-07 [M·C]** **Invalid rows:** a downloadable **per-row error report** is produced. The user chooses **proceed with valid rows** or fix and re-upload.
- **DAT-08 [M·C]** **Editing:** a user can **edit a single row** inside the platform before launch. After launch only Failed or Pending rows can be edited, with an audit entry. **Bulk editing** is not supported.
- **DAT-09 [M·C]** Recipient data **belongs to the campaign**. There is no permanent contact database. A campaign's data can be deleted on request (RET-03).
- **DAT-10 [M·C]** One email per row (**one-to-one**). No CC, BCC, or several To addresses in the MVP.
- **DAT-11 [M·A]** Excel files with several sheets: the user chooses the sheet. Formula cells use their stored value. Files with macros are refused. CSV encoding is UTF-8 with automatic delimiter detection and byte-order-mark tolerance.
- **DAT-12 [M·A]** Import of large files runs **asynchronously** with visible progress, and the UI paginates data of 10,000 rows. Row values are never trusted as markup (see section 24).
