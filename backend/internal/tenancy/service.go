package tenancy

import (
	"context"
	"errors"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/audit"
	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/TaushifReza/email-platform/backend/internal/platform/clock"
	"github.com/TaushifReza/email-platform/backend/internal/platform/db"
)

var ErrForbidden = errors.New("tenancy: forbidden")

type OrganizationRepository interface {
	Get(ctx context.Context, orgID string) (Organization, error)
	UpdateProfile(ctx context.Context, orgID string, p ProfileUpdate, now time.Time, entry audit.Entry) (Organization, error) //nolint:lll
	UpdatePolicy(ctx context.Context, orgID string, p PolicyUpdate, now time.Time, entry audit.Entry) (Organization, error)
}

type Service struct {
	repo  OrganizationRepository
	clock clock.Clock
}

func NewService(repo OrganizationRepository, clk clock.Clock) *Service {
	if clk == nil {
		clk = clock.System{}
	}
	return &Service{repo: repo, clock: clk}
}

func (s *Service) GetSettings(ctx context.Context, actor identity.Actor) (Organization, error) {
	if actor.OrganizationID == "" {
		return Organization{}, identity.ErrUnauthenticated
	}
	return s.repo.Get(ctx, actor.OrganizationID)
}

func (s *Service) UpdateProfile(ctx context.Context, actor identity.Actor, p ProfileUpdate, info identity.ClientInfo) (Organization, error) { //nolint:lll
	if err := identity.Authorize(actor, identity.Request{Capability: identity.CapOrgProfileSettings}); err != nil {
		return Organization{}, err
	}
	now := s.clock.Now()
	entry := audit.Entry{
		OrganizationID: actor.OrganizationID,
		ActorType:      "user",
		ActorID:        actor.UserID,
		ActorRole:      actor.Role,
		Action:         audit.ActionOrgProfileUpdated,
		TargetType:     "organization",
		TargetID:       actor.OrganizationID,
		Outcome:        "success",
		SourceIP:       info.IP,
		Client:         info.UserAgent,
		At:             now,
	}
	org, err := s.repo.UpdateProfile(ctx, actor.OrganizationID, p, now, entry)
	return org, err
}

func (s *Service) UpdatePolicy(ctx context.Context, actor identity.Actor, p PolicyUpdate, info identity.ClientInfo) (Organization, error) { //nolint:lll
	if err := identity.Authorize(actor, identity.Request{Capability: identity.CapOrgPolicySettings}); err != nil {
		return Organization{}, err
	}
	if p.RetentionDays <= 0 {
		return Organization{}, errors.New("tenancy: invalid retention")
	}
	now := s.clock.Now()
	entry := audit.Entry{
		OrganizationID: actor.OrganizationID,
		ActorType:      "user",
		ActorID:        actor.UserID,
		ActorRole:      actor.Role,
		Action:         audit.ActionOrgPolicyUpdated,
		TargetType:     "organization",
		TargetID:       actor.OrganizationID,
		Outcome:        "success",
		SourceIP:       info.IP,
		Client:         info.UserAgent,
		At:             now,
	}
	org, err := s.repo.UpdatePolicy(ctx, actor.OrganizationID, p, now, entry)
	return org, err
}

var _ = db.DB{}
