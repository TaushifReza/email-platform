-- +goose Up
CREATE TABLE system_emails (
    id uuid PRIMARY KEY,
    organization_id uuid NOT NULL,
    user_id uuid,
    kind text NOT NULL CHECK (kind IN ('invitation', 'password_reset')),
    recipient_email text NOT NULL,
    subject text NOT NULL,
    body_text text NOT NULL,
    body_html text,
    status text NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'sending', 'sent', 'failed')),
    attempt_count integer NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    lease_expires_at timestamptz,
    sent_at timestamptz,
    last_error text,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (organization_id, id),
    FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE RESTRICT,
    FOREIGN KEY (organization_id, user_id) REFERENCES users (organization_id, id) ON DELETE CASCADE
);
CREATE INDEX system_emails_status_due_idx ON system_emails (status, next_attempt_at)
    WHERE status IN ('queued', 'sending');

ALTER TABLE system_emails OWNER TO email_owner;

REVOKE ALL ON system_emails FROM PUBLIC;
GRANT SELECT, INSERT ON system_emails TO email_app;
GRANT ALL ON system_emails TO email_system;

ALTER TABLE system_emails ENABLE ROW LEVEL SECURITY;
ALTER TABLE system_emails FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON system_emails TO email_app
    USING (organization_id = current_setting('app.org_id')::uuid)
    WITH CHECK (organization_id = current_setting('app.org_id')::uuid);
CREATE POLICY system_all ON system_emails TO email_system
    USING (true) WITH CHECK (true);

-- SECURITY DEFINER lookups (OQ-12 pattern) for pre-auth token resolution.
CREATE OR REPLACE FUNCTION public.auth_lookup_invitation(p_token_hash bytea)
RETURNS jsonb
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
    SELECT COALESCE((
        SELECT jsonb_build_object(
            'invitation_id', i.id,
            'organization_id', i.organization_id,
            'user_id', i.accepted_user_id,
            'email', i.email,
            'role', i.role,
            'expires_at', i.expires_at,
            'accepted_at', i.accepted_at,
            'revoked_at', i.revoked_at,
            'invited_by_user_id', i.invited_by_user_id
        )
        FROM public.invitations AS i
        WHERE i.token_hash = p_token_hash
        LIMIT 1
    ), 'null'::jsonb)
$$;
ALTER FUNCTION public.auth_lookup_invitation(bytea) OWNER TO email_system;
REVOKE ALL ON FUNCTION public.auth_lookup_invitation(bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.auth_lookup_invitation(bytea) TO email_app;

CREATE OR REPLACE FUNCTION public.auth_lookup_password_reset(p_token_hash bytea)
RETURNS jsonb
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
    SELECT COALESCE((
        SELECT jsonb_build_object(
            'token_id', t.id,
            'organization_id', t.organization_id,
            'user_id', t.user_id,
            'expires_at', t.expires_at,
            'used_at', t.used_at
        )
        FROM public.auth_tokens AS t
        WHERE t.token_hash = p_token_hash
          AND t.purpose = 'password_reset'
        LIMIT 1
    ), 'null'::jsonb)
$$;
ALTER FUNCTION public.auth_lookup_password_reset(bytea) OWNER TO email_system;
REVOKE ALL ON FUNCTION public.auth_lookup_password_reset(bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.auth_lookup_password_reset(bytea) TO email_app;

CREATE OR REPLACE FUNCTION public.auth_lookup_reset_target(p_email text)
RETURNS jsonb
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
    SELECT COALESCE((
        SELECT jsonb_build_object(
            'user_id', u.id,
            'organization_id', u.organization_id,
            'email', u.email,
            'status', u.status
        )
        FROM public.users AS u
        WHERE u.email_normalized = lower(btrim(p_email))
        LIMIT 1
    ), 'null'::jsonb)
$$;
ALTER FUNCTION public.auth_lookup_reset_target(text) OWNER TO email_system;
REVOKE ALL ON FUNCTION public.auth_lookup_reset_target(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.auth_lookup_reset_target(text) TO email_app;

-- +goose Down
DROP FUNCTION IF EXISTS public.auth_lookup_reset_target(text);
DROP FUNCTION IF EXISTS public.auth_lookup_password_reset(bytea);
DROP FUNCTION IF EXISTS public.auth_lookup_invitation(bytea);
DROP TABLE IF EXISTS system_emails;