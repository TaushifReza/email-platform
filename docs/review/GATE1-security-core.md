# Gate 1: security core (end of week 1 to 2)
Do not start building on top until this passes.

## Automated
- [ ] CI green including govulncheck, gosec, secret scan
- [ ] Catalog guard test passes (every tenant table: forced RLS, policy, composite FKs)
- [ ] AC-SEC-01 cross-tenant tests, role matrix test, AC-SEC-02 to AC-SEC-05, AC-PRV-01, AC-PRV-02 pass
- [ ] Planted-secret test: a fake secret never appears in any log line or API response

## Read personally (or by a stronger separate model, with a written threat brief)
- [ ] `internal/platform/crypto`: nonce handling, associated data, key ids, no key in logs
- [ ] `internal/platform/db` WithOrg and RLS policies: no code path runs a tenant query outside WithOrg
- [ ] `internal/identity`: session token hashing, cookie flags, CSRF check, lockout, 2FA replay protection
- [ ] `internal/providers/dialer`: the IP check happens inside the connect path, covers IPv6 and mapped addresses
- [ ] Trusted proxy handling: client IP taken only from nginx

## Click-through (dev stack)
1. Operator creates an organization, Owner accepts the invitation (email appears in Mailpit), enrolls 2FA.
2. Owner invites an Admin and a Viewer. Viewer cannot open provider settings.
3. Owner adds an SMTP provider pointing at a private address: refused. A real provider: verified with test email.
4. Rotate the provider password: old value unusable, never displayed.
5. Look at the audit log: every step above is there.
