package db

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchema_ForcedRLSOnIdentityTables(t *testing.T) {
	db := openTestDB(t)

	tables := []string{
		"organizations",
		"operators",
		"users",
		"sessions",
		"operator_sessions",
		"invitations",
		"auth_tokens",
		"audit_log",
	}

	for _, table := range tables {
		t.Run(table, func(t *testing.T) {
			var rls, forced bool
			err := db.QueryRow(
				`SELECT c.relrowsecurity, c.relforcerowsecurity
				 FROM pg_class c
				 JOIN pg_namespace n ON n.oid = c.relnamespace
				 WHERE n.nspname = 'public' AND c.relname = $1`,
				table,
			).Scan(&rls, &forced)
			require.NoError(t, err)
			assert.True(t, rls, "RLS must be enabled")
			assert.True(t, forced, "RLS must be forced")

			var policies int
			err = db.QueryRow(
				`SELECT count(*) FROM pg_policies WHERE schemaname = 'public' AND tablename = $1`,
				table,
			).Scan(&policies)
			require.NoError(t, err)
			assert.Positive(t, policies, "at least one policy is required")
		})
	}
}

func TestSchema_RolesHaveNoBypassRLS(t *testing.T) {
	db := openTestDB(t)

	for _, role := range []string{"email_owner", "email_app", "email_system"} {
		t.Run(role, func(t *testing.T) {
			var superuser, bypass bool
			err := db.QueryRow(
				`SELECT rolsuper, rolbypassrls FROM pg_roles WHERE rolname = $1`,
				role,
			).Scan(&superuser, &bypass)
			require.NoError(t, err)
			assert.False(t, superuser, "role must not be superuser")
			assert.False(t, bypass, "role must not bypass RLS")
		})
	}
}

func TestSchema_AppCannotAccessOperators(t *testing.T) {
	db := openTestDB(t)
	orgID := seedOrganization(t, db, "Org A")

	for _, query := range []string{
		`SELECT count(*) FROM operators`,
		`SELECT count(*) FROM operator_sessions`,
	} {
		tx := appTx(t, db, orgID)
		var count int
		err := tx.QueryRow(query).Scan(&count)
		require.Error(t, err, "app role must not reach operator tables")
		require.NoError(t, tx.Rollback())
	}
}

func TestSchema_CompositeFKRejectsCrossTenantReference(t *testing.T) {
	db := openTestDB(t)
	orgA := seedOrganization(t, db, "Org A")
	orgB := seedOrganization(t, db, "Org B")
	userA := seedUser(t, db, orgA, fmt.Sprintf("owner-a-%s@example.com", orgA))

	_, err := db.Exec(
		`INSERT INTO invitations (id, organization_id, email, role, invited_by_user_id, token_hash, expires_at)
		 VALUES (gen_random_uuid(), $1, $2, 'viewer', $3, decode(md5(random()::text), 'hex'), now() + interval '72 hours')`,
		orgB, fmt.Sprintf("cross-%s@example.com", orgB), userA,
	)
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "foreign key")

	_, err = db.Exec(
		`INSERT INTO sessions (id, organization_id, user_id, token_hash, idle_expires_at, absolute_expires_at)
		 VALUES (
		     gen_random_uuid(), $1, $2, decode(md5(random()::text), 'hex'),
		     now() + interval '2 hours', now() + interval '24 hours'
		 )`,
		orgB, userA,
	)
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "foreign key")
}

func TestORG_05_EmailUniqueAcrossOrganizations(t *testing.T) {
	db := openTestDB(t)
	orgA := seedOrganization(t, db, "Org A")
	orgB := seedOrganization(t, db, "Org B")
	sharedEmail := fmt.Sprintf("shared-%s@example.com", orgA)
	seedUser(t, db, orgA, sharedEmail)

	_, err := db.Exec(
		`INSERT INTO users (id, organization_id, email, role, status)
		 VALUES (gen_random_uuid(), $1, $2, 'viewer', 'active')`,
		orgB, strings.ToUpper(sharedEmail),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "users_email_normalized_key")
}

func TestORG_06_ApprovalOffByDefault(t *testing.T) {
	db := openTestDB(t)
	orgID := seedOrganization(t, db, "Org A")

	var requireApproval bool
	require.NoError(t, db.QueryRow(
		`SELECT require_approval FROM organizations WHERE id = $1`, orgID,
	).Scan(&requireApproval))
	assert.False(t, requireApproval)
}

func TestORG_09_StorageQuotaDefault(t *testing.T) {
	db := openTestDB(t)
	orgID := seedOrganization(t, db, "Org A")

	var quota int64
	require.NoError(t, db.QueryRow(
		`SELECT storage_quota_bytes FROM organizations WHERE id = $1`, orgID,
	).Scan(&quota))
	assert.Equal(t, int64(20*1024*1024*1024), quota)
}

func TestORG_10_CapsOffByDefault(t *testing.T) {
	db := openTestDB(t)
	orgID := seedOrganization(t, db, "Org A")

	var capsEnabled bool
	require.NoError(t, db.QueryRow(
		`SELECT caps_enabled FROM organizations WHERE id = $1`, orgID,
	).Scan(&capsEnabled))
	assert.False(t, capsEnabled)
}
