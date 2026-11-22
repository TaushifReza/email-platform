-- name: AppendAuditEntry :exec
INSERT INTO audit_log (
    id, organization_id, actor_type, actor_id, actor_role, action,
    target_type, target_id, outcome, source_ip, client, details, created_at
) VALUES (
    sqlc.arg(id)::uuid,
    sqlc.arg(organization_id)::uuid,
    sqlc.arg(actor_type)::text,
    sqlc.arg(actor_id)::uuid,
    sqlc.arg(actor_role)::text,
    sqlc.arg(action)::text,
    sqlc.arg(target_type)::text,
    sqlc.arg(target_id)::uuid,
    sqlc.arg(outcome)::text,
    sqlc.arg(source_ip)::inet,
    sqlc.arg(client)::text,
    sqlc.arg(details)::jsonb,
    sqlc.arg(created_at)::timestamptz
);