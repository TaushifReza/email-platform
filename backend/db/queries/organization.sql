-- name: CreateOrganization :exec
INSERT INTO organizations (id, name, default_sender_name, default_reply_to, time_zone,
    require_approval, retention_days, storage_quota_bytes, caps_enabled,
    daily_email_cap, campaign_recipient_cap)
VALUES (sqlc.arg(id)::uuid, sqlc.arg(name)::text, sqlc.arg(default_sender_name)::text,
    sqlc.arg(default_reply_to)::text, sqlc.arg(time_zone)::text,
    sqlc.arg(require_approval)::boolean, sqlc.arg(retention_days)::int,
    sqlc.arg(storage_quota_bytes)::bigint, sqlc.arg(caps_enabled)::boolean,
    sqlc.arg(daily_email_cap)::int, sqlc.arg(campaign_recipient_cap)::int);

-- name: GetOrganization :one
SELECT id, name, logo_object_key, default_sender_name, default_reply_to, time_zone,
       status, require_approval, retention_days, storage_quota_bytes, caps_enabled,
       daily_email_cap, campaign_recipient_cap, created_at, updated_at
FROM organizations
WHERE id = sqlc.arg(id)::uuid;

-- name: ListOrganizations :many
SELECT id, name, logo_object_key, default_sender_name, default_reply_to, time_zone,
       status, require_approval, retention_days, storage_quota_bytes, caps_enabled,
       daily_email_cap, campaign_recipient_cap, created_at, updated_at
FROM organizations
ORDER BY created_at DESC, id
LIMIT sqlc.arg(row_limit)::int OFFSET sqlc.arg(row_offset)::int;

-- name: UpdateOrganizationProfile :one
UPDATE organizations
SET name = sqlc.arg(name)::text,
    default_sender_name = sqlc.arg(default_sender_name)::text,
    default_reply_to = sqlc.arg(default_reply_to)::text,
    time_zone = sqlc.arg(time_zone)::text,
    updated_at = now()
WHERE id = sqlc.arg(id)::uuid
RETURNING id, name, logo_object_key, default_sender_name, default_reply_to, time_zone,
          status, require_approval, retention_days, storage_quota_bytes, caps_enabled,
          daily_email_cap, campaign_recipient_cap, created_at, updated_at;

-- name: UpdateOrganizationPolicy :one
UPDATE organizations
SET require_approval = sqlc.arg(require_approval)::boolean,
    retention_days = sqlc.arg(retention_days)::int,
    updated_at = now()
WHERE id = sqlc.arg(id)::uuid
RETURNING id, name, logo_object_key, default_sender_name, default_reply_to, time_zone,
          status, require_approval, retention_days, storage_quota_bytes, caps_enabled,
          daily_email_cap, campaign_recipient_cap, created_at, updated_at;

-- name: SetOrganizationStatus :execrows
UPDATE organizations
SET status = sqlc.arg(status)::text,
    updated_at = now()
WHERE id = sqlc.arg(id)::uuid
  AND status = sqlc.arg(expected)::text;

-- name: SetOrganizationLimits :one
UPDATE organizations
SET storage_quota_bytes = sqlc.arg(storage_quota_bytes)::bigint,
    caps_enabled = sqlc.arg(caps_enabled)::boolean,
    daily_email_cap = sqlc.arg(daily_email_cap)::int,
    campaign_recipient_cap = sqlc.arg(campaign_recipient_cap)::int,
    updated_at = now()
WHERE id = sqlc.arg(id)::uuid
RETURNING id, name, logo_object_key, default_sender_name, default_reply_to, time_zone,
          status, require_approval, retention_days, storage_quota_bytes, caps_enabled,
          daily_email_cap, campaign_recipient_cap, created_at, updated_at;