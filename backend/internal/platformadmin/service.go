package platformadmin

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"

	"github.com/TaushifReza/email-platform/backend/internal/audit"
	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/TaushifReza/email-platform/backend/internal/platform/clock"
	"github.com/TaushifReza/email-platform/backend/internal/tenancy"
)

const (
	defaultRetentionDays   = 365
	defaultStorageQuota    = int64(20 * 1024 * 1024 * 1024)
	defaultPageSize        = 50
	maxPageSize            = 100
	maxOrganizationNameLen = 200
	maxTimeZoneLen         = 64
)

// ErrNotFound is returned when an organization does not exist or is not in
// the state the guarded update expected.
var ErrNotFound = errors.New("platformadmin: organization not found")

// Service implements ADM-01 and the operator-level part of AUD-01/AUD-03.
type Service struct {
	repo  Repository
	clock clock.Clock
}

func NewService(repo Repository, clk clock.Clock) *Service {
	if clk == nil {
		clk = clock.System{}
	}
	return &Service{repo: repo, clock: clk}
}

func (s *Service) ListOrganizations(ctx context.Context, actor identity.Actor, page Page) ([]tenancy.Organization, error) { //nolint:lll
	if err := identity.Authorize(actor, identity.Request{Capability: identity.CapManageOrganizations}); err != nil {
		return nil, err
	}
	return s.repo.ListOrganizations(ctx, normalizePage(page))
}

func (s *Service) CreateOrganization(ctx context.Context, actor identity.Actor, in CreateOrganizationInput,
	info identity.ClientInfo,
) (tenancy.Organization, error) {
	if err := identity.Authorize(actor, identity.Request{Capability: identity.CapManageOrganizations}); err != nil {
		return tenancy.Organization{}, err
	}
	normalized, err := normalizeCreate(in)
	if err != nil {
		return tenancy.Organization{}, err
	}
	id, err := newUUID()
	if err != nil {
		return tenancy.Organization{}, err
	}
	entry := s.entry(actor, info, audit.ActionOrgCreated, "organization", id)
	return s.repo.CreateOrganization(ctx, id, normalized, entry)
}

func (s *Service) SuspendOrganization(ctx context.Context, actor identity.Actor, orgID string,
	info identity.ClientInfo,
) error {
	return s.changeStatus(ctx, actor, orgID, "active", "suspended", audit.ActionOrgSuspended, info)
}

func (s *Service) ReinstateOrganization(ctx context.Context, actor identity.Actor, orgID string,
	info identity.ClientInfo,
) error {
	return s.changeStatus(ctx, actor, orgID, "suspended", "active", audit.ActionOrgReinstated, info)
}

func (s *Service) changeStatus(ctx context.Context, actor identity.Actor, orgID, expected, next, action string,
	info identity.ClientInfo,
) error {
	if err := identity.Authorize(actor, identity.Request{Capability: identity.CapManageOrganizations}); err != nil {
		return err
	}
	if orgID == "" {
		return ErrNotFound
	}
	entry := s.entry(actor, info, action, "organization", orgID)
	changed, err := s.repo.SetOrganizationStatus(ctx, orgID, expected, next, entry)
	if err != nil {
		return err
	}
	if !changed {
		return ErrNotFound
	}
	return nil
}

func (s *Service) UpdateLimits(ctx context.Context, actor identity.Actor, orgID string, in LimitsInput,
	info identity.ClientInfo,
) (tenancy.Organization, error) {
	if err := identity.Authorize(actor, identity.Request{Capability: identity.CapOrgLimits}); err != nil {
		return tenancy.Organization{}, err
	}
	if orgID == "" || in.StorageQuotaBytes <= 0 {
		return tenancy.Organization{}, errors.New("platformadmin: invalid limits")
	}
	if !validCap(in.DailyEmailCap) || !validCap(in.CampaignRecipientCap) {
		return tenancy.Organization{}, errors.New("platformadmin: invalid caps")
	}
	entry := s.entry(actor, info, audit.ActionOrgLimitsChanged, "organization", orgID)
	org, err := s.repo.UpdateOrganizationLimits(ctx, orgID, in, entry)
	if err != nil {
		return tenancy.Organization{}, err
	}
	if org.ID == "" {
		return tenancy.Organization{}, ErrNotFound
	}
	return org, nil
}

func (s *Service) ListOperatorAudit(ctx context.Context, actor identity.Actor, page Page) ([]audit.Entry, error) {
	if err := identity.Authorize(actor, identity.Request{Capability: identity.CapOperatorAudit}); err != nil {
		return nil, err
	}
	return s.repo.ListOperatorAudit(ctx, normalizePage(page))
}

func (s *Service) entry(actor identity.Actor, info identity.ClientInfo, action, targetType, targetID string,
) audit.Entry {
	return audit.Entry{
		ActorType:  "operator",
		ActorID:    actor.UserID,
		ActorRole:  actor.Role,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Outcome:    "success",
		SourceIP:   info.IP,
		Client:     info.UserAgent,
		At:         s.clock.Now().UTC(),
	}
}

func normalizeCreate(in CreateOrganizationInput) (CreateOrganizationInput, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" || len(in.Name) > maxOrganizationNameLen {
		return in, errors.New("platformadmin: invalid organization name")
	}
	in.DefaultSenderName = strings.TrimSpace(in.DefaultSenderName)
	if in.DefaultSenderName == "" {
		in.DefaultSenderName = in.Name
	}
	in.DefaultReplyTo = strings.TrimSpace(in.DefaultReplyTo)
	in.TimeZone = strings.TrimSpace(in.TimeZone)
	if in.TimeZone == "" {
		in.TimeZone = "UTC"
	}
	if len(in.TimeZone) > maxTimeZoneLen {
		return in, errors.New("platformadmin: invalid time zone")
	}
	return in, nil
}

func normalizePage(page Page) Page {
	if page.Limit <= 0 {
		page.Limit = defaultPageSize
	}
	if page.Limit > maxPageSize {
		page.Limit = maxPageSize
	}
	if page.Offset < 0 {
		page.Offset = 0
	}
	return page
}

func validCap(value int) bool {
	return value >= 0
}

// NewID generates a time-ordered-format UUID for operator records.
func NewID() (string, error) {
	return newUUID()
}

func newUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("platformadmin: generate id: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
