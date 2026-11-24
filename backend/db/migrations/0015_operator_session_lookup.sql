-- +goose Up
CREATE OR REPLACE FUNCTION public.auth_lookup_operator_session(p_token_hash bytea)
RETURNS jsonb
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
    SELECT COALESCE((
        SELECT jsonb_build_object(
            'session_id', s.id,
            'operator_id', s.operator_id,
            'mfa_verified', s.mfa_verified,
            'revoked_at', s.revoked_at,
            'idle_expires_at', s.idle_expires_at,
            'absolute_expires_at', s.absolute_expires_at
        )
        FROM public.operator_sessions AS s
        JOIN public.operators AS o ON o.id = s.operator_id
        WHERE s.token_hash = p_token_hash
          AND o.status = 'active'
        LIMIT 1
    ), 'null'::jsonb)
$$;

ALTER FUNCTION public.auth_lookup_operator_session(bytea) OWNER TO email_system;
REVOKE ALL ON FUNCTION public.auth_lookup_operator_session(bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.auth_lookup_operator_session(bytea) TO email_app;

-- +goose Down
DROP FUNCTION IF EXISTS public.auth_lookup_operator_session(bytea);