-- Development-only bootstrap roles. Passwords here are disposable and must
-- never be used outside the local compose stack.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'email_owner') THEN
        CREATE ROLE email_owner NOLOGIN;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'email_app') THEN
        CREATE ROLE email_app LOGIN PASSWORD 'app-dev' NOSUPERUSER NOCREATEDB NOCREATEROLE NOBYPASSRLS;
    ELSE
        ALTER ROLE email_app LOGIN PASSWORD 'app-dev' NOSUPERUSER NOCREATEDB NOCREATEROLE NOBYPASSRLS;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'email_system') THEN
        CREATE ROLE email_system LOGIN PASSWORD 'system-dev' NOSUPERUSER NOCREATEDB NOCREATEROLE NOBYPASSRLS;
    ELSE
        ALTER ROLE email_system LOGIN PASSWORD 'system-dev' NOSUPERUSER NOCREATEDB NOCREATEROLE NOBYPASSRLS;
    END IF;
END
$$;

GRANT USAGE, CREATE ON SCHEMA public TO email_owner;
GRANT USAGE ON SCHEMA public TO email_app, email_system;
