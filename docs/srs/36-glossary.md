## 36. Glossary

| Term | Meaning |
|---|---|
| **Accepted** | The provider returned a success reply after receiving the complete message. Shown to users as "Sent". Not the same as delivered |
| **Approval** | Optional second-person review of a campaign before launch |
| **Audit log** | Append-only record of significant actions |
| **Campaign** | One send job: template version, recipient data, attachments, provider, schedule and settings |
| **Campaign type** | Transactional, announcement or marketing. Drives validations and unsubscribe behavior |
| **Column mode** | Attachment matching where a sheet column holds the file names |
| **Default (variable)** | A fallback value defined in a template's variable settings |
| **Freeze / snapshot** | Copy of template version, data, attachments and provider taken at launch |
| **Needs Attention** | Campaign state requiring a human decision before it can proceed |
| **One-to-one** | One email per row, addressed to one recipient |
| **Organization / tenant** | A customer unit with isolated data |
| **Platform Admin / operator** | The product owner's administrative account type outside any tenant |
| **Provider** | The external email service or SMTP server an organization connects |
| **Rate limit** | Maximum messages per minute or hour to a provider |
| **Recipient row** | One imported row, representing one email to be sent |
| **Required variable** | A template variable that must have a value or default |
| **Skipped** | A recipient removed before sending for a specific reason, such as duplicate or missing attachment |
| **SMTP** | The standard protocol for handing email to a mail server |
| **SPF / DKIM / DMARC** | DNS-based email authentication standards that help mailbox providers trust a sender's domain |
| **SSRF** | Server-side request forgery: tricking the server into connecting to an unintended internal address |
| **Suppression list** | Per-organization list of addresses that must not be emailed. Each entry has a scope: `non_transactional` (marketing and announcement) or `all` |
| **System sender** | Operator-owned email account used only for platform notifications |
| **Support grant** | Time-limited, Owner-issued permission for the operator to view tenant content |
| **Template** | Reusable email subject and body with variables |
| **Uncertain** | Recipient whose send outcome is unknown after a failure and needs a human decision |
| **Unsubscribe (one-click)** | Standard mechanism allowing a recipient to opt out directly from the mail client |
| **Variable** | A `{{placeholder}}` replaced per recipient |

---

*End of document. Draft 0.2. Next steps after client review: resolve open decisions (section 34.1), then proceed to technology stack, architecture, database design, API specification, background-job design, infrastructure and deployment planning.*
