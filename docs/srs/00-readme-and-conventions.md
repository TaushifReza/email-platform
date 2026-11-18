# Software Requirements Specification
## Multi-Tenant Email Automation and Personalized Bulk Email Platform

| | |
|---|---|
| **Document status** | Draft 0.2, reviewed and approved by the client on 2026-11-17 (decisions OD-01 and OD-07 resolved; analyst assumptions accepted as defaults) |
| **Date** | 2026-11-17 |
| **Phase** | Requirements complete for MVP; technology, architecture, schema and API design intentionally NOT decided |
| **Audience** | Product owner (client), and AI coding agents (Claude Code, Codex, OpenCode, others) |

---

## How to read this document (instructions for implementing agents)

1. **This SRS defines behavior, not technology.** No language, framework, database, queue, cloud or library is chosen. Choose the stack in a separate design step and record it separately. Do not infer a stack from this document.
2. **Do not invent business rules.** If a behavior is not specified here, ask the product owner or record it as an open question. Never silently assume an important rule.
3. **Requirement IDs.** Every requirement has an ID such as `TPL-03`. Use IDs in code comments, tests, commit messages and task breakdowns.
4. **Tags on each requirement** use the form `[Priority·Basis]`:
   - Priority: **M** = must have for MVP, **S** = should have (build if capacity allows; the MVP may ship without it), **P** = post-MVP (do not build now, but do not block it architecturally).
   - Basis: **C** = explicitly confirmed by the client, **A** = assumption proposed by the analyst and not yet explicitly confirmed. Treat **A** items as defaults that the client may change. They are listed again in section 34. On 2026-11-17 the client reviewed the SRS, approved it without changes and accepted all **A** items as the working defaults (changeable on request).
5. **Normative words.** MUST, MUST NOT, SHOULD and MAY are used in their usual requirement sense.
6. **Precedence.** Section 9 (Business Rules) and section 30 (Acceptance Criteria) win over any looser wording elsewhere.
7. **Architectural constraints that are requirements, not design choices:** a reusable service layer for all business operations (API-01), tenant scoping enforced centrally (SEC-03), and a "never send the same email twice" guarantee (SND-06).

---

## Table of contents

1. Product overview
2. Problem statement
3. Goals
4. Non-goals
5. Target users
6. User roles
7. Tenant / organization model
8. Functional requirements (platform-level)
9. Business rules
10. User workflows
11. Email lifecycle
12. Template requirements
13. Personalization requirements
14. Recipient-data requirements
15. Attachment requirements
16. Campaign requirements
17. Scheduling requirements
18. Sending requirements
19. Retry and failure requirements
20. Tracking requirements
21. Reporting requirements
22. Notification requirements
23. Audit requirements
24. Security requirements
25. Compliance considerations
26. API requirements
27. Scalability requirements
28. Data-retention requirements
29. Edge cases
30. Acceptance criteria
31. MVP scope
32. Post-MVP scope
33. Future considerations
34. Open decisions and assumptions
35. Risks and mitigations
36. Glossary

---

