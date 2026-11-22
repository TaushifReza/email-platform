package identity

import (
	"context"
	"errors"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/audit"
	"github.com/TaushifReza/email-platform/backend/internal/platform/clock"
	"github.com/TaushifReza/email-platform/backend/internal/platform/db"
)

type User struct {
	ID          string
	Email       string
	DisplayName string
	Role        string
	Status      string
	CreatedAt   time.Time
}

type UserRepository interface {
	List(ctx context.Context, orgID string, limit, offset int) ([]User, error)
	Get(ctx context.Context, orgID, userID string) (User, error)
	CreateInvited(ctx context.Context, orgID, userID, email, displayName, role string, now time.Time, entry audit.Entry) error //nolint:lll
	ChangeRole(ctx context.Context, orgID, userID, role string, now time.Time, entry audit.Entry) error
	Deactivate(ctx context.Context, orgID, userID string, now time.Time, entry audit.Entry) error
	LockActiveOwners(ctx context.Context, orgID string) (int, error)
}

type UserChange struct {
	Role       string // "" means no change
	Deactivate bool
}

type InvitationIssuer interface {
	IssueInvitation(ctx context.Context, actor Actor, email, role string) error
}

type UserService struct {
	repo     UserRepository
	sessions SessionRepository
	access   InvitationIssuer
	clock    clock.Clock
}

func NewUserService(repo UserRepository, sessions SessionRepository, access InvitationIssuer, clk clock.Clock) *UserService { //nolint:lll
	if clk == nil {
		clk = clock.System{}
	}
	return &UserService{repo: repo, sessions: sessions, access: access, clock: clk}
}

func (s *UserService) List(ctx context.Context, actor Actor) ([]User, error) {
	if err := Authorize(actor, Request{Capability: CapManageUsers}); err != nil {
		return nil, err
	}
	return s.repo.List(ctx, actor.OrganizationID, 100, 0)
}

func (s *UserService) Invite(ctx context.Context, actor Actor, email, role string, info ClientInfo) error {
	if err := Authorize(actor, Request{Capability: CapManageUsers, TargetRole: role}); err != nil {
		return err
	}
	email = normalizeEmail(email)
	if email == "" || !validRole(role) {
		return errors.New("identity: invalid invitation")
	}
	now := s.clock.Now()
	userID := newUUID()
	entry := audit.Entry{
		OrganizationID: actor.OrganizationID, ActorType: "user", ActorID: actor.UserID,
		ActorRole: actor.Role, Action: audit.ActionUserInvited, TargetType: "user",
		TargetID: userID, Outcome: "success", SourceIP: info.IP, Client: info.UserAgent, At: now,
	}
	if err := s.repo.CreateInvited(ctx, actor.OrganizationID, userID, email, "", role, now, entry); err != nil {
		return err
	}
	return s.access.IssueInvitation(ctx, actor, email, role)
}

func (s *UserService) Update(ctx context.Context, actor Actor, userID string, change UserChange, info ClientInfo) error { //nolint:gocognit,lll
	if err := Authorize(actor, Request{Capability: CapManageUsers, TargetRole: change.Role}); err != nil {
		return err
	}
	// admin cannot manage owners
	if actor.Role == roleAdmin {
		u, err := s.repo.Get(ctx, actor.OrganizationID, userID)
		if err != nil {
			return err
		}
		if u.Role == roleOwner {
			return ErrForbidden
		}
		if change.Role == roleOwner {
			return ErrForbidden
		}
	}
	now := s.clock.Now()
	// last-owner guard
	if change.Role != "" && change.Role != roleOwner {
		n, err := s.repo.LockActiveOwners(ctx, actor.OrganizationID)
		if err != nil {
			return err
		}
		if n <= 1 {
			u, _ := s.repo.Get(ctx, actor.OrganizationID, userID) //nolint:errcheck
			if u.Role == roleOwner {
				return errors.New("identity: last owner cannot be demoted")
			}
		}
	}
	if change.Deactivate {
		n, err := s.repo.LockActiveOwners(ctx, actor.OrganizationID)
		if err != nil {
			return err
		}
		u, _ := s.repo.Get(ctx, actor.OrganizationID, userID) //nolint:errcheck
		if u.Role == roleOwner && n <= 1 {
			return errors.New("identity: last owner cannot be deactivated")
		}
		// step-up required for deactivation
		if !stepUpValid(actor, now) {
			return ErrMFAStepUp
		}
		entry := audit.Entry{
			OrganizationID: actor.OrganizationID, ActorType: "user", ActorID: actor.UserID,
			ActorRole: actor.Role, Action: audit.ActionUserDeactivated, TargetType: "user",
			TargetID: userID, Outcome: "success", SourceIP: info.IP, Client: info.UserAgent, At: now,
		}
		if err := s.repo.Deactivate(ctx, actor.OrganizationID, userID, now, entry); err != nil {
			return err
		}
		// revoke sessions
		s.revokeSessions(ctx, actor.OrganizationID, userID, now)
		return nil
	}
	if change.Role != "" {
		if !stepUpValid(actor, now) {
			return ErrMFAStepUp
		}
		entry := audit.Entry{
			OrganizationID: actor.OrganizationID, ActorType: "user", ActorID: actor.UserID,
			ActorRole: actor.Role, Action: audit.ActionUserRoleChanged, TargetType: "user",
			TargetID: userID, Outcome: "success", SourceIP: info.IP, Client: info.UserAgent, At: now,
			Details: map[string]any{"role": change.Role},
		}
		if err := s.repo.ChangeRole(ctx, actor.OrganizationID, userID, change.Role, now, entry); err != nil {
			return err
		}
		s.revokeSessions(ctx, actor.OrganizationID, userID, now)
	}
	return nil
}

func stepUpValid(actor Actor, now time.Time) bool {
	if actor.StepUpExpiresAt == nil {
		return false
	}
	return actor.StepUpExpiresAt.After(now)
}

func (s *UserService) revokeSessions(ctx context.Context, orgID, userID string, now time.Time) {
	if s.sessions != nil {
		_ = s.sessions.RevokeAllUserSessions(ctx, orgID, userID, now) //nolint:errcheck
	}
}

var _ = db.DB{}
