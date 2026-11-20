package db

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAC_SEC_01_RLSBlocksCrossTenantRead(t *testing.T) {
	db := openTestDB(t)
	orgA := seedOrganization(t, db, "Org A")
	orgB := seedOrganization(t, db, "Org B")
	userA := seedUser(t, db, orgA, fmt.Sprintf("a-%s@example.com", orgA))
	userB := seedUser(t, db, orgB, fmt.Sprintf("b-%s@example.com", orgB))
	invA := seedInvitation(t, db, orgA, userA, fmt.Sprintf("invite-a-%s@example.com", orgA))
	seedInvitation(t, db, orgB, userB, fmt.Sprintf("invite-b-%s@example.com", orgB))
	sessA := seedSession(t, db, orgA, userA)
	seedSession(t, db, orgB, userB)
	auditA := seedAudit(t, db, orgA, "tenant-action")
	seedAudit(t, db, orgB, "other-action")

	tx := appTx(t, db, orgA)
	cleanupTx(t, tx)

	assert.Equal(t, []string{orgA}, queryIDs(t, tx, `SELECT id::text FROM organizations`))
	assert.Equal(t, []string{userA}, queryIDs(t, tx, `SELECT id::text FROM users`))
	assert.Equal(t, []string{invA}, queryIDs(t, tx, `SELECT id::text FROM invitations`))
	assert.Equal(t, []string{sessA}, queryIDs(t, tx, `SELECT id::text FROM sessions`))
	assert.Equal(t, []string{auditA}, queryIDs(t, tx, `SELECT id::text FROM audit_log`))
}

func TestAC_SEC_01_RLSBlocksCrossTenantWrite(t *testing.T) {
	db := openTestDB(t)
	orgA := seedOrganization(t, db, "Org A")
	orgB := seedOrganization(t, db, "Org B")
	userB := seedUser(t, db, orgB, fmt.Sprintf("b-%s@example.com", orgB))
	sessB := seedSession(t, db, orgB, userB)

	tx := appTx(t, db, orgA)
	cleanupTx(t, tx)

	res, err := tx.Exec(`UPDATE users SET display_name = 'blocked' WHERE id = $1`, userB)
	require.NoError(t, err)
	affected, err := res.RowsAffected()
	require.NoError(t, err)
	assert.Zero(t, affected)

	res, err = tx.Exec(`DELETE FROM sessions WHERE id = $1`, sessB)
	require.NoError(t, err)
	affected, err = res.RowsAffected()
	require.NoError(t, err)
	assert.Zero(t, affected)

	res, err = tx.Exec(`UPDATE organizations SET name = 'blocked' WHERE id = $1`, orgB)
	require.NoError(t, err)
	affected, err = res.RowsAffected()
	require.NoError(t, err)
	assert.Zero(t, affected)

	tx = appTx(t, db, orgA)
	_, err = tx.Exec(
		`INSERT INTO users (id, organization_id, email, role, status)
		 VALUES (gen_random_uuid(), $1, 'cross@example.com', 'viewer', 'active')`,
		orgB,
	)
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "row-level security")
	require.NoError(t, tx.Rollback())

	tx = appTx(t, db, orgA)
	_, err = tx.Exec(
		`INSERT INTO audit_log (id, organization_id, actor_type, action, outcome)
		 VALUES (gen_random_uuid(), $1, 'user', 'cross-tenant', 'success')`,
		orgB,
	)
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "row-level security")
	require.NoError(t, tx.Rollback())
}

func TestAC_SEC_01_OrganizationsRLS(t *testing.T) {
	db := openTestDB(t)
	orgA := seedOrganization(t, db, "Org A")
	orgB := seedOrganization(t, db, "Org B")

	tx := appTx(t, db, orgA)
	cleanupTx(t, tx)

	assert.Equal(t, []string{orgA}, queryIDs(t, tx, `SELECT id::text FROM organizations`))

	res, err := tx.Exec(`UPDATE organizations SET name = 'blocked' WHERE id = $1`, orgB)
	require.NoError(t, err)
	affected, err := res.RowsAffected()
	require.NoError(t, err)
	assert.Zero(t, affected)

	res, err = tx.Exec(`UPDATE organizations SET name = 'allowed' WHERE id = $1`, orgA)
	require.NoError(t, err)
	affected, err = res.RowsAffected()
	require.NoError(t, err)
	assert.Equal(t, int64(1), affected)
}

func TestAC_AUD_01_AuditLogAppendOnlyForTenant(t *testing.T) {
	db := openTestDB(t)
	orgA := seedOrganization(t, db, "Org A")
	userA := seedUser(t, db, orgA, fmt.Sprintf("a-%s@example.com", orgA))

	tx := appTx(t, db, orgA)
	var auditID string
	err := tx.QueryRow(
		`INSERT INTO audit_log (id, organization_id, actor_type, actor_id, action, outcome)
		 VALUES (gen_random_uuid(), $1, 'user', $2, 'login', 'success')
		 RETURNING id::text`,
		orgA, userA,
	).Scan(&auditID)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	tx = appTx(t, db, orgA)
	_, err = tx.Exec(`UPDATE audit_log SET action = 'tampered' WHERE id = $1`, auditID)
	require.Error(t, err, "tenant role must not update audit entries")
	require.NoError(t, tx.Rollback())

	tx = appTx(t, db, orgA)
	_, err = tx.Exec(`DELETE FROM audit_log WHERE id = $1`, auditID)
	require.Error(t, err, "tenant role must not delete audit entries")
	require.NoError(t, tx.Rollback())

	var count int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM audit_log WHERE id = $1`, auditID).Scan(&count))
	assert.Equal(t, 1, count)
}

func TestAC_AUD_01_AuditLogAcceptsOperatorAction(t *testing.T) {
	db := openTestDB(t)
	orgA := seedOrganization(t, db, "Org A")
	seedAudit(t, db, orgA, "tenant-action")

	tx := systemTx(t, db)
	var operatorAuditID string
	err := tx.QueryRow(
		`INSERT INTO audit_log (id, organization_id, actor_type, actor_id, action, outcome)
		 VALUES (gen_random_uuid(), NULL, 'operator', gen_random_uuid(), 'organization.suspend', 'success')
		 RETURNING id::text`,
	).Scan(&operatorAuditID)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	tx = appTx(t, db, orgA)
	cleanupTx(t, tx)

	var nullCount int
	require.NoError(t, tx.QueryRow(`SELECT count(*) FROM audit_log WHERE organization_id IS NULL`).Scan(&nullCount))
	assert.Zero(t, nullCount)

	var operatorActions int
	require.NoError(t, tx.QueryRow(
		`SELECT count(*) FROM audit_log WHERE action = 'organization.suspend'`,
	).Scan(&operatorActions))
	assert.Zero(t, operatorActions)

	var stored int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM audit_log WHERE id = $1`, operatorAuditID).Scan(&stored))
	assert.Equal(t, 1, stored)
}
