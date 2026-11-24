package providers

import (
	"context"
	"errors"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/audit"
	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/TaushifReza/email-platform/backend/internal/platform/crypto"
)

var (
	ErrNotFound    = errors.New("providers: not found")
	ErrForbidden   = errors.New("providers: forbidden")
	ErrStepUp      = identity.ErrMFAStepUp
	ErrDeactivated = errors.New("providers: deactivated")
	ErrDefault     = errors.New("providers: is default")
)

func (s *Service) List(ctx context.Context, actor identity.Actor) ([]Provider, error) {
	if err := identity.Authorize(actor, identity.Request{Capability: identity.CapManageProviders}); err != nil {
		return nil, err
	}
	return s.repo.List(ctx, actor.OrganizationID)
}

func (s *Service) Create(ctx context.Context, actor identity.Actor, in Input, info identity.ClientInfo) (Provider, error) { //nolint:lll
	if err := identity.Authorize(actor, identity.Request{Capability: identity.CapManageProviders}); err != nil {
		return Provider{}, err
	}
	if err := s.targets.ValidateTarget(ctx, in.Host, in.Port); err != nil {
		return Provider{}, errors.New("providers: host not allowed")
	}
	now := s.clock.Now()
	id, err := newID()
	if err != nil {
		return Provider{}, err
	}
	enc, keyID, err := s.keyring.Encrypt(crypto.PurposeProviderPassword, crypto.AAD{
		Table: "email_providers", RowID: id, Column: "password", OrganizationID: actor.OrganizationID,
	}, []byte(in.Password))
	if err != nil {
		return Provider{}, err
	}
	p := Provider{
		ID: id, OrganizationID: actor.OrganizationID, Host: in.Host, Port: in.Port,
		TLSMode: in.TLSMode, Username: in.Username, FromAddress: in.FromAddress,
		IsDefault: in.IsDefault, Status: "needs_attention", DailyLimit: 500,
		MinuteLimit: 20, HourLimit: 1200, ConfigVersion: 1,
	}
	if in.DailyLimit > 0 {
		p.DailyLimit = in.DailyLimit
	}
	if in.MinuteLimit > 0 {
		p.MinuteLimit = in.MinuteLimit
	}
	if in.HourLimit > 0 {
		p.HourLimit = in.HourLimit
	}
	entry := audit.Entry{
		OrganizationID: actor.OrganizationID, ActorType: "user", ActorID: actor.UserID,
		ActorRole: actor.Role, Action: audit.ActionProviderCreated, TargetType: "provider",
		TargetID: id, Outcome: "success", SourceIP: info.IP, Client: info.UserAgent, At: now,
	}
	if err := s.repo.Create(ctx, actor.OrganizationID, p, enc, keyID, now, entry); err != nil {
		return Provider{}, err
	}
	// reload
	return s.repo.Get(ctx, actor.OrganizationID, id)
}

func (s *Service) Update(ctx context.Context, actor identity.Actor, id string, u Update, info identity.ClientInfo) (Provider, error) { //nolint:lll
	if err := identity.Authorize(actor, identity.Request{Capability: identity.CapManageProviders}); err != nil {
		return Provider{}, err
	}
	if u.Host != nil || u.Port != nil {
		cur, err := s.repo.Get(ctx, actor.OrganizationID, id)
		if err != nil {
			return Provider{}, err
		}
		host := cur.Host
		port := cur.Port
		if u.Host != nil {
			host = *u.Host
		}
		if u.Port != nil {
			port = *u.Port
		}
		if err := s.targets.ValidateTarget(ctx, host, port); err != nil {
			return Provider{}, errors.New("providers: host not allowed")
		}
	}
	now := s.clock.Now()
	entry := audit.Entry{
		OrganizationID: actor.OrganizationID, ActorType: "user", ActorID: actor.UserID,
		ActorRole: actor.Role, Action: audit.ActionProviderUpdated, TargetType: "provider",
		TargetID: id, Outcome: "success", SourceIP: info.IP, Client: info.UserAgent, At: now,
	}
	p, err := s.repo.Update(ctx, actor.OrganizationID, id, u, now, entry)
	if err != nil {
		return Provider{}, err
	}
	if u.IsDefault != nil && *u.IsDefault {
		entry2 := audit.Entry{
			OrganizationID: actor.OrganizationID, ActorType: "user", ActorID: actor.UserID,
			ActorRole: actor.Role, Action: audit.ActionProviderUpdated, TargetType: "provider",
			TargetID: id, Outcome: "success", SourceIP: info.IP, Client: info.UserAgent, At: now,
		}
		p2, err := s.repo.SetDefault(ctx, actor.OrganizationID, id, now, entry2)
		if err != nil {
			return Provider{}, err
		}
		p = p2
	}
	return p, nil
}

func (s *Service) Rotate(ctx context.Context, actor identity.Actor, id string, username, password string, info identity.ClientInfo) (Provider, error) { //nolint:lll
	if err := identity.Authorize(actor, identity.Request{Capability: identity.CapManageProviders}); err != nil {
		return Provider{}, err
	}
	if !stepUpValid(actor, s.clock.Now()) {
		return Provider{}, ErrStepUp
	}
	now := s.clock.Now()
	enc, keyID, err := s.keyring.Encrypt(crypto.PurposeProviderPassword, crypto.AAD{
		Table: "email_providers", RowID: id, Column: "password", OrganizationID: actor.OrganizationID,
	}, []byte(password))
	if err != nil {
		return Provider{}, err
	}
	entry := audit.Entry{
		OrganizationID: actor.OrganizationID, ActorType: "user", ActorID: actor.UserID,
		ActorRole: actor.Role, Action: audit.ActionProviderRotated, TargetType: "provider",
		TargetID: id, Outcome: "success", SourceIP: info.IP, Client: info.UserAgent, At: now,
	}
	p, err := s.repo.Rotate(ctx, actor.OrganizationID, id, username, enc, keyID, now, entry)
	if err != nil {
		return Provider{}, err
	}
	return p, nil
}

func (s *Service) Deactivate(ctx context.Context, actor identity.Actor, id string, info identity.ClientInfo) (Provider, error) { //nolint:lll
	if err := identity.Authorize(actor, identity.Request{Capability: identity.CapManageProviders}); err != nil {
		return Provider{}, err
	}
	cur, err := s.repo.Get(ctx, actor.OrganizationID, id)
	if err != nil {
		return Provider{}, err
	}
	if cur.IsDefault {
		return Provider{}, ErrDefault
	}
	now := s.clock.Now()
	entry := audit.Entry{
		OrganizationID: actor.OrganizationID, ActorType: "user", ActorID: actor.UserID,
		ActorRole: actor.Role, Action: audit.ActionProviderDeactivated, TargetType: "provider",
		TargetID: id, Outcome: "success", SourceIP: info.IP, Client: info.UserAgent, At: now,
	}
	return s.repo.Deactivate(ctx, actor.OrganizationID, id, now, entry)
}

func (s *Service) Credentials(ctx context.Context, actor identity.Actor, id string) (Credentials, error) {
	if err := identity.Authorize(actor, identity.Request{Capability: identity.CapManageProviders}); err != nil {
		return Credentials{}, err
	}
	username, ct, keyID, err := s.repo.EncryptedCredentials(ctx, actor.OrganizationID, id)
	if err != nil {
		return Credentials{}, err
	}
	pt, err := s.keyring.Decrypt(crypto.PurposeProviderPassword, crypto.AAD{
		Table: "email_providers", RowID: id, Column: "password", OrganizationID: actor.OrganizationID,
	}, ct, keyID)
	if err != nil {
		return Credentials{}, err
	}
	return Credentials{Username: username, Password: string(pt)}, nil
}

func stepUpValid(actor identity.Actor, now time.Time) bool {
	if actor.StepUpExpiresAt == nil {
		return false
	}
	return actor.StepUpExpiresAt.After(now)
}
