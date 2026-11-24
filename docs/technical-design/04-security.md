# Security design

## Keys and secrets
Master key (256-bit) as a Docker secret; backup in the owner's password manager plus an offline copy. HKDF-derived purpose keys: provider passwords, 2FA seeds, unsubscribe tokens, file encryption if used. AES-256-GCM, fresh random nonce per value, associated data = table, row id, column, organization. Key id stored with each ciphertext. Rotation = new key id + re-encrypt job. This protects a database leak or backup theft, not a compromised app server.

## SMTP connection safety
1. At configuration: resolve hostname, reject if **any** address is non-public (private, loopback, link-local incl. metadata, CGNAT, multicast, reserved, IPv6 equivalents, IPv4-mapped). Ports 25, 465, 587, 2525 only.
2. At every connection: repeat the check **inside the dialer's Control function on the actual address** (defeats DNS rebinding).
3. TLS mandatory (STARTTLS or implicit). A server without STARTTLS is refused. Certificates and hostnames verified. Minimum TLS 1.2. No skip-verify option ever.
4. Generic error messages ("host not allowed").
5. Defense in depth: host firewall blocks worker egress to private ranges except postgres, redis, rustfs.

## Authentication and sessions
argon2id; minimum 12 characters plus built-in common-password list (breached lookup is M1). TOTP with replay protection, 10 hashed single-use recovery codes, enrollment forced at first login for Owner, Admin, Platform Admin. Session: random 256-bit token stored hashed; cookie `__Host-`, HttpOnly, Secure, SameSite; privileged idle 30 min and absolute 12 h, others idle 2 h and absolute 24 h. CSRF: SameSite + required custom header + Origin check. Step-up (password + 2FA) before: rotating provider credentials, changing roles, deleting recipient data, deactivating users, granting support access. Lockout and Redis rate limits per account and IP, uniform responses (no account enumeration). Reset links 30 min, invitations 72 h, hash stored, single use; reset ends all sessions. Operator console supports frontend operator login via app-issued sessions; Cloudflare Access header remains a transitional fallback.

## Operator trust boundary
Cloudflare Access is the operator identity provider (OQ-16). On `admin.<domain>` it authenticates the human (MFA enforced there) and injects `Cf-Access-Authenticated-User-Email`. The app trusts that header only when the immediate socket peer is in `TRUSTED_PROXY_CIDRs`; `X-Forwarded-For` is never used to establish this trust, and the tenant host explicitly strips the header. The email is normalized and resolved through `auth_lookup_operator` (`SECURITY DEFINER`, `email_app` gets execute only, active rows only, no secret columns). Unknown or suspended operators get the same 403 with no enumeration detail. Cross-organization operator writes run on a separate `SYSTEM_DATABASE_URL` connection as `email_system`; `email_app` must not be a member of that role because RLS policies granted to it would then be inherited and tenant isolation would fail. Operator authorization reuses the platform-admin row of the SRS 6.2 matrix. Dev uses an nginx shim on `admin.localhost` that sets the header and a CLI-created operator row. Before staging, restrict header forwarding to the authenticated tunnel or verify the CF JWT in the app (OQ-17).

## Authorization
Every service call has an Actor. Organization from the session only. One central permission function implements the SRS 6.2 matrix. Table-driven test covers every role x capability.

## Files and data
- Size limits enforced while streaming; content-type by file signature against the allow-list; double extensions rejected.
- XLSX is a ZIP: apply zip-bomb and entry limits, refuse macro files, cap rows, columns, cell size; parse within worker memory and time limits.
- ZIP: stream entries straight to object storage, never write by entry path; object keys are generated; reject encrypted, nested, symlink, traversal, extreme ratio, zero-byte; max about 10,000 entries.
- Downloads: `Content-Disposition: attachment`, `nosniff`, never inline.
- Template preview in a sandboxed iframe with strict CSP.
- HTML sanitized with an email-safe allow-list; links limited to http, https, mailto, tel.
- Renderer is single-pass, values never re-expanded. Headers reject CR/LF. Exports neutralize cells starting with `=`, `+`, `-`, `@`, tab, CR.
- Logs: redaction layer, emails masked, request bodies never logged.

## Public unsubscribe endpoint
GET `/u/{token}` shows a confirmation page and does nothing else. POST confirms. One-click uses a separate POST endpoint (`List-Unsubscribe-Post`). Token is encrypted and tamper-proof. Page has no scripts, no cookies, strict CSP, `Referrer-Policy: no-referrer`. Identical responses whether or not already suppressed. Rate limited by IP. Idempotent. Audited.

## Browser headers
HSTS, `frame-ancestors none`, nosniff, restrictive referrer and permissions policies. Next.js static export injects small inline scripts: generate script hashes at build time and list them in the CSP; if brittle, relax the script policy only for the app shell. Unsubscribe page and preview iframe keep the strict policy. Report which approach was used.

## Rate limits (initial)
Login 5/min per account and IP; 2FA 5 attempts then lockout; reset request 3/h per account; uploads 20/h per user; test sends 10/h per user; unsubscribe 30/min per IP; API 300/min per user.

## Hardening
Containers non-root, read-only filesystem, dropped capabilities, minimal images pinned by digest. Secrets only via Docker secrets. CI: govulncheck, gosec, npm audit, secret scanning, dependency updates.
