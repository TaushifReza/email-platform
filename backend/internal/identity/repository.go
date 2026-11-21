package identity

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/platform/db"
	"github.com/TaushifReza/email-platform/backend/internal/platform/db/sqlc"
)

// Repository implements identity persistence.
type repository struct {
	db *db.DB
}

// NewRepository creates an identity repository.
func NewRepository(database *db.DB) Repository {
	return &repository{db: database}
}

type lookupPayload struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organization_id"`
	PasswordHash   *string    `json:"password_hash"`
	Status         string     `json:"status"`
	Role           string     `json:"role"`
	FailedCount    int        `json:"failed_login_count"`
	LockedUntil    *time.Time `json:"locked_until"`
	TOTPEnrolled   *time.Time `json:"totp_enrolled_at"`
}

func (p lookupPayload) toCredentials() Credentials {
	hash := ""
	if p.PasswordHash != nil {
		hash = *p.PasswordHash
	}
	return Credentials{
		ID:               p.ID,
		OrganizationID:   p.OrganizationID,
		PasswordHash:     hash,
		Status:           p.Status,
		Role:             p.Role,
		FailedLoginCount: p.FailedCount,
		LockedUntil:      p.LockedUntil,
		MFAEnrolled:      p.TOTPEnrolled != nil,
	}
}

// LookupForLogin finds a user without tenant context.
func (r *repository) LookupForLogin(ctx context.Context, emailNormalized string) (Credentials, error) {
	var creds Credentials
	err := r.db.WithPreAuth(ctx, func(tx *sql.Tx) error {
		payloadRaw, err := sqlc.New(tx).LookupUserForLogin(ctx, emailNormalized)
		if err != nil {
			return err
		}
		if len(payloadRaw) == 0 || string(payloadRaw) == "null" {
			return ErrUserNotFound
		}
		var payload lookupPayload
		if err := json.Unmarshal(payloadRaw, &payload); err != nil {
			return err
		}
		creds = payload.toCredentials()
		return nil
	})
	if err != nil {
		return Credentials{}, err
	}
	return creds, nil
}

// RecordFailedLogin updates failure state.
func (r *repository) RecordFailedLogin(ctx context.Context,
	organizationID, userID string, now time.Time, threshold int,
	lockUntil time.Time,
) error {
	return r.db.WithOrg(ctx, organizationID, func(tx *sql.Tx) error {
		return sqlc.New(tx).RecordFailedLogin(ctx, sqlc.RecordFailedLoginParams{
			Now:            now,
			Threshold:      int32(threshold), // #nosec G115
			LockUntil:      lockUntil,
			OrganizationID: organizationID,
			ID:             userID,
		})
	})
}

// ClearFailedLogins resets failure state.
func (r *repository) ClearFailedLogins(ctx context.Context, organizationID, userID string) error {
	return r.db.WithOrg(ctx, organizationID, func(tx *sql.Tx) error {
		return sqlc.New(tx).ClearFailedLogins(ctx, sqlc.ClearFailedLoginsParams{
			OrganizationID: organizationID,
			ID:             userID,
		})
	})
}
