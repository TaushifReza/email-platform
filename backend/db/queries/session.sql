-- name: SetAppRole :exec
SET LOCAL ROLE email_app;

-- name: SetOrgID :exec
SELECT set_config('app.org_id', sqlc.arg(org_id)::text, true);
