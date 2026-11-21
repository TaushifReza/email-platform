-- name: LookupUserForLogin :one
SELECT auth_lookup_user(sqlc.arg(email)::text) AS payload;

-- name: RecordFailedLogin :exec
UPDATE users
SET failed_login_count = CASE
        WHEN locked_until IS NOT NULL AND locked_until <= sqlc.arg(now)::timestamptz THEN 1
        ELSE failed_login_count + 1
    END,
    locked_until = CASE
        WHEN (CASE
                WHEN locked_until IS NOT NULL AND locked_until <= sqlc.arg(now)::timestamptz THEN 1
                ELSE failed_login_count + 1
              END) >= sqlc.arg(threshold)::int
        THEN sqlc.arg(lock_until)::timestamptz
        ELSE locked_until
    END,
    updated_at = now()
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND id = sqlc.arg(id)::uuid
  AND (locked_until IS NULL OR locked_until <= sqlc.arg(now)::timestamptz);

-- name: ClearFailedLogins :exec
UPDATE users
SET failed_login_count = 0,
    locked_until = NULL,
    updated_at = now()
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND id = sqlc.arg(id)::uuid;
