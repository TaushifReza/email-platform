# Plan: Operator auth moved to frontend + UI overhaul

## Goal
Replace Cloudflare-Access/header-based operator identity with a first-class frontend operator login flow, while preserving the operator allowlist and system DB isolation. Also fix the broken visual theme (invisible input text) and improve overall UI.

## Phase 1 — Operator authentication on app side
1. Add operator login contract:
   - `POST /auth/operator/login` (email + password or shared secret? preferred: operator password in separate `operator_credentials` table, argon2id, MFA optional/required).
   - Reuse existing session/cookie/CSRF machinery but mark session as operator context.
   - `GET /operator/session` resolves from session, not trusted header.
2. Remove trusted-proxy header auth from operator routes; keep header path only as optional dev shim if needed, but default to session cookie.
3. Add operator logout.
4. Frontend:
   - `/operator/login` page with form.
   - On success redirect to `/admin` (or `/operator`).
   - Guard `/admin` route: if no operator session, redirect to `/operator/login`.
   - Operator service calls use credentials/cookies/CSRF like tenant flows.
5. CLI:
   - `operator create` also sets initial password or prints one-time setup link.
   - Add `operator reset-password` maybe.

## Phase 2 — Security/docs
- Update `docs/technical-design/04-security.md` operator trust boundary.
- Update `docs/technical-design/01-architecture.md` and `05-operations.md`.
- Update `docs/tasks/WP02-operator.md`.
- Update `docs/OPEN_QUESTIONS.md` OQ-16/OQ-17.
- Update decision D-13 if Cloudflare Access no longer required (or keep as optional edge layer).
- Keep `SYSTEM_DATABASE_URL` isolation and audit.

## Phase 3 — UI overhaul
Root cause: `globals.css` lacks shadcn CSS variables (`--background`, `--foreground`, `--input`, `--muted-foreground`, etc.) and Tailwind config does not map them. Input text becomes invisible.

Fix:
1. Define full shadcn color token set in `:root` (and dark mode).
2. Map tokens in `tailwind.config.js` under `theme.extend.colors`.
3. Set body font, focus rings, card, table, alert styles.
4. Improve layout: centered auth cards, responsive grids, consistent spacing, readable tables.
5. Add loading/empty/error states.
6. Test with `npm run build`, `npm test`, visual check.

## Risks
- Operator auth becomes app responsibility; must enforce MFA for operators.
- Migration for operator credentials.
- Need maintain backward compatibility during transition.

## Immediate next sub-task
Start with CSS token fix + operator login backend contract.