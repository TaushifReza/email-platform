package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSEC_03_WithOrgRejectsEmptyOrganization(t *testing.T) {
	raw := openTestDB(t)
	d := &DB{sql: raw}

	called := false
	err := d.WithOrg(context.Background(), "", func(*sql.Tx) error {
		called = true
		return nil
	})

	require.ErrorIs(t, err, ErrNoOrganization)
	assert.False(t, called, "fn must not run without an organization")
}

func TestSEC_03_WithOrgSetsScopeAndCommits(t *testing.T) {
	ctx := context.Background()
	raw := openTestDB(t)
	d := &DB{sql: raw}
	orgID := seedOrganization(t, raw, "WithOrg commit")
	email := fmt.Sprintf("withorg-%s@example.com", orgID)

	err := d.WithOrg(ctx, orgID, func(tx *sql.Tx) error {
		var scope, role string
		err := tx.QueryRowContext(ctx,
			"SELECT current_setting('app.org_id'), current_user").Scan(&scope, &role)
		if err != nil {
			return err
		}
		assert.Equal(t, orgID, scope)
		assert.Equal(t, "email_app", role)

		_, err = tx.ExecContext(ctx,
			`INSERT INTO users (id, organization_id, email, display_name, role, status, password_hash)
			 VALUES (gen_random_uuid(), $1, $2, 'WithOrg User', 'owner', 'active', 'hash')`,
			orgID, email)
		return err
	})
	require.NoError(t, err)

	var count int
	require.NoError(t, raw.QueryRow(
		`SELECT count(*) FROM users WHERE organization_id = $1 AND email = $2`, orgID, email,
	).Scan(&count))
	assert.Equal(t, 1, count)
}

func TestSEC_03_WithOrgRollsBackOnError(t *testing.T) {
	ctx := context.Background()
	raw := openTestDB(t)
	d := &DB{sql: raw}
	orgID := seedOrganization(t, raw, "WithOrg rollback")
	email := fmt.Sprintf("rollback-%s@example.com", orgID)
	boom := errors.New("boom")

	err := d.WithOrg(ctx, orgID, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO users (id, organization_id, email, display_name, role, status, password_hash)
			 VALUES (gen_random_uuid(), $1, $2, 'Rollback User', 'owner', 'active', 'hash')`,
			orgID, email)
		if err != nil {
			return err
		}
		return boom
	})
	require.ErrorIs(t, err, boom)

	var count int
	require.NoError(t, raw.QueryRow(`SELECT count(*) FROM users WHERE email = $1`, email).Scan(&count))
	assert.Zero(t, count)
}

func TestAC_SEC_01_WithOrgBlocksCrossTenantReadWrite(t *testing.T) {
	ctx := context.Background()
	raw := openTestDB(t)
	d := &DB{sql: raw}
	orgA := seedOrganization(t, raw, "WithOrg A")
	orgB := seedOrganization(t, raw, "WithOrg B")
	userA := seedUser(t, raw, orgA, fmt.Sprintf("a-%s@example.com", orgA))
	userB := seedUser(t, raw, orgB, fmt.Sprintf("b-%s@example.com", orgB))

	err := d.WithOrg(ctx, orgA, func(tx *sql.Tx) error {
		assert.Equal(t, []string{userA}, queryIDs(t, tx, `SELECT id::text FROM users`))

		res, err := tx.ExecContext(ctx, `UPDATE users SET display_name = 'blocked' WHERE id = $1`, userB)
		if err != nil {
			return err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		assert.Zero(t, affected)
		return nil
	})
	require.NoError(t, err)

	err = d.WithOrg(ctx, orgA, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO users (id, organization_id, email, role, status)
			 VALUES (gen_random_uuid(), $1, 'cross@example.com', 'viewer', 'active')`,
			orgB)
		return err
	})
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "row-level security")
}
