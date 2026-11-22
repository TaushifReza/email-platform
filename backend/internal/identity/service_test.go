package identity

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/platform/ratelimit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	creds       Credentials
	lookupErr   error
	failErr     error
	clearErr    error
	failedCalls int
	clearCalls  int
}

func (f *fakeRepo) LookupForLogin(_ context.Context, _ string) (Credentials, error) {
	return f.creds, f.lookupErr
}

func (f *fakeRepo) RecordFailedLogin(_ context.Context, _ string, _ string, _ time.Time, _ int, _ time.Time) error {
	f.failedCalls++
	return f.failErr
}

func (f *fakeRepo) ClearFailedLogins(_ context.Context, _, _ string) error {
	f.clearCalls++
	return f.clearErr
}

type fakeVerifier struct {
	valid bool
	err   error
	dummy int
}

func (f *fakeVerifier) Verify(_ string, _ string) (bool, error) {
	return f.valid, f.err
}

func (f *fakeVerifier) VerifyDummy(_ string) error {
	f.dummy++
	return nil
}

type fakeLimiter struct {
	decision ratelimit.Decision
	err      error
	keys     []ratelimit.Key
}

func (f *fakeLimiter) Allow(_ context.Context, keys []ratelimit.Key) (ratelimit.Decision, error) {
	f.keys = keys
	return f.decision, f.err
}

type fakeClock struct{ now time.Time }

func (f fakeClock) Now() time.Time { return f.now }

func TestAUTH_01_LoginValid(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeRepo{creds: Credentials{
		ID: "u1", OrganizationID: "o1", PasswordHash: "hash",
		Status: statusActive, Role: "viewer",
	}}
	verifier := &fakeVerifier{valid: true}
	svc := NewService(repo, nil, verifier,
		&fakeLimiter{decision: ratelimit.Decision{Allowed: true}},
		fakeClock{now: now}, LockoutPolicy{Threshold: 2, Duration: time.Minute}, DefaultSessionPolicy)
	res, err := svc.Login(context.Background(), LoginRequest{Email: "User@Example.com", Password: "secret"})
	require.NoError(t, err)
	assert.Equal(t, "u1", res.UserID)
	assert.Equal(t, "o1", res.OrganizationID)
	assert.False(t, res.MFARequired)
	assert.Equal(t, 1, repo.clearCalls)
}

func TestAUTH_01_LoginNormalizesEmail(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeRepo{creds: Credentials{
		ID: "u1", OrganizationID: "o1", PasswordHash: "hash",
		Status: statusActive, Role: "viewer",
	}}
	verifier := &fakeVerifier{valid: true}
	svc := NewService(repo, nil, verifier,
		&fakeLimiter{decision: ratelimit.Decision{Allowed: true}},
		fakeClock{now: now}, LockoutPolicy{Threshold: 2, Duration: time.Minute}, DefaultSessionPolicy)
	_, err := svc.Login(context.Background(), LoginRequest{Email: "  USER@Example.com  ", Password: "secret"})
	require.NoError(t, err)
}

func TestAUTH_02_LoginMFAForOwnerAdmin(t *testing.T) {
	now := time.Now().UTC()
	for _, role := range []string{roleOwner, roleAdmin} {
		repo := &fakeRepo{creds: Credentials{
			ID: "u1", OrganizationID: "o1", PasswordHash: "hash",
			Status: statusActive, Role: role,
		}}
		verifier := &fakeVerifier{valid: true}
		svc := NewService(repo, nil, verifier,
			&fakeLimiter{decision: ratelimit.Decision{Allowed: true}},
			fakeClock{now: now}, LockoutPolicy{Threshold: 2, Duration: time.Minute}, DefaultSessionPolicy)
		res, err := svc.Login(context.Background(), LoginRequest{Email: "a@example.com", Password: "secret"})
		require.NoError(t, err)
		assert.True(t, res.MFARequired)
	}
	repo := &fakeRepo{creds: Credentials{
		ID: "u1", OrganizationID: "o1", PasswordHash: "hash",
		Status: statusActive, Role: "viewer",
	}}
	verifier := &fakeVerifier{valid: true}
	svc := NewService(repo, nil, verifier,
		&fakeLimiter{decision: ratelimit.Decision{Allowed: true}},
		fakeClock{now: now}, LockoutPolicy{Threshold: 2, Duration: time.Minute}, DefaultSessionPolicy)
	res, err := svc.Login(context.Background(), LoginRequest{Email: "a@example.com", Password: "secret"})
	require.NoError(t, err)
	assert.False(t, res.MFARequired)
}

func TestAUTH_04_LoginUnknownUniform(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeRepo{lookupErr: ErrUserNotFound}
	verifier := &fakeVerifier{}
	svc := NewService(repo, nil, verifier,
		&fakeLimiter{decision: ratelimit.Decision{Allowed: true}},
		fakeClock{now: now}, LockoutPolicy{Threshold: 2, Duration: time.Minute}, DefaultSessionPolicy)
	_, err := svc.Login(context.Background(), LoginRequest{Email: "missing@example.com", Password: "secret"})
	require.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Equal(t, 1, verifier.dummy)
}

func TestAUTH_04_LoginWrongUniform(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeRepo{creds: Credentials{
		ID: "u1", OrganizationID: "o1", PasswordHash: "hash",
		Status: statusActive, Role: "viewer",
	}}
	verifier := &fakeVerifier{valid: false}
	svc := NewService(repo, nil, verifier,
		&fakeLimiter{decision: ratelimit.Decision{Allowed: true}},
		fakeClock{now: now}, LockoutPolicy{Threshold: 2, Duration: time.Minute}, DefaultSessionPolicy)
	_, err := svc.Login(context.Background(), LoginRequest{Email: "a@example.com", Password: "wrong"})
	require.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Equal(t, 1, repo.failedCalls)
}

func TestAUTH_04_LoginInactiveUniform(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeRepo{creds: Credentials{
		ID: "u1", OrganizationID: "o1", PasswordHash: "hash",
		Status: "deactivated", Role: "viewer",
	}}
	verifier := &fakeVerifier{valid: true}
	svc := NewService(repo, nil, verifier,
		&fakeLimiter{decision: ratelimit.Decision{Allowed: true}},
		fakeClock{now: now}, LockoutPolicy{Threshold: 2, Duration: time.Minute}, DefaultSessionPolicy)
	_, err := svc.Login(context.Background(), LoginRequest{Email: "a@example.com", Password: "secret"})
	require.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Equal(t, 0, repo.failedCalls)
}

func TestAUTH_04_LoginLockedUniform(t *testing.T) {
	now := time.Now().UTC()
	locked := now.Add(time.Minute)
	repo := &fakeRepo{creds: Credentials{
		ID: "u1", OrganizationID: "o1", PasswordHash: "hash",
		Status: statusActive, Role: "viewer", LockedUntil: &locked,
	}}
	verifier := &fakeVerifier{valid: true}
	svc := NewService(repo, nil, verifier,
		&fakeLimiter{decision: ratelimit.Decision{Allowed: true}},
		fakeClock{now: now}, LockoutPolicy{Threshold: 2, Duration: time.Minute}, DefaultSessionPolicy)
	_, err := svc.Login(context.Background(), LoginRequest{Email: "a@example.com", Password: "secret"})
	require.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Equal(t, 0, repo.failedCalls)
}

func TestAUTH_04_LoginLocksAfterThreshold(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeRepo{creds: Credentials{
		ID: "u1", OrganizationID: "o1", PasswordHash: "hash",
		Status: statusActive, Role: "viewer",
	}}
	verifier := &fakeVerifier{valid: false}
	svc := NewService(repo, nil, verifier,
		&fakeLimiter{decision: ratelimit.Decision{Allowed: true}},
		fakeClock{now: now}, LockoutPolicy{Threshold: 2, Duration: time.Minute}, DefaultSessionPolicy)
	_, _ = svc.Login(context.Background(), LoginRequest{Email: "a@example.com", Password: "wrong"}) //nolint:errcheck
	_, _ = svc.Login(context.Background(), LoginRequest{Email: "a@example.com", Password: "wrong"}) //nolint:errcheck
	assert.Equal(t, 2, repo.failedCalls)
}

func TestAUTH_04_LoginClearsFailures(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeRepo{creds: Credentials{
		ID: "u1", OrganizationID: "o1", PasswordHash: "hash",
		Status: statusActive, Role: "viewer", FailedLoginCount: 3,
	}}
	verifier := &fakeVerifier{valid: true}
	svc := NewService(repo, nil, verifier,
		&fakeLimiter{decision: ratelimit.Decision{Allowed: true}},
		fakeClock{now: now}, LockoutPolicy{Threshold: 2, Duration: time.Minute}, DefaultSessionPolicy)
	_, err := svc.Login(context.Background(), LoginRequest{Email: "a@example.com", Password: "secret"})
	require.NoError(t, err)
	assert.Equal(t, 1, repo.clearCalls)
}

func TestAUTH_04_LoginLockExpiryResets(t *testing.T) {
	// covered by repository SQL; service passes now.
	now := time.Now().UTC()
	repo := &fakeRepo{creds: Credentials{
		ID: "u1", OrganizationID: "o1", PasswordHash: "hash",
		Status: statusActive, Role: "viewer",
	}}
	verifier := &fakeVerifier{valid: false}
	svc := NewService(repo, nil, verifier,
		&fakeLimiter{decision: ratelimit.Decision{Allowed: true}},
		fakeClock{now: now}, LockoutPolicy{Threshold: 2, Duration: time.Minute}, DefaultSessionPolicy)
	_, _ = svc.Login(context.Background(), LoginRequest{Email: "a@example.com", Password: "wrong"}) //nolint:errcheck
	assert.Equal(t, 1, repo.failedCalls)
}

func TestSEC_12_LoginRateLimited(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeRepo{creds: Credentials{
		ID: "u1", OrganizationID: "o1", PasswordHash: "hash",
		Status: statusActive, Role: "viewer",
	}}
	verifier := &fakeVerifier{valid: true}
	limiter := &fakeLimiter{decision: ratelimit.Decision{Allowed: false, RetryAfter: 30 * time.Second}}
	svc := NewService(repo, nil, verifier, limiter, fakeClock{now: now}, LockoutPolicy{Threshold: 2, Duration: time.Minute}, DefaultSessionPolicy) //nolint:lll
	_, err := svc.Login(context.Background(), LoginRequest{Email: "a@example.com", Password: "secret"})
	var rl *RateLimitError
	ok := errors.As(err, &rl)
	assert.True(t, ok)
	assert.Equal(t, 30*time.Second, rl.RetryAfter)
	assert.Equal(t, 0, repo.failedCalls)
}

func TestSEC_12_LoginLimiterUnavailableFailsClosed(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeRepo{creds: Credentials{
		ID: "u1", OrganizationID: "o1", PasswordHash: "hash",
		Status: statusActive, Role: "viewer",
	}}
	verifier := &fakeVerifier{valid: true}
	limiter := &fakeLimiter{err: ratelimit.ErrCounterUnavailable}
	svc := NewService(repo, nil, verifier, limiter, fakeClock{now: now}, LockoutPolicy{Threshold: 2, Duration: time.Minute}, DefaultSessionPolicy) //nolint:lll
	_, err := svc.Login(context.Background(), LoginRequest{Email: "a@example.com", Password: "secret"})
	require.ErrorIs(t, err, ErrRateLimiterUnavailable)
}

func TestSEC_16_RateLimitKeyHidesEmail(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeRepo{creds: Credentials{
		ID: "u1", OrganizationID: "o1", PasswordHash: "hash",
		Status: statusActive, Role: "viewer",
	}}
	verifier := &fakeVerifier{valid: true}
	limiter := &fakeLimiter{decision: ratelimit.Decision{Allowed: true}}
	svc := NewService(repo, nil, verifier, limiter, fakeClock{now: now}, LockoutPolicy{Threshold: 2, Duration: time.Minute}, DefaultSessionPolicy) //nolint:lll
	_, err := svc.Login(context.Background(), LoginRequest{Email: "secret@example.com", Password: "password"})
	require.NoError(t, err)
	require.Len(t, limiter.keys, 1)
	assert.NotContains(t, limiter.keys[0].Name, "secret@example.com")
}

func TestAUTH_04_LoginNoLimiter(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeRepo{creds: Credentials{
		ID: "u1", OrganizationID: "o1", PasswordHash: "hash",
		Status: statusActive, Role: "viewer",
	}}
	verifier := &fakeVerifier{valid: true}
	svc := NewService(repo, nil, verifier, nil, fakeClock{now: now}, LockoutPolicy{Threshold: 2, Duration: time.Minute}, DefaultSessionPolicy) //nolint:lll
	_, err := svc.Login(context.Background(), LoginRequest{Email: "a@example.com", Password: "secret"})
	require.ErrorIs(t, err, ErrRateLimiterUnavailable)
}

func TestAUTH_04_LoginMFAEnrolled(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeRepo{creds: Credentials{
		ID: "u1", OrganizationID: "o1", PasswordHash: "hash",
		Status: statusActive, Role: "viewer", MFAEnrolled: true,
	}}
	verifier := &fakeVerifier{valid: true}
	svc := NewService(repo, nil, verifier,
		&fakeLimiter{decision: ratelimit.Decision{Allowed: true}},
		fakeClock{now: now}, LockoutPolicy{Threshold: 2, Duration: time.Minute}, DefaultSessionPolicy)
	res, err := svc.Login(context.Background(), LoginRequest{Email: "a@example.com", Password: "secret"})
	require.NoError(t, err)
	assert.True(t, res.MFARequired)
}
