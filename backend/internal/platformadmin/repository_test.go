package platformadmin

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/audit"
	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/TaushifReza/email-platform/backend/internal/platform/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func openOperatorDB(t *testing.T) *db.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5432/email_platform_dev?sslmode=disable"
	}
	raw, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = raw.Close() })
	require.NoError(t, raw.Ping())
	require.NoError(t, db.Migrate(context.Background(), raw))
	database, err := db.New(context.Background(), dsn)
	require.NoError(t, err)
	t.Cleanup(func() { database.Close() })
	return database
}

func TestOperatorRepositoryRoundTrip(t *testing.T) {
	database := openOperatorDB(t)
	repo := NewRepository(database)
	ctx := context.Background()
	email := fmt.Sprintf("operator-%d@example.com", time.Now().UnixNano())

	created, err := repo.CreateOperator(ctx, NewTestID(t), email)
	require.NoError(t, err)
	require.True(t, created)

	op, err := repo.LookupOperatorByEmail(ctx, email)
	require.NoError(t, err)
	assert.Equal(t, email, op.Email)
	assert.NotEmpty(t, op.ID)

	svc := NewService(repo, nil)
	actor := identity.PlatformAdminActor(op.ID)
	org, err := svc.CreateOrganization(ctx, actor, CreateOrganizationInput{Name: "Operator Org", TimeZone: "UTC"},
		identity.ClientInfo{IP: "203.0.113.4"})
	require.NoError(t, err)
	assert.Equal(t, "active", org.Status)
	assert.Equal(t, defaultStorageQuota, org.StorageQuotaBytes)
	assert.False(t, org.CapsEnabled)

	require.NoError(t, svc.SuspendOrganization(ctx, actor, org.ID, identity.ClientInfo{}))
	orgs, err := svc.ListOrganizations(ctx, actor, Page{})
	require.NoError(t, err)
	found := false
	for _, o := range orgs {
		if o.ID == org.ID {
			found = true
			assert.Equal(t, "suspended", o.Status)
		}
	}
	assert.True(t, found)

	entries, err := svc.ListOperatorAudit(ctx, actor, Page{})
	require.NoError(t, err)
	actions := make([]string, 0, len(entries))
	for _, e := range entries {
		actions = append(actions, e.Action)
		assert.Empty(t, e.OrganizationID)
		assert.Equal(t, "operator", e.ActorType)
	}
	assert.Contains(t, actions, audit.ActionOrgCreated)
	assert.Contains(t, actions, audit.ActionOrgSuspended)
}

func NewTestID(t *testing.T) string {
	t.Helper()
	id, err := NewID()
	require.NoError(t, err)
	return id
}
