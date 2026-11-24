-- +goose Up
CREATE TABLE email_providers (
    id uuid PRIMARY KEY,
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    host text NOT NULL CHECK (host <> ''),
    port integer NOT NULL CHECK (port IN (25, 465, 587, 2525)),
    tls_mode text NOT NULL CHECK (tls_mode IN ('starttls', 'implicit')),
    username text NOT NULL,
    password_encrypted bytea NOT NULL,
    password_key_id text NOT NULL,
    from_address text NOT NULL CHECK (from_address <> ''),
    is_default boolean NOT NULL DEFAULT false,
    status text NOT NULL DEFAULT 'needs_attention' CHECK (status IN ('active', 'needs_attention', 'deactivated')),
    daily_limit integer NOT NULL DEFAULT 500 CHECK (daily_limit > 0),
    minute_limit integer NOT NULL DEFAULT 20 CHECK (minute_limit > 0),
    hour_limit integer NOT NULL DEFAULT 1200 CHECK (hour_limit > 0),
    config_version integer NOT NULL DEFAULT 1 CHECK (config_version > 0),
    verified_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (organization_id, id)
);
CREATE UNIQUE INDEX email_providers_default_key ON email_providers (organization_id) WHERE is_default;
CREATE INDEX email_providers_org_idx ON email_providers (organization_id);

ALTER TABLE email_providers OWNER TO email_owner;

REVOKE ALL ON email_providers FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE ON email_providers TO email_app;
GRANT ALL ON email_providers TO email_system;

ALTER TABLE email_providers ENABLE ROW LEVEL SECURITY;
ALTER TABLE email_providers FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON email_providers TO email_app
    USING (organization_id = current_setting('app.org_id')::uuid)
    WITH CHECK (organization_id = current_setting('app.org_id')::uuid);
CREATE POLICY system_all ON email_providers TO email_system
    USING (true) WITH CHECK (true);

-- +goose Down
DROP TABLE IF EXISTS email_providers;
