-- name: ListProviders :many
SELECT id, organization_id, host, port, tls_mode, username, from_address, is_default,
       status, daily_limit, minute_limit, hour_limit, config_version, verified_at,
       created_at, updated_at
FROM email_providers
WHERE organization_id = sqlc.arg(organization_id)::uuid
ORDER BY is_default DESC, created_at ASC, id ASC;

-- name: GetProvider :one
SELECT id, organization_id, host, port, tls_mode, username, from_address, is_default,
       status, daily_limit, minute_limit, hour_limit, config_version, verified_at,
       created_at, updated_at
FROM email_providers
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND id = sqlc.arg(id)::uuid;

-- name: GetProviderCredentials :one
SELECT username, password_encrypted, password_key_id
FROM email_providers
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND id = sqlc.arg(id)::uuid;

-- name: CreateProvider :one
INSERT INTO email_providers (
    id, organization_id, host, port, tls_mode, username, password_encrypted,
    password_key_id, from_address, is_default, status, daily_limit, minute_limit,
    hour_limit, config_version, verified_at
) VALUES (
    sqlc.arg(id)::uuid,
    sqlc.arg(organization_id)::uuid,
    sqlc.arg(host)::text,
    sqlc.arg(port)::int,
    sqlc.arg(tls_mode)::text,
    sqlc.arg(username)::text,
    sqlc.arg(password_encrypted)::bytea,
    sqlc.arg(password_key_id)::text,
    sqlc.arg(from_address)::text,
    sqlc.arg(is_default)::boolean,
    sqlc.arg(status)::text,
    sqlc.arg(daily_limit)::int,
    sqlc.arg(minute_limit)::int,
    sqlc.arg(hour_limit)::int,
    1,
    NULL
)
RETURNING id, organization_id, host, port, tls_mode, username, from_address, is_default,
          status, daily_limit, minute_limit, hour_limit, config_version, verified_at,
          created_at, updated_at;

-- name: UpdateProviderConfig :one
UPDATE email_providers
SET host = sqlc.arg(host)::text,
    port = sqlc.arg(port)::int,
    tls_mode = sqlc.arg(tls_mode)::text,
    from_address = sqlc.arg(from_address)::text,
    daily_limit = sqlc.arg(daily_limit)::int,
    minute_limit = sqlc.arg(minute_limit)::int,
    hour_limit = sqlc.arg(hour_limit)::int,
    config_version = config_version + 1,
    status = 'needs_attention',
    verified_at = NULL,
    updated_at = now()
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND id = sqlc.arg(id)::uuid
  AND status <> 'deactivated'
RETURNING id, organization_id, host, port, tls_mode, username, from_address, is_default,
          status, daily_limit, minute_limit, hour_limit, config_version, verified_at,
          created_at, updated_at;

-- name: RotateProviderSecret :one
UPDATE email_providers
SET username = sqlc.arg(username)::text,
    password_encrypted = sqlc.arg(password_encrypted)::bytea,
    password_key_id = sqlc.arg(password_key_id)::text,
    config_version = config_version + 1,
    status = 'needs_attention',
    verified_at = NULL,
    updated_at = now()
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND id = sqlc.arg(id)::uuid
  AND status <> 'deactivated'
RETURNING id, organization_id, host, port, tls_mode, username, from_address, is_default,
          status, daily_limit, minute_limit, hour_limit, config_version, verified_at,
          created_at, updated_at;

-- name: DeactivateProvider :one
UPDATE email_providers
SET status = 'deactivated',
    is_default = false,
    config_version = config_version + 1,
    updated_at = now()
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND id = sqlc.arg(id)::uuid
  AND status <> 'deactivated'
RETURNING id, organization_id, host, port, tls_mode, username, from_address, is_default,
          status, daily_limit, minute_limit, hour_limit, config_version, verified_at,
          created_at, updated_at;

-- name: ClearProviderDefault :exec
UPDATE email_providers
SET is_default = false, updated_at = now()
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND is_default = true;

-- name: SetProviderDefault :one
UPDATE email_providers
SET is_default = true, updated_at = now()
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND id = sqlc.arg(id)::uuid
  AND status <> 'deactivated'
RETURNING id, organization_id, host, port, tls_mode, username, from_address, is_default,
          status, daily_limit, minute_limit, hour_limit, config_version, verified_at,
          created_at, updated_at;

-- name: CountProviders :one
SELECT COUNT(*)::int
FROM email_providers
WHERE organization_id = sqlc.arg(organization_id)::uuid;
