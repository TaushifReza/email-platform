-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'email_owner') THEN
        CREATE ROLE email_owner NOLOGIN;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'email_app') THEN
        CREATE ROLE email_app NOLOGIN;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'email_system') THEN
        CREATE ROLE email_system NOLOGIN;
    END IF;
END
$$;
-- +goose StatementEnd

GRANT USAGE, CREATE ON SCHEMA public TO email_owner;
GRANT USAGE ON SCHEMA public TO email_app, email_system;

-- +goose Down
-- Roles are cluster-scoped and may be shared; they are intentionally not dropped.
