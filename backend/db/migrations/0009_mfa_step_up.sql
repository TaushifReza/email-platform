-- +goose Up
ALTER TABLE sessions ADD COLUMN step_up_expires_at timestamptz;

CREATE OR REPLACE FUNCTION public.auth_lookup_session(p_token_hash bytea)
RETURNS jsonb
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
    SELECT COALESCE((
        SELECT jsonb_build_object(
            'session_id', s.id,
            'organization_id', s.organization_id,
            'user_id', s.user_id,
            'mfa_verified', s.mfa_verified,
            'revoked_at', s.revoked_at,
            'idle_expires_at', s.idle_expires_at,
            'absolute_expires_at', s.absolute_expires_at,
            'step_up_expires_at', s.step_up_expires_at,
            'role', u.role,
            'status', u.status
        )
        FROM public.sessions AS s
        JOIN public.users AS u
          ON u.organization_id = s.organization_id
         AND u.id = s.user_id
        WHERE s.token_hash = p_token_hash
        LIMIT 1
    ), 'null'::jsonb)
$$;

ALTER FUNCTION public.auth_lookup_session(bytea) OWNER TO email_system;
REVOKE ALL ON FUNCTION public.auth_lookup_session(bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.auth_lookup_session(bytea) TO email_app;

-- +goose Down
DROP FUNCTION IF EXISTS public.auth_lookup_session(bytea);
ALTER TABLE sessions DROP COLUMN IF EXISTS step_up_expires_at;