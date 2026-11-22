package identity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/platform/clock"
	"github.com/TaushifReza/email-platform/backend/internal/platform/outbox"
	"github.com/TaushifReza/email-platform/backend/internal/platform/ratelimit"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
}

type AccessService struct {
	repo    AccessRepository
	hasher  PasswordHasher
	clock   clock.Clock
	baseURL string
	limiter RateLimiter
}

func NewAccessService(repo AccessRepository, hasher PasswordHasher, clk clock.Clock,
	baseURL string, limiter RateLimiter,
) *AccessService {
	if clk == nil {
		clk = clock.System{}
	}
	return &AccessService{repo: repo, hasher: hasher, clock: clk, baseURL: baseURL, limiter: limiter}
}

func (s *AccessService) IssueInvitation(ctx context.Context, actor Actor, email, role string) error {
	if err := Authorize(actor, Request{Capability: CapManageUsers, TargetRole: role}); err != nil {
		return err
	}
	email = normalizeEmail(email)
	if email == "" || !validRole(role) {
		return errors.New("identity: invalid invitation")
	}
	now := s.clock.Now()
	invitationID := newUUID()
	token := newToken()
	tokenHash := sha256Bytes(token)
	expiresAt := now.Add(invitationExpiry)
	msg := outbox.Message{
		ID:             newUUID(),
		OrganizationID: actor.OrganizationID,
		UserID:         "", // user created at accept time; WP02.8 may pre-create; keep none
		Kind:           "invitation",
		Recipient:      email,
		Subject:        "You have been invited",
		TextBody:       "You have been invited. Accept: " + invitationURL(s.baseURL, token),
		NextAttemptAt:  now,
	}
	// Store token in invitation via repository? tokenHash stored. Good.
	err := s.repo.IssueInvitation(ctx, actor.OrganizationID, actor.UserID, email, role,
		now, expiresAt, tokenHash, invitationID, msg)
	if err != nil {
		return err
	}
	return nil
}

func (s *AccessService) AcceptInvitation(ctx context.Context, token, password string) error {
	if err := ValidatePassword(password); err != nil {
		return err
	}
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return err
	}
	tokenHash := sha256Bytes(token)
	inv, err := s.repo.LookupInvitation(ctx, tokenHash)
	if err != nil {
		return err
	}
	now := s.clock.Now()
	if inv.AcceptedAt != nil || inv.RevokedAt != nil {
		return errors.New("identity: invitation no longer valid")
	}
	if now.After(inv.ExpiresAt) {
		return errors.New("identity: invitation expired")
	}
	// user id must exist; if empty, create user? WP02.8 pre-creates users. For now require existing.
	if inv.UserID == "" {
		return errors.New("identity: invitation has no user")
	}
	return s.repo.AcceptInvitation(ctx, inv.OrganizationID, inv.UserID, inv.InvitationID, hash, now)
}

func (s *AccessService) RequestPasswordReset(ctx context.Context, email string, info ClientInfo) error {
	email = normalizeEmail(email)
	if email == "" {
		return nil
	}
	// rate limit per account and IP
	if s.limiter != nil {
		keys := []ratelimit.Key{{Name: "reset:account:" + email, Limit: 3, Window: time.Hour}}
		if info.IP != "" {
			keys = append(keys, ratelimit.Key{Name: "reset:ip:" + info.IP, Limit: 10, Window: time.Hour})
		}
		decision, err := s.limiter.Allow(ctx, keys)
		if err != nil {
			return err
		}
		if !decision.Allowed {
			return &RateLimitError{RetryAfter: decision.RetryAfter}
		}
	}
	target, err := s.repo.LookupResetTarget(ctx, email)
	if err != nil {
		// uniform: do not reveal
		return nil //nolint:nilerr
	}
	if target.Status != "active" {
		return nil
	}
	now := s.clock.Now()
	tokenID := newUUID()
	token := newToken()
	tokenHash := sha256Bytes(token)
	expiresAt := now.Add(resetExpiry)
	msg := outbox.Message{
		ID:             newUUID(),
		OrganizationID: target.OrganizationID,
		UserID:         target.UserID,
		Kind:           "password_reset",
		Recipient:      target.Email,
		Subject:        "Password reset",
		TextBody:       fmt.Sprintf("Reset your password: %s/reset-password/%s", s.baseURL, token),
		NextAttemptAt:  now,
	}
	return s.repo.RequestPasswordReset(ctx, target.OrganizationID, target.UserID, now, tokenID, tokenHash, expiresAt, msg)
}

func (s *AccessService) ResetPassword(ctx context.Context, token, password string) error {
	if err := ValidatePassword(password); err != nil {
		return err
	}
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return err
	}
	tokenHash := sha256Bytes(token)
	info, err := s.repo.LookupPasswordReset(ctx, tokenHash)
	if err != nil {
		return errors.New("identity: invalid reset token")
	}
	now := s.clock.Now()
	if info.UsedAt != nil || now.After(info.ExpiresAt) {
		return errors.New("identity: invalid reset token")
	}
	return s.repo.ResetPassword(ctx, info.OrganizationID, info.UserID, info.TokenID, hash, now)
}

func sha256Bytes(b string) []byte {
	h := sha256.Sum256([]byte(b))
	out := make([]byte, len(h))
	copy(out, h[:])
	return out
}

func invitationURL(baseURL, token string) string {
	return fmt.Sprintf("%s/invitations/accept#token=%s", baseURL, token)
}

func newToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func newUUID() string {
	// time-ordered uuid v7 would be nicer; use v4 via crypto/rand for now
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
