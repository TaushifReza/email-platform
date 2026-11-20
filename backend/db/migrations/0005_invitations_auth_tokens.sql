-- +goose Up
CREATE TABLE invitations (
    id uuid PRIMARY KEY,
    organization_id uuid NOT NULL,
    email text NOT NULL,
    email_normalized text GENERATED ALWAYS AS (lower(btrim(email))) STORED,
    role text NOT NULL CHECK (role IN ('owner', 'admin', 'campaign_manager', 'reviewer', 'viewer')),
    invited_by_user_id uuid NOT NULL,
    accepted_user_id uuid,
    token_hash bytea NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    accepted_at timestamptz,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (organization_id, id),
    FOREIGN KEY (organization_id, invited_by_user_id) REFERENCES users (organization_id, id) ON DELETE RESTRICT,
    FOREIGN KEY (organization_id, accepted_user_id) REFERENCES users (organization_id, id) ON DELETE RESTRICT
);
CREATE INDEX invitations_organization_email_idx ON invitations (organization_id, email_normalized)
    WHERE accepted_at IS NULL AND revoked_at IS NULL;

CREATE TABLE auth_tokens (
    id uuid PRIMARY KEY,
    organization_id uuid NOT NULL,
    user_id uuid NOT NULL,
    purpose text NOT NULL CHECK (purpose IN ('password_reset', 'email_verification')),
    token_hash bytea NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    used_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (organization_id, id),
    FOREIGN KEY (organization_id, user_id) REFERENCES users (organization_id, id) ON DELETE CASCADE
);
CREATE INDEX auth_tokens_user_id_idx ON auth_tokens (user_id);
CREATE INDEX auth_tokens_expiry_idx ON auth_tokens (expires_at);

ALTER TABLE invitations OWNER TO email_owner;
ALTER TABLE auth_tokens OWNER TO email_owner;

REVOKE ALL ON invitations, auth_tokens FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE ON invitations TO email_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON auth_tokens TO email_app;
GRANT ALL ON invitations, auth_tokens TO email_system;

ALTER TABLE invitations ENABLE ROW LEVEL SECURITY;
ALTER TABLE invitations FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON invitations TO email_app
    USING (organization_id = current_setting('app.org_id')::uuid)
    WITH CHECK (organization_id = current_setting('app.org_id')::uuid);
CREATE POLICY system_all ON invitations TO email_system
    USING (true) WITH CHECK (true);

ALTER TABLE auth_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE auth_tokens FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON auth_tokens TO email_app
    USING (organization_id = current_setting('app.org_id')::uuid)
    WITH CHECK (organization_id = current_setting('app.org_id')::uuid);
CREATE POLICY system_all ON auth_tokens TO email_system
    USING (true) WITH CHECK (true);

-- +goose Down
DROP TABLE IF EXISTS auth_tokens;
DROP TABLE IF EXISTS invitations;
