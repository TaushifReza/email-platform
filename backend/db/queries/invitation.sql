-- name: CreateInvitation :exec
INSERT INTO invitations (id, organization_id, email, role, invited_by_user_id, token_hash, expires_at)
VALUES (sqlc.arg(id)::uuid, sqlc.arg(organization_id)::uuid, sqlc.arg(email)::text, sqlc.arg(role)::text, sqlc.arg(invited_by_user_id)::uuid, sqlc.arg(token_hash)::bytea, sqlc.arg(expires_at)::timestamptz);

-- name: LookupInvitationByToken :one
SELECT auth_lookup_invitation(sqlc.arg(token_hash)::bytea) AS payload;

-- name: MarkInvitationAccepted :execrows
UPDATE invitations
SET accepted_user_id = sqlc.arg(user_id)::uuid,
    accepted_at = sqlc.arg(accepted_at)::timestamptz
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND id = sqlc.arg(id)::uuid
  AND accepted_at IS NULL
  AND revoked_at IS NULL
  AND expires_at > sqlc.arg(accepted_at)::timestamptz;

-- name: ActivateInvitedUser :execrows
UPDATE users
SET password_hash = sqlc.arg(password_hash)::text,
    password_changed_at = sqlc.arg(password_changed_at)::timestamptz,
    status = 'active',
    updated_at = now()
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND id = sqlc.arg(id)::uuid
  AND status = 'invited';

-- name: RevokePriorInvitations :exec
UPDATE invitations
SET revoked_at = now()
WHERE organization_id = sqlc.arg(organization_id)::uuid
  AND email_normalized = lower(btrim(sqlc.arg(email)::text))
  AND accepted_at IS NULL
  AND revoked_at IS NULL;