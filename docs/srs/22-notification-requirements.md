## 22. Notification requirements

- **NTF-01 [M·C]** Notifications are sent by **email through the system sender** and also **shown in the app**.
- **NTF-02 [M·C]** Events and recipients (C unless noted):

| Event | Recipients |
|---|---|
| Campaign completed | Creator |
| Campaign completed with errors, or failed | Creator and Owner |
| Campaign paused automatically (authentication failure, connection failures, high failure rate) | Creator and Owner |
| Campaign needs attention (missed schedule, credentials changed, approval not given in time) | Creator and Owner |
| Campaign awaiting approval | Approvers |
| Campaign approved or rejected | Creator |
| Provider daily limit reached, campaign continues tomorrow | Creator |
| Scheduled campaign starting (**S**) | Creator |
| Support access granted, used or expired | Owner |
| Storage quota nearly full, suggested at 80% (A) | Owner and Admin |

- **NTF-03 [P·C]** Per-user notification preferences are post-MVP.
- **NTF-04 [M·C]** Notification emails MUST NOT contain recipient personal data beyond counts, and MUST NOT contain secrets.
