package identity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/platform/clock"
	"github.com/TaushifReza/email-platform/backend/internal/platform/db"
)

// ErrUnauthenticated is returned when no valid session exists.
var ErrUnauthenticated = errors.New("identity: unauthenticated")

// ErrMFARequired is returned when the session has not completed 2FA.
var ErrMFARequired = errors.New("identity: mfa required")

// Actor describes the authenticated caller.
type Actor struct {
	UserID          string
	OrganizationID  string
	Role            string
	SessionID       string
	StepUpExpiresAt *time.Time
}

// PendingSession is a session that has passed password auth but not 2FA.
type PendingSession struct {
	SessionID      string
	OrganizationID string
	UserID         string
	Role           string
	Token          string
}

// ClientInfo carries transport metadata for session creation.
type ClientInfo struct {
	IP        string
	UserAgent string
}

// SessionResult describes a newly created session.
type SessionResult struct {
	Token             string
	UserID            string
	OrganizationID    string
	Role              string
	MFARequired       bool
	MFAVerified       bool
	IdleExpiresAt     time.Time
	AbsoluteExpiresAt time.Time
}

// SessionRecord is the looked-up session data.
type SessionRecord struct {
	SessionID         string
	OrganizationID    string
	UserID            string
	MFAVerified       bool
	RevokedAt         *time.Time
	IdleExpiresAt     time.Time
	AbsoluteExpiresAt time.Time
	StepUpExpiresAt   *time.Time
	Role              string
	Status            string
}

// SessionRepository is the identity persistence interface for sessions.
type SessionRepository interface {
	CreateSession(ctx context.Context, params CreateSessionParams) error
	LookupSession(ctx context.Context, tokenHash []byte) (SessionRecord, error)
	TouchSession(ctx context.Context, params TouchSessionParams) error
	RevokeSession(ctx context.Context, params RevokeSessionParams) error
	RevokeAllUserSessions(ctx context.Context, organizationID, userID string, now time.Time) error
	MarkSessionMFAVerified(ctx context.Context, organizationID, sessionID string, now time.Time) error
	SetSessionStepUp(ctx context.Context, organizationID, sessionID string, until, now time.Time) error
}

type CreateSessionParams struct {
	ID                string
	OrganizationID    string
	UserID            string
	TokenHash         []byte
	MFAVerified       bool
	IP                string
	UserAgent         string
	IdleExpiresAt     time.Time
	AbsoluteExpiresAt time.Time
}

type TouchSessionParams struct {
	OrganizationID string
	SessionID      string
	Now            time.Time
	IdleExpiresAt  time.Time
}

type RevokeSessionParams struct {
	OrganizationID string
	TokenHash      []byte
	Now            time.Time
}

// SessionPolicy configures session expiry.
type SessionPolicy struct {
	PrivilegedIdle time.Duration
	PrivilegedAbs  time.Duration
	StandardIdle   time.Duration
	StandardAbs    time.Duration
}

// DefaultSessionPolicy is the production default.
//
//nolint:gochecknoglobals
var DefaultSessionPolicy = SessionPolicy{
	PrivilegedIdle: 30 * time.Minute,
	PrivilegedAbs:  12 * time.Hour,
	StandardIdle:   2 * time.Hour,
	StandardAbs:    24 * time.Hour,
}

func privilegedRole(role string) bool {
	return role == roleOwner || role == roleAdmin
}

func sessionPolicyFor(policy SessionPolicy, role string) (time.Duration, time.Duration) {
	if privilegedRole(role) {
		return policy.PrivilegedIdle, policy.PrivilegedAbs
	}
	return policy.StandardIdle, policy.StandardAbs
}

func generateToken() (string, []byte, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", nil, err
	}
	raw := base64.RawURLEncoding.EncodeToString(b)
	h := sha256.Sum256([]byte(raw))
	return raw, h[:], nil
}

func tokenHash(raw string) []byte {
	h := sha256.Sum256([]byte(raw))
	return h[:]
}

// IssueSession creates a session for a successful login.
func IssueSession(ctx context.Context, repo SessionRepository, clk clock.Clock, policy SessionPolicy, creds LoginResult, info ClientInfo) (SessionResult, error) { //nolint:lll
	idle, abs := sessionPolicyFor(policy, creds.Role)
	now := clk.Now()
	mfaVerified := !creds.MFARequired

	raw, hash, err := generateToken()
	if err != nil {
		return SessionResult{}, err
	}
	id := newID()
	idleExp := now.Add(idle)
	absExp := now.Add(abs)
	if err := repo.CreateSession(ctx, CreateSessionParams{
		ID:                id,
		OrganizationID:    creds.OrganizationID,
		UserID:            creds.UserID,
		TokenHash:         hash,
		MFAVerified:       mfaVerified,
		IP:                info.IP,
		UserAgent:         info.UserAgent,
		IdleExpiresAt:     idleExp,
		AbsoluteExpiresAt: absExp,
	}); err != nil {
		return SessionResult{}, err
	}
	return SessionResult{
		Token:             raw,
		UserID:            creds.UserID,
		OrganizationID:    creds.OrganizationID,
		Role:              creds.Role,
		MFARequired:       creds.MFARequired,
		MFAVerified:       mfaVerified,
		IdleExpiresAt:     idleExp,
		AbsoluteExpiresAt: absExp,
	}, nil
}

// ValidateSession validates a raw session token and returns an Actor.
func ValidateSession(ctx context.Context, repo SessionRepository, clk clock.Clock, token string) (Actor, SessionRecord, error) { //nolint:lll
	if token == "" {
		return Actor{}, SessionRecord{}, ErrUnauthenticated
	}
	rec, err := repo.LookupSession(ctx, tokenHash(token))
	if err != nil {
		return Actor{}, SessionRecord{}, ErrUnauthenticated
	}
	now := clk.Now()
	if rec.RevokedAt != nil {
		return Actor{}, rec, ErrUnauthenticated
	}
	if !rec.IdleExpiresAt.After(now) || !rec.AbsoluteExpiresAt.After(now) {
		return Actor{}, rec, ErrUnauthenticated
	}
	if rec.Status != statusActive {
		return Actor{}, rec, ErrUnauthenticated
	}
	if !rec.MFAVerified {
		return Actor{}, rec, ErrMFARequired
	}
	return Actor{
		UserID:          rec.UserID,
		OrganizationID:  rec.OrganizationID,
		Role:            rec.Role,
		SessionID:       rec.SessionID,
		StepUpExpiresAt: rec.StepUpExpiresAt,
	}, rec, nil
}

// ValidatePendingSession validates a session for the MFA-completion step.
func ValidatePendingSession(ctx context.Context, repo SessionRepository, clk clock.Clock, token string) (PendingSession, SessionRecord, error) { //nolint:lll
	if token == "" {
		return PendingSession{}, SessionRecord{}, ErrUnauthenticated
	}
	rec, err := repo.LookupSession(ctx, tokenHash(token))
	if err != nil {
		return PendingSession{}, SessionRecord{}, ErrUnauthenticated
	}
	now := clk.Now()
	if rec.RevokedAt != nil {
		return PendingSession{}, rec, ErrUnauthenticated
	}
	if !rec.IdleExpiresAt.After(now) || !rec.AbsoluteExpiresAt.After(now) {
		return PendingSession{}, rec, ErrUnauthenticated
	}
	if rec.Status != statusActive {
		return PendingSession{}, rec, ErrUnauthenticated
	}
	if rec.MFAVerified {
		return PendingSession{}, rec, ErrUnauthenticated
	}
	return PendingSession{
		SessionID:      rec.SessionID,
		OrganizationID: rec.OrganizationID,
		UserID:         rec.UserID,
		Role:           rec.Role,
		Token:          token,
	}, rec, nil
}

// TouchSessionIfNeeded extends the idle expiry (best-effort).
func TouchSessionIfNeeded(ctx context.Context, repo SessionRepository, clk clock.Clock, policy SessionPolicy, rec SessionRecord) error { //nolint:lll
	if rec.RevokedAt != nil {
		return nil
	}
	idle, _ := sessionPolicyFor(policy, rec.Role)
	now := clk.Now()
	newIdle := now.Add(idle)
	if newIdle.After(rec.AbsoluteExpiresAt) {
		newIdle = rec.AbsoluteExpiresAt
	}
	return repo.TouchSession(ctx, TouchSessionParams{
		OrganizationID: rec.OrganizationID,
		SessionID:      rec.SessionID,
		Now:            now,
		IdleExpiresAt:  newIdle,
	})
}

// RevokeSessionToken revokes a session by raw token.
func RevokeSessionToken(ctx context.Context, repo SessionRepository, clk clock.Clock, token string) error {
	if token == "" {
		return nil
	}
	h := tokenHash(token)
	// Lookup to get org; if not found, nothing to do.
	rec, err := repo.LookupSession(ctx, h)
	if err != nil {
		return nil //nolint:nilerr
	}
	return repo.RevokeSession(ctx, RevokeSessionParams{
		OrganizationID: rec.OrganizationID,
		TokenHash:      h,
		Now:            clk.Now(),
	})
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b) //nolint:errcheck
	return base64.RawURLEncoding.EncodeToString(b)
}

// Package-level reference retained for dependency wiring.
var _ = db.DB{}
