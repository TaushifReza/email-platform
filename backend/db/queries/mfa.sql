-- name: GetMFAState :one
SELECT totp_seed_encrypted, totp_key_id, totp_enrolled_at, totp_last_used_step,
       recovery_code_hashes, locked_until, failed_login_count
FROM users
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND id = sqlc.arg(id)::uuid;

-- name: SetTOTPSeed :execrows
UPDATE users
SET totp_seed_encrypted = sqlc.arg(seed)::bytea,
    totp_key_id = sqlc.arg(key_id)::text,
    updated_at = now()
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND id = sqlc.arg(id)::uuid
  AND totp_enrolled_at IS NULL
  AND status = 'active';

-- name: ConfirmTOTPEnrollment :execrows
UPDATE users
SET totp_enrolled_at = sqlc.arg(now)::timestamptz,
    recovery_code_hashes = sqlc.arg(codes)::text[],
    updated_at = sqlc.arg(now)::timestamptz
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND id = sqlc.arg(id)::uuid
  AND totp_enrolled_at IS NULL
  AND totp_seed_encrypted IS NOT NULL;

-- name: ConsumeTOTPStep :execrows
UPDATE users
SET totp_last_used_step = sqlc.arg(step)::bigint,
    updated_at = now()
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND id = sqlc.arg(id)::uuid
  AND (totp_last_used_step IS NULL OR totp_last_used_step < sqlc.arg(step)::bigint);

-- name: ConsumeRecoveryCode :execrows
UPDATE users
SET recovery_code_hashes = array_remove(recovery_code_hashes, sqlc.arg(hash)::text),
    updated_at = now()
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND id = sqlc.arg(id)::uuid
  AND sqlc.arg(hash)::text = ANY(recovery_code_hashes);

-- name: MarkSessionMFAVerified :execrows
UPDATE sessions
SET mfa_verified = true,
    last_seen_at = sqlc.arg(now)::timestamptz
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND id = sqlc.arg(id)::uuid
  AND revoked_at IS NULL;

-- name: SetSessionStepUp :execrows
UPDATE sessions
SET step_up_expires_at = sqlc.arg(until)::timestamptz,
    last_seen_at = sqlc.arg(now)::timestamptz
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND id = sqlc.arg(id)::uuid
  AND revoked_at IS NULL
  AND mfa_verified = true;
