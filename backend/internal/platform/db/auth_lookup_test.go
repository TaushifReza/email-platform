package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/platform/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func timeNowNano() int64 { return time.Now().UnixNano() }

func authLookupTestDSN() string {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}
	return "postgres://postgres:password@localhost:5432/email_platform_dev?sslmode=disable"
}

func TestAUTH_01_AuthLookupSecurityDefiner(t *testing.T) {
	raw := openTestDB(t)
	var prosecdef bool
	var owner string
	var config string
	err := raw.QueryRow(`SELECT p.prosecdef, pg_get_userbyid(p.proowner), COALESCE(array_to_string(p.proconfig, ','), '')
		FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace
		WHERE n.nspname='public' AND p.proname='auth_lookup_user'`).Scan(&prosecdef, &owner, &config)
	require.NoError(t, err)
	assert.True(t, prosecdef)
	assert.Equal(t, "email_system", owner)
	assert.Contains(t, config, "search_path=pg_catalog, public")

	var appExec bool
	q := `SELECT has_function_privilege(` +
		`'email_app', 'public.auth_lookup_user(text)', 'EXECUTE')`
	err = raw.QueryRow(q).Scan(&appExec)
	require.NoError(t, err)
	assert.True(t, appExec)

	var publicExec int
	q2 := `SELECT count(*) FROM aclexplode((SELECT proacl FROM pg_proc p ` +
		`JOIN pg_namespace n ON n.oid=p.pronamespace ` +
		`WHERE n.nspname='public' AND p.proname='auth_lookup_user')) ` +
		`WHERE grantee=0 AND privilege_type='EXECUTE'`
	err = raw.QueryRow(q2).Scan(&publicExec)
	require.NoError(t, err)
	assert.Equal(t, 0, publicExec)
}

func TestSEC_16_AuthLookupReturnsOnlyLoginFields(t *testing.T) {
	raw := openTestDB(t)
	orgID := seedOrganization(t, raw, "Org A")
	seedUser(t, raw, orgID, fmt.Sprintf("fields-%d@example.com", timeNowNano()))
	var payload []byte
	err := raw.QueryRow(`SELECT auth_lookup_user($1)`, "fields@example.com").Scan(&payload)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(payload, &m))
	keys := []string{
		"id", "organization_id", "password_hash", "status",
		"role", "failed_login_count", "locked_until", "totp_enrolled_at",
	}
	for _, key := range keys {
		assert.Contains(t, m, key)
	}
	for _, forbidden := range []string{"totp_seed_encrypted", "recovery_code_hashes", "email"} {
		assert.NotContains(t, m, forbidden)
	}
}

func TestAC_SEC_01_PreAuthCannotReadTenantTables(t *testing.T) {
	raw := openTestDB(t)
	orgID := seedOrganization(t, raw, "Org A")
	seedUser(t, raw, orgID, fmt.Sprintf("preauth-%d@example.com", timeNowNano()))
	handle, err := New(context.Background(), authLookupTestDSN())
	require.NoError(t, err)
	defer handle.Close()
	err = handle.WithPreAuth(context.Background(), func(tx *sql.Tx) error {
		var count int
		return tx.QueryRow(`SELECT count(*) FROM users`).Scan(&count)
	})
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "app.org_id")
}

func TestAUTH_01_AuthLookupWithoutOrg(t *testing.T) {
	raw := openTestDB(t)
	orgID := seedOrganization(t, raw, "Org A")
	seedUser(t, raw, orgID, fmt.Sprintf("preauth2-%d@example.com", timeNowNano()))
	handle, err := New(context.Background(), authLookupTestDSN())
	require.NoError(t, err)
	defer handle.Close()
	var payload []byte
	err = handle.WithPreAuth(context.Background(), func(tx *sql.Tx) error {
		p, err := sqlc.New(tx).LookupUserForLogin(context.Background(), "preauth2@example.com")
		if err != nil {
			return err
		}
		payload = p
		return nil
	})
	require.NoError(t, err)
	assert.NotEmpty(t, payload)
}
