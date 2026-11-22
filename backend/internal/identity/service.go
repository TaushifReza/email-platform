package identity

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/platform/clock"
	"github.com/TaushifReza/email-platform/backend/internal/platform/ratelimit"
)

const (
	roleOwner    = "owner"
	roleAdmin    = "admin"
	statusActive = "active"

	loginRateLimit  = 5
	loginRateWindow = time.Minute
)

// ErrInvalidCredentials is returned for any failed login.
var ErrInvalidCredentials = errors.New("identity: invalid credentials")

// ErrRateLimited is returned when login is rate limited.
var ErrRateLimited = errors.New("identity: rate limited")

// ErrRateLimiterUnavailable is returned when rate limiting cannot run.
var ErrRateLimiterUnavailable = errors.New("identity: rate limiter unavailable")

// ErrUserNotFound is returned when no matching user exists.
var ErrUserNotFound = errors.New("identity: user not found")

// LoginRequest contains credentials.
type LoginRequest struct {
	Email    string
	Password string
}

// LoginResult describes a successful password verification.
type LoginResult struct {
	UserID                string
	OrganizationID        string
	Role                  string
	MFARequired           bool
	MFAEnrollmentRequired bool
}

// RateLimitError carries a retry delay.
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string { return ErrRateLimited.Error() }

func (e *RateLimitError) Is(target error) bool { return target == ErrRateLimited }

// Credentials is the pre-auth user data.
type Credentials struct {
	ID               string
	OrganizationID   string
	PasswordHash     string
	Status           string
	Role             string
	FailedLoginCount int
	LockedUntil      *time.Time
	MFAEnrolled      bool
}

// Repository is the identity persistence interface.
type Repository interface {
	LookupForLogin(ctx context.Context, emailNormalized string) (Credentials, error)
	RecordFailedLogin(ctx context.Context, organizationID, userID string,
		now time.Time, threshold int, lockUntil time.Time) error
	ClearFailedLogins(ctx context.Context, organizationID, userID string) error
}

// RateLimiter is the rate limiting interface used by identity.
type RateLimiter interface {
	Allow(ctx context.Context, keys []ratelimit.Key) (ratelimit.Decision, error)
}

// PasswordVerifier verifies passwords.
type PasswordVerifier interface {
	Verify(encoded, password string) (bool, error)
	VerifyDummy(password string) error
}

// LockoutPolicy configures lockout.
type LockoutPolicy struct {
	Threshold int
	Duration  time.Duration
}

// DefaultLockoutPolicy is the production default.
//
//nolint:gochecknoglobals
var DefaultLockoutPolicy = LockoutPolicy{
	Threshold: 5,
	Duration:  15 * time.Minute,
}

// Service authenticates users.
type Service struct {
	repo     Repository
	sessions SessionRepository
	hasher   PasswordVerifier
	limiter  RateLimiter
	clock    clock.Clock
	lockout  LockoutPolicy
	session  SessionPolicy
}

// NewService creates an identity service.
func NewService(repo Repository, sessions SessionRepository, hasher PasswordVerifier,
	limiter RateLimiter, clk clock.Clock, lockout LockoutPolicy, sessionPolicy SessionPolicy,
) *Service {
	if clk == nil {
		clk = clock.System{}
	}
	if lockout.Threshold <= 0 {
		lockout = DefaultLockoutPolicy
	}
	if lockout.Duration <= 0 {
		lockout = DefaultLockoutPolicy
	}
	if sessionPolicy.PrivilegedIdle <= 0 {
		sessionPolicy = DefaultSessionPolicy
	}
	return &Service{repo: repo, sessions: sessions, hasher: hasher, limiter: limiter, clock: clk, lockout: lockout, session: sessionPolicy} //nolint:lll
}

// Login verifies credentials and applies lockout and rate limits.
func (s *Service) Login(ctx context.Context, req LoginRequest) (LoginResult, error) {
	email := normalizeEmail(req.Email)
	if email == "" {
		return LoginResult{}, ErrInvalidCredentials
	}
	if err := s.checkRateLimit(ctx, email, ""); err != nil {
		return LoginResult{}, err
	}

	creds, err := s.repo.LookupForLogin(ctx, email)
	if errors.Is(err, ErrUserNotFound) {
		_ = s.hasher.VerifyDummy(req.Password) //nolint:errcheck //nolint:errcheck
		return LoginResult{}, ErrInvalidCredentials
	}
	if err != nil {
		return LoginResult{}, fmt.Errorf("identity: lookup user: %w", err)
	}

	now := s.clock.Now()
	if creds.PasswordHash == "" {
		_ = s.hasher.VerifyDummy(req.Password) //nolint:errcheck //nolint:errcheck
		return LoginResult{}, ErrInvalidCredentials
	}
	valid, err := s.hasher.Verify(creds.PasswordHash, req.Password)
	if err != nil {
		return LoginResult{}, fmt.Errorf("identity: verify password: %w", err)
	}
	locked := creds.LockedUntil != nil && creds.LockedUntil.After(now)

	if !valid {
		if creds.Status == statusActive && !locked {
			if err := s.repo.RecordFailedLogin(ctx, creds.OrganizationID, creds.ID,
				now, s.lockout.Threshold, now.Add(s.lockout.Duration)); err != nil {
				return LoginResult{}, fmt.Errorf("identity: record failed login: %w", err)
			}
		}
		return LoginResult{}, ErrInvalidCredentials
	}
	if creds.Status != statusActive || locked {
		return LoginResult{}, ErrInvalidCredentials
	}

	if err := s.repo.ClearFailedLogins(ctx, creds.OrganizationID, creds.ID); err != nil {
		return LoginResult{}, fmt.Errorf("identity: clear failed logins: %w", err)
	}
	return LoginResult{
		UserID:                creds.ID,
		OrganizationID:        creds.OrganizationID,
		Role:                  creds.Role,
		MFARequired:           mfaRequired(creds.Role) || creds.MFAEnrolled,
		MFAEnrollmentRequired: mfaRequired(creds.Role) && !creds.MFAEnrolled,
	}, nil
}

func (s *Service) checkRateLimit(ctx context.Context, email, ip string) error {
	if s.limiter == nil {
		return ErrRateLimiterUnavailable
	}
	keys := []ratelimit.Key{{Name: loginAccountKey(email), Limit: loginRateLimit, Window: loginRateWindow}}
	if ip != "" {
		keys = append(keys, ratelimit.Key{Name: "login:ip:" + ip, Limit: loginRateLimit, Window: loginRateWindow})
	}
	decision, err := s.limiter.Allow(ctx, keys)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRateLimiterUnavailable, err)
	}
	if !decision.Allowed {
		return &RateLimitError{RetryAfter: decision.RetryAfter}
	}
	return nil
}

// SignIn authenticates and issues a session.
func (s *Service) SignIn(ctx context.Context, req LoginRequest, info ClientInfo) (SessionResult, error) {
	email := normalizeEmail(req.Email)
	if email == "" {
		return SessionResult{}, ErrInvalidCredentials
	}
	if err := s.checkRateLimit(ctx, email, info.IP); err != nil {
		return SessionResult{}, err
	}

	creds, err := s.repo.LookupForLogin(ctx, email)
	if errors.Is(err, ErrUserNotFound) {
		_ = s.hasher.VerifyDummy(req.Password) //nolint:errcheck
		return SessionResult{}, ErrInvalidCredentials
	}
	if err != nil {
		return SessionResult{}, fmt.Errorf("identity: lookup user: %w", err)
	}

	now := s.clock.Now()
	if creds.PasswordHash == "" {
		_ = s.hasher.VerifyDummy(req.Password) //nolint:errcheck
		return SessionResult{}, ErrInvalidCredentials
	}
	valid, err := s.hasher.Verify(creds.PasswordHash, req.Password)
	if err != nil {
		return SessionResult{}, fmt.Errorf("identity: verify password: %w", err)
	}
	locked := creds.LockedUntil != nil && creds.LockedUntil.After(now)

	if !valid {
		if creds.Status == statusActive && !locked {
			if err := s.repo.RecordFailedLogin(ctx, creds.OrganizationID, creds.ID,
				now, s.lockout.Threshold, now.Add(s.lockout.Duration)); err != nil {
				return SessionResult{}, fmt.Errorf("identity: record failed login: %w", err)
			}
		}
		return SessionResult{}, ErrInvalidCredentials
	}
	if creds.Status != statusActive || locked {
		return SessionResult{}, ErrInvalidCredentials
	}

	if err := s.repo.ClearFailedLogins(ctx, creds.OrganizationID, creds.ID); err != nil {
		return SessionResult{}, fmt.Errorf("identity: clear failed logins: %w", err)
	}

	result := LoginResult{
		UserID:                creds.ID,
		OrganizationID:        creds.OrganizationID,
		Role:                  creds.Role,
		MFARequired:           mfaRequired(creds.Role) || creds.MFAEnrolled,
		MFAEnrollmentRequired: mfaRequired(creds.Role) && !creds.MFAEnrolled,
	}

	sess, err := IssueSession(ctx, s.sessions, s.clock, s.session, result, info)
	if err != nil {
		return SessionResult{}, fmt.Errorf("identity: issue session: %w", err)
	}
	return sess, nil
}

// Authenticate validates a raw session token and returns an Actor.
func (s *Service) Authenticate(ctx context.Context, token string) (Actor, error) {
	actor, rec, err := ValidateSession(ctx, s.sessions, s.clock, token)
	if err != nil {
		if errors.Is(err, ErrMFARequired) {
			return Actor{}, ErrMFARequired
		}
		return Actor{}, ErrUnauthenticated
	}
	// Touch idle expiry best-effort.
	_ = TouchSessionIfNeeded(ctx, s.sessions, s.clock, s.session, rec) //nolint:errcheck
	return actor, nil
}

// AuthenticatePending validates a raw session token that has not yet completed 2FA.
func (s *Service) AuthenticatePending(ctx context.Context, token string) (PendingSession, error) {
	ps, _, err := ValidatePendingSession(ctx, s.sessions, s.clock, token)
	if err != nil {
		return PendingSession{}, ErrUnauthenticated
	}
	return ps, nil
}

// SignOut revokes a session by raw token.
func (s *Service) SignOut(ctx context.Context, token string) error {
	return RevokeSessionToken(ctx, s.sessions, s.clock, token)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func mfaRequired(role string) bool {
	return role == roleOwner || role == roleAdmin
}

func loginAccountKey(email string) string {
	sum := sha256.Sum256([]byte(email))
	return "login:account:" + hex.EncodeToString(sum[:])
}
