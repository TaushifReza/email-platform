## 13. Personalization requirements

- **PER-01 [M·C]** Any mapped column can be used as a variable in subject and body.
- **PER-02 [M·C]** **Missing value rule:** if a value is empty, use the template default if defined. If there is none and the variable is **required**, the recipient is **Skipped (MISSING_REQUIRED_VALUE)** or the whole campaign is **blocked**, depending on a **per-campaign strictness setting**. An empty **optional** variable with no default is rendered as an empty string. (Strictness default: block. Default A.)
- **PER-03 [M·C]** Values are **HTML-escaped** in HTML bodies. Raw HTML from data is not supported. In the plain-text part values are inserted literally, with control characters stripped.
- **PER-04 [M·C]** Values inserted into the **subject** or any header MUST have line breaks and control characters removed (header-injection protection).
- **PER-05 [M·C]** The rendered result per recipient can be **re-created on demand** from the frozen snapshot. Full rendered bodies are **not stored** per recipient.
- **PER-06 [M·A]** Leading and trailing whitespace in values is trimmed. Numeric-looking cells such as ticket numbers keep their displayed text. A warning is shown if leading zeros appear lost (EDGE-05).
