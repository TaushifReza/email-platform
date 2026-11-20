-- +goose Up
CREATE TABLE users (
    id uuid PRIMARY KEY,
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    email text NOT NULL,
    email_normalized text GENERATED ALWAYS AS (lower(btrim(email))) STORED,
    display_name text NOT NULL DEFAULT '',
    role text NOT NULL CHECK (role IN ('owner', 'admin', 'campaign_manager', 'reviewer', 'viewer')),
    status text NOT NULL DEFAULT 'invited' CHECK (status IN ('invited', 'active', 'deactivated')),
    password_hash text,
    password_changed_at timestamptz,
    totp_seed_encrypted bytea,
    totp_key_id text,
    totp_enrolled_at timestamptz,
    totp_last_used_step bigint,
    recovery_code_hashes text[] NOT NULL DEFAULT '{}',
    failed_login_count integer NOT NULL DEFAULT 0 CHECK (failed_login_count >= 0),
    locked_until timestamptz,
    last_login_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (organization_id, id)
);
CREATE UNIQUE INDEX users_email_normalized_key ON users (email_normalized);
CREATE INDEX users_organization_id_idx ON users (organization_id);

CREATE TABLE sessions (
    id uuid PRIMARY KEY,
    organization_id uuid NOT NULL,
    user_id uuid NOT NULL,
    token_hash bytea NOT NULL UNIQUE,
    mfa_verified boolean NOT NULL DEFAULT false,
    ip inet,
    user_agent text,
    created_at timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    idle_expires_at timestamptz NOT NULL,
    absolute_expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    UNIQUE (organization_id, id),
    FOREIGN KEY (organization_id, user_id) REFERENCES users (organization_id, id) ON DELETE CASCADE
);
CREATE INDEX sessions_user_id_idx ON sessions (user_id);
CREATE INDEX sessions_expiry_idx ON sessions (idle_expires_at, absolute_expires_at);

CREATE TABLE operator_sessions (
    id uuid PRIMARY KEY,
    operator_id uuid NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    token_hash bytea NOT NULL UNIQUE,
    mfa_verified boolean NOT NULL DEFAULT false,
    ip inet,
    user_agent text,
    created_at timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    idle_expires_at timestamptz NOT NULL,
    absolute_expires_at timestamptz NOT NULL,
    revoked_at timestamptz
);
CREATE INDEX operator_sessions_operator_id_idx ON operator_sessions (operator_id);
CREATE INDEX operator_sessions_expiry_idx ON operator_sessions (idle_expires_at, absolute_expires_at);

ALTER TABLE users OWNER TO email_owner;
ALTER TABLE sessions OWNER TO email_owner;
ALTER TABLE operator_sessions OWNER TO email_owner;

REVOKE ALL ON users, sessions, operator_sessions FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE ON users TO email_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON sessions TO email_app;
GRANT SELECT, UPDATE ON users TO email_system;
GRANT ALL ON sessions TO email_system;
GRANT ALL ON operator_sessions TO email_system;

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE users FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON users TO email_app
    USING (organization_id = current_setting('app.org_id')::uuid)
    WITH CHECK (organization_id = current_setting('app.org_id')::uuid);
CREATE POLICY system_all ON users TO email_system
    USING (true) WITH CHECK (true);

ALTER TABLE sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE sessions FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON sessions TO email_app
    USING (organization_id = current_setting('app.org_id')::uuid)
    WITH CHECK (organization_id = current_setting('app.org_id')::uuid);
CREATE POLICY system_all ON sessions TO email_system
    USING (true) WITH CHECK (true);

ALTER TABLE operator_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE operator_sessions FORCE ROW LEVEL SECURITY;
CREATE POLICY system_all ON operator_sessions TO email_system
    USING (true) WITH CHECK (true);

-- +goose Down
DROP TABLE IF EXISTS operator_sessions;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
