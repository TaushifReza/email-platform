-- name: SetAppRole :exec
SET LOCAL ROLE email_app;

-- name: SetOrgID :exec
SELECT set_config('app.org_id', sqlc.arg(org_id)::text, true);

-- name: CreateSession :exec
INSERT INTO sessions (id, organization_id, user_id, token_hash, mfa_verified, ip, user_agent, idle_expires_at, absolute_expires_at)
VALUES (sqlc.arg(id)::uuid, sqlc.arg(organization_id)::uuid, sqlc.arg(user_id)::uuid, sqlc.arg(token_hash)::bytea, sqlc.arg(mfa_verified)::boolean, sqlc.arg(ip)::inet, sqlc.arg(user_agent)::text, sqlc.arg(idle_expires_at)::timestamptz, sqlc.arg(absolute_expires_at)::timestamptz);

-- name: LookupSession :one
SELECT auth_lookup_session(sqlc.arg(token_hash)::bytea) AS payload;

-- name: TouchSession :exec
UPDATE sessions
SET last_seen_at = sqlc.arg(now)::timestamptz,
    idle_expires_at = sqlc.arg(idle_expires_at)::timestamptz
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND id = sqlc.arg(id)::uuid
  AND revoked_at IS NULL
  AND idle_expires_at > sqlc.arg(now)::timestamptz
  AND absolute_expires_at > sqlc.arg(now)::timestamptz
  AND last_seen_at <= sqlc.arg(now)::timestamptz - interval '1 minute';

-- name: RevokeSession :exec
UPDATE sessions
SET revoked_at = sqlc.arg(now)::timestamptz
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND token_hash = sqlc.arg(token_hash)::bytea
  AND revoked_at IS NULL;
