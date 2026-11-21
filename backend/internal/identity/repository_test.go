package identity

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/platform/db"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func identityTestDSN() string {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}
	return "postgres://postgres:password@localhost:5432/email_platform_dev?sslmode=disable"
}

func openIdentityDB(t *testing.T) (*sql.DB, *db.DB) {
	t.Helper()
	dsn := identityTestDSN()
	raw, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { raw.Close() })
	require.NoError(t, raw.Ping())
	require.NoError(t, db.Migrate(context.Background(), raw))
	database, err := db.New(context.Background(), dsn)
	require.NoError(t, err)
	t.Cleanup(func() { database.Close() })
	return raw, database
}

func TestAUTH_01_AuthLookupWithoutOrg(t *testing.T) {
	raw, database := openIdentityDB(t)
	orgID := seedIdentityOrg(t, raw)
	email := fmt.Sprintf("lookup-%d@example.com", time.Now().UnixNano())
	userID := seedIdentityUser(t, raw, orgID, email)
	repo := NewRepository(database)
	creds, err := repo.LookupForLogin(context.Background(), email)
	require.NoError(t, err)
	assert.Equal(t, userID, creds.ID)
	assert.Equal(t, orgID, creds.OrganizationID)
	assert.Equal(t, "hash", creds.PasswordHash)
	assert.Equal(t, "active", creds.Status)
}

func TestAUTH_04_RecordFailedLoginLocks(t *testing.T) {
	raw, database := openIdentityDB(t)
	orgID := seedIdentityOrg(t, raw)
	email := fmt.Sprintf("fail-%d@example.com", time.Now().UnixNano())
	userID := seedIdentityUser(t, raw, orgID, email)
	repo := NewRepository(database)
	now := time.Now().UTC()
	for range 5 {
		err := repo.RecordFailedLogin(context.Background(), orgID, userID, now, 5, now.Add(15*time.Minute))
		require.NoError(t, err)
	}
	var count int
	var locked sql.NullTime
	q := `SELECT failed_login_count, locked_until FROM users WHERE id=$1`
	require.NoError(t, raw.QueryRow(q, userID).Scan(&count, &locked))
	assert.Equal(t, 5, count)
	assert.True(t, locked.Valid)
}

func TestAUTH_04_ClearFailedLogins(t *testing.T) {
	raw, database := openIdentityDB(t)
	orgID := seedIdentityOrg(t, raw)
	email := fmt.Sprintf("clear-%d@example.com", time.Now().UnixNano())
	userID := seedIdentityUser(t, raw, orgID, email)
	_, err := raw.Exec(`UPDATE users SET failed_login_count=4, locked_until=now()+interval '1 minute' WHERE id=$1`, userID)
	require.NoError(t, err)
	repo := NewRepository(database)
	require.NoError(t, repo.ClearFailedLogins(context.Background(), orgID, userID))
	var count int
	var locked sql.NullTime
	q := `SELECT failed_login_count, locked_until FROM users WHERE id=$1`
	require.NoError(t, raw.QueryRow(q, userID).Scan(&count, &locked))
	assert.Equal(t, 0, count)
	assert.False(t, locked.Valid)
}

func TestBR_01_FailedLoginCrossTenantNoop(t *testing.T) {
	raw, database := openIdentityDB(t)
	orgA := seedIdentityOrg(t, raw)
	orgB := seedIdentityOrg(t, raw)
	emailB := fmt.Sprintf("cross-%d@example.com", time.Now().UnixNano())
	userB := seedIdentityUser(t, raw, orgB, emailB)
	repo := NewRepository(database)
	now := time.Now().UTC()
	err := repo.RecordFailedLogin(context.Background(), orgA, userB, now, 5, now.Add(15*time.Minute))
	require.NoError(t, err)
	var count int
	require.NoError(t, raw.QueryRow(`SELECT failed_login_count FROM users WHERE id=$1`, userB).Scan(&count))
	assert.Equal(t, 0, count)
}

func seedIdentityOrg(t *testing.T, db *sql.DB) string {
	t.Helper()
	var id string
	q := `INSERT INTO organizations (id, name) VALUES (gen_random_uuid(), $1) RETURNING id::text`
	err := db.QueryRow(q, "Org").Scan(&id)
	require.NoError(t, err)
	return id
}

func seedIdentityUser(t *testing.T, db *sql.DB, orgID, email string) string {
	t.Helper()
	var id string
	q := `INSERT INTO users (id, organization_id, email, display_name, role, status, password_hash) ` +
		`VALUES (gen_random_uuid(), $1, $2, 'Test', 'owner', 'active', 'hash') RETURNING id::text`
	err := db.QueryRow(q, orgID, email).Scan(&id)
	require.NoError(t, err)
	return id
}

var _ = fmt.Sprint
