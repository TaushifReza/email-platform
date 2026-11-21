-- +goose Up
CREATE OR REPLACE FUNCTION public.auth_lookup_user(p_email text)
RETURNS jsonb
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
    SELECT COALESCE((
        SELECT jsonb_build_object(
            'id', u.id,
            'organization_id', u.organization_id,
            'password_hash', u.password_hash,
            'status', u.status,
            'role', u.role,
            'failed_login_count', u.failed_login_count,
            'locked_until', u.locked_until,
            'totp_enrolled_at', u.totp_enrolled_at
        )
        FROM public.users AS u
        WHERE u.email_normalized = lower(btrim(p_email))
        LIMIT 1
    ), 'null'::jsonb)
$$;

ALTER FUNCTION public.auth_lookup_user(text) OWNER TO email_system;
REVOKE ALL ON FUNCTION public.auth_lookup_user(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.auth_lookup_user(text) TO email_app;

-- +goose Down
DROP FUNCTION IF EXISTS public.auth_lookup_user(text);
