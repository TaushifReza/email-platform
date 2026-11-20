package db

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("pgx", testDSN())
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	require.NoError(t, db.Ping())
	require.NoError(t, Migrate(context.Background(), db))

	return db
}

func seedOrganization(t *testing.T, db *sql.DB, name string) string {
	t.Helper()

	var id string
	err := db.QueryRow(
		`INSERT INTO organizations (id, name) VALUES (gen_random_uuid(), $1) RETURNING id::text`,
		name,
	).Scan(&id)
	require.NoError(t, err)

	return id
}

func seedUser(t *testing.T, db *sql.DB, organizationID, email string) string {
	t.Helper()

	var id string
	err := db.QueryRow(
		`INSERT INTO users (id, organization_id, email, display_name, role, status, password_hash)
		 VALUES (gen_random_uuid(), $1, $2, 'Test User', 'owner', 'active', 'hash')
		 RETURNING id::text`,
		organizationID, email,
	).Scan(&id)
	require.NoError(t, err)

	return id
}

func seedInvitation(t *testing.T, db *sql.DB, organizationID, invitedByUserID, email string) string {
	t.Helper()

	var id string
	err := db.QueryRow(
		`INSERT INTO invitations (id, organization_id, email, role, invited_by_user_id, token_hash, expires_at)
		 VALUES (gen_random_uuid(), $1, $2, 'viewer', $3, decode(md5(random()::text), 'hex'), now() + interval '72 hours')
		 RETURNING id::text`,
		organizationID, email, invitedByUserID,
	).Scan(&id)
	require.NoError(t, err)

	return id
}

func seedSession(t *testing.T, db *sql.DB, organizationID, userID string) string {
	t.Helper()

	var id string
	err := db.QueryRow(
		`INSERT INTO sessions (id, organization_id, user_id, token_hash, idle_expires_at, absolute_expires_at)
		 VALUES (
		     gen_random_uuid(), $1, $2, decode(md5(random()::text), 'hex'),
		     now() + interval '2 hours', now() + interval '24 hours'
		 )
		 RETURNING id::text`,
		organizationID, userID,
	).Scan(&id)
	require.NoError(t, err)

	return id
}

func seedAudit(t *testing.T, db *sql.DB, organizationID any, action string) string {
	t.Helper()

	var id string
	err := db.QueryRow(
		`INSERT INTO audit_log (id, organization_id, actor_type, action, outcome)
		 VALUES (gen_random_uuid(), $1, 'user', $2, 'success')
		 RETURNING id::text`,
		organizationID, action,
	).Scan(&id)
	require.NoError(t, err)

	return id
}

func appTx(t *testing.T, db *sql.DB, organizationID string) *sql.Tx {
	t.Helper()

	tx, err := db.Begin()
	require.NoError(t, err)
	_, err = tx.Exec("SET LOCAL ROLE email_app")
	require.NoError(t, err)
	_, err = tx.Exec("SELECT set_config('app.org_id', $1, true)", organizationID)
	require.NoError(t, err)

	return tx
}

func systemTx(t *testing.T, db *sql.DB) *sql.Tx {
	t.Helper()

	tx, err := db.Begin()
	require.NoError(t, err)
	_, err = tx.Exec("SET LOCAL ROLE email_system")
	require.NoError(t, err)

	return tx
}

func cleanupTx(t *testing.T, tx *sql.Tx) {
	t.Helper()

	t.Cleanup(func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			t.Errorf("rollback: %v", err)
		}
	})
}

func queryIDs(t *testing.T, tx *sql.Tx, query string) []string {
	t.Helper()

	rows, err := tx.Query(query)
	require.NoError(t, err)
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		require.NoError(t, rows.Scan(&id))
		ids = append(ids, id)
	}
	require.NoError(t, rows.Err())

	return ids
}
