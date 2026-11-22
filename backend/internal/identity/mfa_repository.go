package identity

import (
	"context"
	"database/sql"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/platform/db"
	"github.com/TaushifReza/email-platform/backend/internal/platform/db/sqlc"
)

// MFARepository is the persistence interface for MFA state.
type MFARepository interface {
	GetMFAState(ctx context.Context, organizationID, userID string) (MFAState, error)
	SetTOTPSeed(ctx context.Context, organizationID, userID string, seed []byte, keyID string, now time.Time) (int64, error) //nolint:lll
	ConfirmEnrollment(ctx context.Context, organizationID, userID string, now time.Time, codeHashes []string) (int64, error) //nolint:lll
	ConsumeTOTPStep(ctx context.Context, organizationID, userID string, step int64) (int64, error)
	ConsumeRecoveryCode(ctx context.Context, organizationID, userID string, hash string) (int64, error)
}

type mfaRepository struct {
	db *db.DB
}

// NewMFARepository wraps the shared database.
func NewMFARepository(database *db.DB) MFARepository {
	return &mfaRepository{db: database}
}

func (r *mfaRepository) GetMFAState(ctx context.Context, organizationID, userID string) (MFAState, error) {
	var state MFAState
	err := r.db.WithOrg(ctx, organizationID, func(tx *sql.Tx) error {
		row, err := sqlc.New(tx).GetMFAState(ctx, sqlc.GetMFAStateParams{
			OrganizationID: organizationID,
			ID:             userID,
		})
		if err != nil {
			return err
		}
		state.SeedEncrypted = row.TotpSeedEncrypted
		if row.TotpKeyID.Valid {
			state.KeyID = row.TotpKeyID.String
		}
		if row.TotpEnrolledAt.Valid {
			t := row.TotpEnrolledAt.Time
			state.EnrolledAt = &t
		}
		if row.TotpLastUsedStep.Valid {
			state.LastUsedStep = row.TotpLastUsedStep.Int64
		}
		state.RecoveryHashes = row.RecoveryCodeHashes
		if row.LockedUntil.Valid {
			t := row.LockedUntil.Time
			state.LockedUntil = &t
		}
		return nil
	})
	if err != nil {
		return MFAState{}, err
	}
	return state, nil
}

func (r *mfaRepository) SetTOTPSeed(ctx context.Context, organizationID, userID string, seed []byte, keyID string, _ time.Time) (int64, error) { //nolint:lll
	var n int64
	err := r.db.WithOrg(ctx, organizationID, func(tx *sql.Tx) error {
		affected, err := sqlc.New(tx).SetTOTPSeed(ctx, sqlc.SetTOTPSeedParams{
			OrganizationID: organizationID,
			ID:             userID,
			Seed:           seed,
			KeyID:          keyID,
		})
		if err != nil {
			return err
		}
		n = affected
		return nil
	})
	return n, err
}

func (r *mfaRepository) ConfirmEnrollment(ctx context.Context, organizationID, userID string, now time.Time, codeHashes []string) (int64, error) { //nolint:lll
	var n int64
	err := r.db.WithOrg(ctx, organizationID, func(tx *sql.Tx) error {
		affected, err := sqlc.New(tx).ConfirmTOTPEnrollment(ctx, sqlc.ConfirmTOTPEnrollmentParams{
			OrganizationID: organizationID,
			ID:             userID,
			Now:            now,
			Codes:          codeHashes,
		})
		if err != nil {
			return err
		}
		n = affected
		return nil
	})
	return n, err
}

func (r *mfaRepository) ConsumeTOTPStep(ctx context.Context, organizationID, userID string, step int64) (int64, error) {
	var n int64
	err := r.db.WithOrg(ctx, organizationID, func(tx *sql.Tx) error {
		affected, err := sqlc.New(tx).ConsumeTOTPStep(ctx, sqlc.ConsumeTOTPStepParams{
			OrganizationID: organizationID,
			ID:             userID,
			Step:           step,
		})
		if err != nil {
			return err
		}
		n = affected
		return nil
	})
	return n, err
}

func (r *mfaRepository) ConsumeRecoveryCode(ctx context.Context, organizationID, userID, hash string) (int64, error) { //nolint:lll
	var n int64
	err := r.db.WithOrg(ctx, organizationID, func(tx *sql.Tx) error {
		affected, err := sqlc.New(tx).ConsumeRecoveryCode(ctx, sqlc.ConsumeRecoveryCodeParams{
			OrganizationID: organizationID,
			ID:             userID,
			Hash:           hash,
		})
		if err != nil {
			return err
		}
		n = affected
		return nil
	})
	return n, err
}
