-- +goose Up
-- Operator identity is asserted by Cloudflare Access. The application stores
-- no operator password or TOTP seed; keep the legacy column nullable so the
-- bootstrap CLI can create a row without inventing a credential.
ALTER TABLE operators ALTER COLUMN password_hash DROP NOT NULL;

-- Pre-authentication lookup for the trusted identity header. Only active
-- operators are returned, and no secret columns are exposed to email_app.
CREATE OR REPLACE FUNCTION public.auth_lookup_operator(p_email text)
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
            'status', o.status
        )
        FROM public.operators AS o
        WHERE o.email_normalized = lower(btrim(p_email))
          AND o.status = 'active'
        LIMIT 1
    ), 'null'::jsonb)
$$;

ALTER FUNCTION public.auth_lookup_operator(text) OWNER TO email_system;
REVOKE ALL ON FUNCTION public.auth_lookup_operator(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.auth_lookup_operator(text) TO email_app;

-- +goose Down
DROP FUNCTION IF EXISTS public.auth_lookup_operator(text);
-- Expand-only: password_hash stays nullable until all CF-Access rows are gone.
