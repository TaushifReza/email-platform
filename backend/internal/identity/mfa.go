package identity

import (
	"context"
	"crypto/subtle"
	"encoding/base32"
	"errors"
	"strings"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/platform/clock"
	"github.com/TaushifReza/email-platform/backend/internal/platform/crypto"
	"github.com/TaushifReza/email-platform/backend/internal/platform/ratelimit"
)

// Errors for MFA operations.
var (
	ErrMFAInvalidCode = errors.New("identity: invalid mfa code")
	ErrMFAReplay      = errors.New("identity: code already used")
	ErrMFALocked      = errors.New("identity: account locked")
	ErrMFAStepUp      = errors.New("identity: step-up required")
)

// MFAState holds persisted MFA data.
type MFAState struct {
	SeedEncrypted  []byte
	KeyID          string
	EnrolledAt     *time.Time
	LastUsedStep   int64
	RecoveryHashes []string
	LockedUntil    *time.Time
}

// MFAEnrollment is returned when enrollment starts.
type MFAEnrollment struct {
	Secret     string
	OtpauthURI string
}

// MFAService handles TOTP enrollment, verification and step-up.
type MFAService struct {
	repo     MFARepository
	sessions SessionRepository
	hasher   PasswordVerifier
	limiter  RateLimiter
	clock    clock.Clock
	keyring  *crypto.Keyring
}

// NewMFAService creates an MFA service.
func NewMFAService(repo MFARepository, sessions SessionRepository, hasher PasswordVerifier,
	limiter RateLimiter, clk clock.Clock, keyring *crypto.Keyring,
) *MFAService {
	if clk == nil {
		clk = clock.System{}
	}
	return &MFAService{repo: repo, sessions: sessions, hasher: hasher, limiter: limiter, clock: clk, keyring: keyring}
}

func (s *MFAService) BeginEnrollment(ctx context.Context, actor Actor) (MFAEnrollment, error) {
	secret, err := TOTPSecret()
	if err != nil {
		return MFAEnrollment{}, err
	}
	seed, _ := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret) //nolint:errcheck
	ct, keyID, err := s.keyring.Encrypt("totp_seed", crypto.AAD{
		Table: "users", RowID: actor.UserID, Column: "totp_seed", OrganizationID: actor.OrganizationID,
	}, seed)
	if err != nil {
		return MFAEnrollment{}, err
	}
	now := s.clock.Now()
	n, err := s.repo.SetTOTPSeed(ctx, actor.OrganizationID, actor.UserID, ct, keyID, now)
	if err != nil {
		return MFAEnrollment{}, err
	}
	if n == 0 {
		return MFAEnrollment{}, errors.New("identity: enrollment not allowed")
	}
	return MFAEnrollment{
		Secret:     secret,
		OtpauthURI: TOTPOtpauthURI("Email Platform", actor.UserID, secret),
	}, nil
}

func (s *MFAService) ConfirmEnrollment(ctx context.Context, actor Actor, code string) ([]string, error) {
	state, err := s.repo.GetMFAState(ctx, actor.OrganizationID, actor.UserID)
	if err != nil {
		return nil, err
	}
	if state.EnrolledAt != nil {
		return nil, errors.New("identity: already enrolled")
	}
	if len(state.SeedEncrypted) == 0 {
		return nil, errors.New("identity: no pending enrollment")
	}
	seed, err := s.keyring.Decrypt("totp_seed", crypto.AAD{
		Table: "users", RowID: actor.UserID, Column: "totp_seed", OrganizationID: actor.OrganizationID,
	}, state.SeedEncrypted, state.KeyID)
	if err != nil {
		return nil, err
	}
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(seed)
	step, ok := ValidateTOTP(secret, code, s.clock.Now())
	if !ok {
		return nil, ErrMFAInvalidCode
	}
	// replay guard
	n, err := s.repo.ConsumeTOTPStep(ctx, actor.OrganizationID, actor.UserID, step)
	if err != nil || n == 0 {
		return nil, ErrMFAReplay
	}
	codes, err := GenerateRecoveryCodes(10)
	if err != nil {
		return nil, err
	}
	hashes := make([]string, len(codes))
	for i, c := range codes {
		hashes[i] = HashRecoveryCode(c)
	}
	now := s.clock.Now()
	nn, err := s.repo.ConfirmEnrollment(ctx, actor.OrganizationID, actor.UserID, now, hashes)
	if err != nil || nn == 0 {
		return nil, errors.New("identity: enrollment confirm failed")
	}
	_ = s.sessions.MarkSessionMFAVerified(ctx, actor.OrganizationID, actor.SessionID, now) //nolint:errcheck
	return codes, nil
}

func (s *MFAService) VerifyMFA(ctx context.Context, ps PendingSession, code string, info ClientInfo) error {
	state, err := s.repo.GetMFAState(ctx, ps.OrganizationID, ps.UserID)
	if err != nil {
		return err
	}
	now := s.clock.Now()
	if state.LockedUntil != nil && state.LockedUntil.After(now) {
		return ErrMFALocked
	}
	// rate limit
	if s.limiter != nil {
		keys := []ratelimit.Key{{Name: "mfa:account:" + ps.UserID, Limit: 5, Window: time.Minute}}
		if info.IP != "" {
			keys = append(keys, ratelimit.Key{Name: "mfa:ip:" + info.IP, Limit: 5, Window: time.Minute})
		}
		decision, err := s.limiter.Allow(ctx, keys)
		if err != nil {
			return err
		}
		if !decision.Allowed {
			return &RateLimitError{RetryAfter: decision.RetryAfter}
		}
	}
	code = strings.TrimSpace(code)
	// TOTP path
	hasSeed := state.EnrolledAt != nil && len(state.SeedEncrypted) > 0
	if hasSeed {
		step, ok := s.tryTOTP(state, ps.OrganizationID, ps.UserID, code, now)
		_ = step
		if ok {
			n, err := s.repo.ConsumeTOTPStep(ctx, ps.OrganizationID, ps.UserID, step)
			if err == nil && n > 0 {
				_ = s.sessions.MarkSessionMFAVerified(ctx, ps.OrganizationID, ps.SessionID, now) //nolint:errcheck
				return nil
			}
			return ErrMFAReplay
		}
	}
	// recovery code path
	hash := HashRecoveryCode(code)
	for _, h := range state.RecoveryHashes {
		if subtle.ConstantTimeCompare([]byte(h), []byte(hash)) == 1 {
			n, err := s.repo.ConsumeRecoveryCode(ctx, ps.OrganizationID, ps.UserID, hash)
			if err == nil && n > 0 {
				_ = s.sessions.MarkSessionMFAVerified(ctx, ps.OrganizationID, ps.SessionID, now) //nolint:errcheck
				return nil
			}
		}
	}
	return ErrMFAInvalidCode
}

func (s *MFAService) StepUp(ctx context.Context, actor Actor, _ string, code string) error {
	// Password verification is handled by the caller; here we require a verified TOTP or recovery code.
	state, err := s.repo.GetMFAState(ctx, actor.OrganizationID, actor.UserID)
	if err != nil {
		return err
	}
	now := s.clock.Now()
	if state.LockedUntil != nil && state.LockedUntil.After(now) {
		return ErrMFALocked
	}
	// verify second factor
	ok := false
	var step int64
	hasSeed := state.EnrolledAt != nil && len(state.SeedEncrypted) > 0
	if hasSeed {
		st, v := s.tryTOTP(state, actor.OrganizationID, actor.UserID, code, now)
		step, ok = st, v
	}
	if !ok {
		hash := HashRecoveryCode(code)
		for _, h := range state.RecoveryHashes {
			if subtle.ConstantTimeCompare([]byte(h), []byte(hash)) == 1 {
				// consume recovery code on step-up
				n, _ := s.repo.ConsumeRecoveryCode(ctx, actor.OrganizationID, actor.UserID, hash) //nolint:errcheck
				if n > 0 {
					ok = true
				}
			}
		}
	}
	if !ok {
		return ErrMFAInvalidCode
	}
	if step > 0 {
		n, err := s.repo.ConsumeTOTPStep(ctx, actor.OrganizationID, actor.UserID, step)
		if err != nil || n == 0 {
			return ErrMFAReplay
		}
	}
	until := now.Add(5 * time.Minute)
	return s.sessions.SetSessionStepUp(ctx, actor.OrganizationID, actor.SessionID, until, now)
}

func (s *MFAService) tryTOTP(state MFAState, organizationID, userID string, code string, now time.Time) (int64, bool) {
	seed, err := s.keyring.Decrypt("totp_seed", crypto.AAD{
		Table: "users", RowID: userID, Column: "totp_seed", OrganizationID: organizationID,
	}, state.SeedEncrypted, state.KeyID)
	if err != nil {
		return 0, false
	}
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(seed)
	return ValidateTOTP(secret, code, now)
}
