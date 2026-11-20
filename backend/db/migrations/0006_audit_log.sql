-- +goose Up
CREATE TABLE audit_log (
    id uuid NOT NULL,
    organization_id uuid,
    actor_type text NOT NULL CHECK (actor_type IN ('user', 'operator', 'system')),
    actor_id uuid,
    actor_role text,
    action text NOT NULL,
    target_type text,
    target_id uuid,
    outcome text NOT NULL CHECK (outcome IN ('success', 'failure')),
    source_ip inet,
    client text,
    details jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (created_at, id)
) PARTITION BY RANGE (created_at);

-- +goose StatementBegin
DO $$
BEGIN
    EXECUTE 'CREATE TABLE audit_log_default PARTITION OF audit_log DEFAULT';
END
$$;
-- +goose StatementEnd

CREATE INDEX audit_log_organization_created_idx ON audit_log (organization_id, created_at DESC);
CREATE INDEX audit_log_created_idx ON audit_log (created_at DESC);
CREATE INDEX audit_log_actor_idx ON audit_log (actor_id);

ALTER TABLE audit_log OWNER TO email_owner;

REVOKE ALL ON audit_log FROM PUBLIC;
GRANT SELECT, INSERT ON audit_log TO email_app;
GRANT SELECT, INSERT, DELETE ON audit_log TO email_system;

ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_log FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_select ON audit_log FOR SELECT TO email_app
    USING (organization_id = current_setting('app.org_id')::uuid);
CREATE POLICY tenant_insert ON audit_log FOR INSERT TO email_app
    WITH CHECK (organization_id = current_setting('app.org_id')::uuid);
CREATE POLICY system_all ON audit_log TO email_system
    USING (true) WITH CHECK (true);

-- +goose Down
DROP TABLE IF EXISTS audit_log;
