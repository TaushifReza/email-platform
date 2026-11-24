-- name: LookupOperatorByEmail :one
SELECT auth_lookup_operator(sqlc.arg(email)::text) AS payload;

-- name: LookupOperatorCredentials :one
SELECT auth_lookup_operator_credentials(sqlc.arg(email)::text) AS payload;

-- name: LookupOperatorSession :one
SELECT auth_lookup_operator_session(sqlc.arg(token_hash)::bytea) AS payload;

-- name: CreateOperatorSession :exec
INSERT INTO operator_sessions (id, operator_id, token_hash, mfa_verified, ip, user_agent, idle_expires_at, absolute_expires_at)
VALUES (sqlc.arg(id)::uuid, sqlc.arg(operator_id)::uuid, sqlc.arg(token_hash)::bytea, sqlc.arg(mfa_verified)::boolean, sqlc.arg(ip)::inet, sqlc.arg(user_agent)::text, sqlc.arg(idle_expires_at)::timestamptz, sqlc.arg(absolute_expires_at)::timestamptz);

-- name: RevokeOperatorSession :exec
UPDATE operator_sessions
SET revoked_at = sqlc.arg(now)::timestamptz
WHERE token_hash = sqlc.arg(token_hash)::bytea
  AND revoked_at IS NULL;

-- name: CreateOperator :execrows
INSERT INTO operators (id, email, status)
VALUES (sqlc.arg(id)::uuid, sqlc.arg(email)::text, 'active')
ON CONFLICT (email_normalized) DO NOTHING;

-- name: ListOperatorAuditEntries :many
SELECT id, organization_id, actor_type, actor_id, actor_role, action,
       target_type, target_id, outcome, source_ip, client, details, created_at
FROM audit_log
WHERE organization_id IS NULL
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(row_limit)::int OFFSET sqlc.arg(row_offset)::int;
