## 31. MVP scope

Legend: **M** must, **S** should, **P** post-MVP. This is the client-approved scope table (C).

### Organizations, users, access
| Feature | Priority | IDs |
|---|---|---|
| Operator creates and suspends organizations, org settings | M | ORG-01 to ORG-07 |
| Invitation-only users, one organization per account, roles | M | AUTH-03, ORG-05, section 6 |
| Email and password login with mandatory 2FA for privileged roles | M | AUTH-01, AUTH-02 |
| Departments, teams or projects as labels | S | ORG-08 |
| Single sign-on | P | AUTH-06 |

### Providers
| Feature | Priority | IDs |
|---|---|---|
| Generic SMTP (TLS required), Gmail and Microsoft 365 via app passwords | M | PRV-02, PRV-03 |
| Several providers per org, default confirmed at send | M | PRV-01 |
| Verification by connection and test email, fixed From, Reply-To | M | PRV-06, PRV-07 |
| Encrypted write-only credentials, rotation, SSRF protection | M | PRV-04, PRV-05 |
| Daily and rate limits per provider | M | PRV-09, SND-02, SND-03 |
| SPF, DKIM, DMARC warnings | S | PRV-08 |
| OAuth, provider APIs, automatic failover | P | PRV-02, SND-13 |

### Templates and data
| Feature | Priority | IDs |
|---|---|---|
| HTML and plain text, rich editor plus raw HTML, variables, subject variables | M | TPL-01 to TPL-04 |
| Defaults and required flags, escaped values | M | TPL-05, PER-03 |
| Versioning, snapshot, duplicate, archive | M | TPL-06, TPL-07 |
| Preview with real rows, test send | M | TPL-08, TPL-09 |
| Template validation (advisory spam check is S) | M / S | TPL-11 |
| Excel and CSV import, column mapping, validation, error report, single-row edit | M | DAT-01 to DAT-08 |
| Conditionals, loops, other data sources | P | TPL-03, DAT-01 |

### Attachments
| Feature | Priority | IDs |
|---|---|---|
| ZIP or multi-file upload, column-mode matching, up to 5 per email | M | ATT-01 to ATT-08 |
| Shared attachment, inline images | M | ATT-01, TPL-10 |
| Allow-list, limits, copy on upload, per-org quota | M | ATT-09, ATT-12, ORG-09 |
| URL per row | S | ATT-01 |
| Malware scanning, presigned links, generated files, pattern matching | P | ATT-02, ATT-10, ATT-17 |

### Campaigns and sending
| Feature | Priority | IDs |
|---|---|---|
| Campaign types, full state model, freeze at launch | M | CMP-01 to CMP-05 |
| Optional approval workflow | M | CMP-08 to CMP-11 |
| Immediate and scheduled sending, explicit time zone | M | SCH-01 to SCH-03 |
| Pause, resume, cancel, duplicate, resend failed | M | CMP-06, CMP-07 |
| Retries, Uncertain handling, never-send-twice | M | SND-06, RTY-01 to RTY-05 |
| Auto-pause, per-provider queue, auto-resume after restart | M | RTY-06 to RTY-09, SND-04, SND-07 |
| Sending windows | S | SCH-04 |
| Recurring campaigns, blackout periods | P | SCH-05 |

### Tracking, reporting, notifications, audit
| Feature | Priority | IDs |
|---|---|---|
| Status through Accepted, failure reasons | M | TRK-01, TRK-05 |
| Campaign report, per-recipient table, CSV and error exports | M | RPT-01 to RPT-04 |
| Organization suppression list, unsubscribe page and one-click header | M | SUP-01 to SUP-05, UNS-01 to UNS-05 |
| Email and in-app notifications | M | NTF-01, NTF-02 |
| Audit log and support-access grants | M | AUD-01 to AUD-04, ADM-03 |
| Provider usage view | S | RPT-05 |
| Bounce handling, opens, clicks, notification preferences | P | TRK-03, TRK-04, NTF-03 |

### Platform and non-functional
| Feature | Priority | IDs |
|---|---|---|
| Tenant isolation with automated cross-tenant tests | M | SEC-03, AC-SEC-01 |
| Retention purge and deletion on request | M | RET-01 to RET-03 |
| Abuse controls (suspension, caps switchable, declaration, rate limits) | M | ADM-01 to ADM-04, ORG-10 |
| System sender for notifications | M | PRV-12 |
| Reusable service layer | M | API-01 |
| Public API, keys, webhooks, billing, plans, legal documents | P | API-02, COM-04 |
