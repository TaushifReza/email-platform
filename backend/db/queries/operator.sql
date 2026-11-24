-- name: LookupOperatorByEmail :one
SELECT auth_lookup_operator(sqlc.arg(email)::text) AS payload;

-- name: LookupOperatorCredentials :one
SELECT auth_lookup_operator_credentials(sqlc.arg(email)::text) AS payload;

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
