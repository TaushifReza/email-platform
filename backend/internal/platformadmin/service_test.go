package platformadmin

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/audit"
	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/TaushifReza/email-platform/backend/internal/tenancy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRepository struct {
	operator     Operator
	lookedUp     string
	createdOrgID string
	orgEntry     audit.Entry
	statusCalls  []string
	statusEntry  audit.Entry
	limitEntry   audit.Entry
	auditPage    Page
}

func (f *fakeRepository) LookupOperatorByEmail(_ context.Context, email string) (Operator, error) {
	f.lookedUp = email
	return f.operator, nil
}

func (f *fakeRepository) LookupOperatorCredentials(_ context.Context, email string) (OperatorCredentials, error) {
	f.lookedUp = email
	return OperatorCredentials{Operator: f.operator}, nil
}

func (f *fakeRepository) CreateOperator(context.Context, string, string) (bool, error) {
	return true, nil
}

func (f *fakeRepository) ListOrganizations(context.Context, Page) ([]tenancy.Organization, error) {
	return nil, nil
}

func (f *fakeRepository) CreateOrganization(_ context.Context, id string, in CreateOrganizationInput,
	entry audit.Entry,
) (tenancy.Organization, error) {
	f.createdOrgID = id
	f.orgEntry = entry
	return tenancy.Organization{ID: id, Name: in.Name, TimeZone: in.TimeZone}, nil
}

func (f *fakeRepository) SetOrganizationStatus(_ context.Context, _, expected, next string,
	entry audit.Entry,
) (bool, error) {
	f.statusCalls = append(f.statusCalls, expected+"->"+next)
	f.statusEntry = entry
	return true, nil
}

func (f *fakeRepository) UpdateOrganizationLimits(_ context.Context, orgID string, _ LimitsInput,
	entry audit.Entry,
) (tenancy.Organization, error) {
	f.limitEntry = entry
	return tenancy.Organization{ID: orgID}, nil
}

func (f *fakeRepository) ListOperatorAudit(_ context.Context, page Page) ([]audit.Entry, error) {
	f.auditPage = page
	return nil, nil
}

type fixedClock struct{ at time.Time }

func (c fixedClock) Now() time.Time { return c.at }

func operatorActor() identity.Actor { return identity.PlatformAdminActor("op-1") }

func fixedAt() time.Time { return time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC) }

func newTestService(repo *fakeRepository) *Service {
	return NewService(repo, fixedClock{at: fixedAt()})
}

func TestOnlyPlatformAdminMayManageOrganizations(t *testing.T) {
	repo := &fakeRepository{}
	svc := newTestService(repo)
	tenant := identity.Actor{UserID: "u1", OrganizationID: "o1", Role: "owner"}
	info := identity.ClientInfo{IP: "203.0.113.4", UserAgent: "test"}

	_, err := svc.CreateOrganization(context.Background(), tenant, CreateOrganizationInput{Name: "Acme"}, info)
	require.ErrorIs(t, err, identity.ErrForbidden)
	_, err = svc.UpdateLimits(context.Background(), tenant, "o1", LimitsInput{StorageQuotaBytes: 100}, info)
	require.ErrorIs(t, err, identity.ErrForbidden)
	require.ErrorIs(t, svc.SuspendOrganization(context.Background(), tenant, "o1", info), identity.ErrForbidden)
	_, err = svc.ListOperatorAudit(context.Background(), tenant, Page{})
	require.ErrorIs(t, err, identity.ErrForbidden)
	assert.Empty(t, repo.statusCalls)
}

func TestCreateOrganizationDefaultsAndOperatorAudit(t *testing.T) {
	repo := &fakeRepository{}
	svc := newTestService(repo)
	info := identity.ClientInfo{IP: "203.0.113.4", UserAgent: "console"}

	org, err := svc.CreateOrganization(context.Background(), operatorActor(), CreateOrganizationInput{
		Name: "  Acme  ", TimeZone: " ",
	}, info)
	require.NoError(t, err)
	assert.Equal(t, "Acme", org.Name)
	assert.NotEmpty(t, repo.createdOrgID)
	assert.Equal(t, "UTC", org.TimeZone)

	e := repo.orgEntry
	assert.Equal(t, "", e.OrganizationID, "operator actions are operator-level, not tenant-level")
	assert.Equal(t, "operator", e.ActorType)
	assert.Equal(t, "op-1", e.ActorID)
	assert.Equal(t, "platform_admin", e.ActorRole)
	assert.Equal(t, audit.ActionOrgCreated, e.Action)
	assert.Equal(t, "organization", e.TargetType)
	assert.Equal(t, repo.createdOrgID, e.TargetID)
	assert.Equal(t, "203.0.113.4", e.SourceIP)
	assert.Equal(t, fixedAt(), e.At)
}

func TestSuspendAndReinstateAreGuardedAndAudited(t *testing.T) {
	repo := &fakeRepository{}
	svc := newTestService(repo)
	info := identity.ClientInfo{}

	require.NoError(t, svc.SuspendOrganization(context.Background(), operatorActor(), "o1", info))
	require.NoError(t, svc.ReinstateOrganization(context.Background(), operatorActor(), "o1", info))
	assert.Equal(t, []string{"active->suspended", "suspended->active"}, repo.statusCalls)
	assert.Equal(t, audit.ActionOrgReinstated, repo.statusEntry.Action)
	require.ErrorIs(t, svc.SuspendOrganization(context.Background(), operatorActor(), "", info), ErrNotFound)
}

func TestLimitValidation(t *testing.T) {
	repo := &fakeRepository{}
	svc := newTestService(repo)
	valid := LimitsInput{StorageQuotaBytes: 100, CapsEnabled: true, DailyEmailCap: 500}

	_, err := svc.UpdateLimits(context.Background(), operatorActor(), "o1", valid, identity.ClientInfo{})
	require.NoError(t, err)
	assert.Equal(t, audit.ActionOrgLimitsChanged, repo.limitEntry.Action)

	_, err = svc.UpdateLimits(context.Background(), operatorActor(), "o1",
		LimitsInput{StorageQuotaBytes: 0}, identity.ClientInfo{})
	require.Error(t, err)
	_, err = svc.UpdateLimits(context.Background(), operatorActor(), "o1",
		LimitsInput{StorageQuotaBytes: 100, DailyEmailCap: -1}, identity.ClientInfo{})
	require.Error(t, err)
}

func TestAuditPageIsBounded(t *testing.T) {
	repo := &fakeRepository{}
	svc := newTestService(repo)
	_, err := svc.ListOperatorAudit(context.Background(), operatorActor(), Page{Limit: 5000, Offset: -2})
	require.NoError(t, err)
	assert.Equal(t, maxPageSize, repo.auditPage.Limit)
	assert.Equal(t, 0, repo.auditPage.Offset)
}

func TestAuthenticateOperator(t *testing.T) {
	repo := &fakeRepository{operator: Operator{ID: "op-1", Email: "dev@example.com"}}
	svc := NewIdentityService(repo)

	op, err := svc.Authenticate(context.Background(), "  DEV@Example.com ")
	require.NoError(t, err)
	assert.Equal(t, "op-1", op.ID)
	assert.Equal(t, "dev@example.com", repo.lookedUp)

	for _, invalid := range []string{"", "not-an-email", "a@b@c.example.com", "a b@example.com"} {
		_, err = svc.Authenticate(context.Background(), invalid)
		require.ErrorIs(t, err, ErrOperatorNotFound, invalid)
	}

	repo.operator = Operator{}
	_, err = svc.Authenticate(context.Background(), "unknown@example.com")
	require.ErrorIs(t, err, ErrOperatorNotFound)
}

func TestRepositoryFailureIsNotSwallowed(t *testing.T) {
	svc := NewIdentityService(&failingLookupRepository{})
	_, err := svc.Authenticate(context.Background(), "dev@example.com")
	require.ErrorIs(t, err, errLookupDown)
}

var errLookupDown = errors.New("lookup down")

type failingLookupRepository struct{ fakeRepository }

func (f *failingLookupRepository) LookupOperatorByEmail(context.Context, string) (Operator, error) {
	return Operator{}, errLookupDown
}

func (f *failingLookupRepository) LookupOperatorCredentials(context.Context, string) (OperatorCredentials, error) {
	return OperatorCredentials{}, errLookupDown
}
