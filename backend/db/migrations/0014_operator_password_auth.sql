-- +goose Up
-- Operator frontend authentication: allow app-level login against the operator allowlist.
-- Password hash stays nullable for legacy Cloudflare-Access-created rows; frontend login
-- requires a password to be set via the CLI/bootstrap flow.

CREATE OR REPLACE FUNCTION public.auth_lookup_operator_credentials(p_email text)
RETURNS jsonb
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
    SELECT COALESCE((
        SELECT jsonb_build_object(
            'id', o.id,
            'email', o.email,
            'status', o.status,
            'password_hash', o.password_hash,
            'failed_login_count', o.failed_login_count,
            'locked_until', o.locked_until
        )
        FROM public.operators AS o
        WHERE o.email_normalized = lower(btrim(p_email))
        LIMIT 1
    ), 'null'::jsonb)
$$;

ALTER FUNCTION public.auth_lookup_operator_credentials(text) OWNER TO email_system;
REVOKE ALL ON FUNCTION public.auth_lookup_operator_credentials(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.auth_lookup_operator_credentials(text) TO email_app;

-- +goose Down
DROP FUNCTION IF EXISTS public.auth_lookup_operator_credentials(text);