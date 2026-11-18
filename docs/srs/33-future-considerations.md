## 33. Future considerations

Features that should not be built now but may influence design:

- **Provider abstraction** must allow API-based providers, OAuth-based providers and webhook-based events.
- **Per-email Message-ID and provider response** are stored so bounces and events can be matched later.
- **Recipient states** must be extensible with Delivered, Bounced, Opened and Clicked without breaking reports.
- **Suppression** may later gain a platform layer and automatic population from bounces and complaints (scope by campaign type is already in the MVP, SUP-05).
- **Commercialization:** plans, quotas, metering, tenant self-signup, providing a managed sender (for example SES) to tenants who have none. Providing a managed sender makes the operator the account owner with all the reputation and abuse consequences, so it must be a conscious later decision.
- **Contact lists** are a possible future feature but conflict with the current non-goal.
- **Multi-organization user membership** was declined for now. A later change affects the identity model.
- **Localization** of the interface and templates, multi-language recipients.
- **100,000-recipient campaigns** and multi-day sending.
- **Data residency** options per organization.
