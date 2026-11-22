-- name: CreateAuthToken :exec
INSERT INTO auth_tokens (id, organization_id, user_id, purpose, token_hash, expires_at)
VALUES (sqlc.arg(id)::uuid, sqlc.arg(organization_id)::uuid, sqlc.arg(user_id)::uuid, 'password_reset', sqlc.arg(token_hash)::bytea, sqlc.arg(expires_at)::timestamptz);

-- name: LookupPasswordResetByToken :one
SELECT auth_lookup_password_reset(sqlc.arg(token_hash)::bytea) AS payload;

-- name: ConsumePasswordResetToken :execrows
UPDATE auth_tokens
SET used_at = sqlc.arg(used_at)::timestamptz
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND id = sqlc.arg(id)::uuid
  AND purpose = 'password_reset'
  AND used_at IS NULL
  AND expires_at > sqlc.arg(used_at)::timestamptz;

-- name: SetUserPassword :execrows
UPDATE users
SET password_hash = sqlc.arg(password_hash)::text,
    password_changed_at = sqlc.arg(password_changed_at)::timestamptz,
    failed_login_count = 0,
    locked_until = NULL,
    updated_at = now()
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND id = sqlc.arg(id)::uuid;

-- name: RevokeAllUserSessions :exec
UPDATE sessions
SET revoked_at = sqlc.arg(revoked_at)::timestamptz
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND user_id = sqlc.arg(user_id)::uuid
  AND revoked_at IS NULL;

-- name: LookupResetTarget :one
SELECT auth_lookup_reset_target(sqlc.arg(email)::text) AS payload;