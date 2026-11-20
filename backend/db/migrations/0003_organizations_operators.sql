-- +goose Up
CREATE TABLE organizations (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    logo_object_key text,
    default_sender_name text NOT NULL DEFAULT '',
    default_reply_to text,
    time_zone text NOT NULL DEFAULT 'UTC',
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended')),
    require_approval boolean NOT NULL DEFAULT false,
    retention_days integer NOT NULL DEFAULT 365 CHECK (retention_days > 0),
    storage_quota_bytes bigint NOT NULL DEFAULT 21474836480 CHECK (storage_quota_bytes > 0),
    caps_enabled boolean NOT NULL DEFAULT false,
    daily_email_cap integer CHECK (daily_email_cap IS NULL OR daily_email_cap > 0),
    campaign_recipient_cap integer CHECK (campaign_recipient_cap IS NULL OR campaign_recipient_cap > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE operators (
    id uuid PRIMARY KEY,
    email text NOT NULL,
    email_normalized text GENERATED ALWAYS AS (lower(btrim(email))) STORED,
    password_hash text NOT NULL,
    totp_seed_encrypted bytea,
    totp_key_id text,
    totp_enrolled_at timestamptz,
    totp_last_used_step bigint,
    recovery_code_hashes text[] NOT NULL DEFAULT '{}',
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended')),
    failed_login_count integer NOT NULL DEFAULT 0 CHECK (failed_login_count >= 0),
    locked_until timestamptz,
    last_login_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX operators_email_normalized_key ON operators (email_normalized);

ALTER TABLE organizations OWNER TO email_owner;
ALTER TABLE operators OWNER TO email_owner;

REVOKE ALL ON organizations, operators FROM PUBLIC;
GRANT SELECT, UPDATE ON organizations TO email_app;
GRANT ALL ON organizations TO email_system;
GRANT ALL ON operators TO email_system;

ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE organizations FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON organizations TO email_app
    USING (id = current_setting('app.org_id')::uuid)
    WITH CHECK (id = current_setting('app.org_id')::uuid);
CREATE POLICY system_all ON organizations TO email_system
    USING (true) WITH CHECK (true);

ALTER TABLE operators ENABLE ROW LEVEL SECURITY;
ALTER TABLE operators FORCE ROW LEVEL SECURITY;
CREATE POLICY system_all ON operators TO email_system
    USING (true) WITH CHECK (true);

-- +goose Down
DROP TABLE IF EXISTS operators;
DROP TABLE IF EXISTS organizations;
