-- name: CreateInvitedUser :exec
INSERT INTO users (id, organization_id, email, display_name, role, status)
VALUES (sqlc.arg(id)::uuid, sqlc.arg(organization_id)::uuid,
        sqlc.arg(email)::text, sqlc.arg(display_name)::text,
        sqlc.arg(role)::text, 'invited');

-- name: ListUsers :many
SELECT id, organization_id, email, display_name, role, status,
       totp_enrolled_at, last_login_at, created_at, updated_at
FROM users
WHERE organization_id = sqlc.arg(organization_id)::uuid
ORDER BY created_at DESC, id
LIMIT sqlc.arg(row_limit)::int OFFSET sqlc.arg(row_offset)::int;

-- name: GetUser :one
SELECT id, organization_id, email, display_name, role, status,
       totp_enrolled_at, last_login_at, created_at, updated_at
FROM users
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND id = sqlc.arg(id)::uuid;

-- name: LockActiveOwners :many
SELECT id
FROM users
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND role = 'owner'
  AND status = 'active'
FOR UPDATE;

-- name: ChangeUserRole :execrows
UPDATE users
SET role = sqlc.arg(role)::text,
    updated_at = now()
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND id = sqlc.arg(id)::uuid
  AND status <> 'deactivated';

-- name: DeactivateUser :execrows
UPDATE users
SET status = 'deactivated',
    updated_at = now()
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND id = sqlc.arg(id)::uuid
  AND status <> 'deactivated';