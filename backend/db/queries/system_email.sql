-- name: EnqueueSystemEmail :exec
INSERT INTO system_emails (id, organization_id, user_id, kind, recipient_email, subject, body_text, body_html, next_attempt_at)
VALUES (sqlc.arg(id)::uuid, sqlc.arg(organization_id)::uuid, sqlc.narg(user_id)::uuid, sqlc.arg(kind)::text, sqlc.arg(recipient_email)::text, sqlc.arg(subject)::text, sqlc.arg(body_text)::text, sqlc.narg(body_html)::text, sqlc.arg(next_attempt_at)::timestamptz);

-- name: ClaimSystemEmails :many
WITH claimed AS (
    SELECT id
      FROM system_emails
     WHERE (status = 'queued' AND next_attempt_at <= now())
        OR (status = 'sending' AND lease_expires_at IS NOT NULL AND lease_expires_at < now())
     ORDER BY next_attempt_at
     LIMIT sqlc.arg(row_limit)::int
     FOR UPDATE SKIP LOCKED
)
UPDATE system_emails AS se
   SET status = 'sending',
       attempt_count = se.attempt_count + 1,
       lease_expires_at = sqlc.arg(lease_expires_at)::timestamptz
  FROM claimed
 WHERE se.id = claimed.id
RETURNING se.id, se.organization_id, se.user_id, se.kind, se.recipient_email,
          se.subject, se.body_text, se.body_html, se.attempt_count;

-- name: MarkSystemEmailSent :exec
UPDATE system_emails
SET status = 'sent',
    sent_at = now(),
    lease_expires_at = NULL,
    last_error = NULL
WHERE id = sqlc.arg(id)::uuid;

-- name: MarkSystemEmailFailed :exec
UPDATE system_emails
SET status = CASE WHEN sqlc.arg(attempt_count)::int >= sqlc.arg(max_attempts)::int THEN 'failed' ELSE 'queued' END,
    next_attempt_at = CASE WHEN sqlc.arg(attempt_count)::int >= sqlc.arg(max_attempts)::int THEN next_attempt_at ELSE sqlc.arg(next_attempt_at)::timestamptz END,
    lease_expires_at = NULL,
    last_error = sqlc.arg(last_error)::text
WHERE id = sqlc.arg(id)::uuid;