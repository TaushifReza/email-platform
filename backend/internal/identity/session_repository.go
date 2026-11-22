package identity

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/platform/db"
	"github.com/TaushifReza/email-platform/backend/internal/platform/db/sqlc"
)

type sessionRepository struct {
	db *db.DB
}

func NewSessionRepository(database *db.DB) SessionRepository {
	return &sessionRepository{db: database}
}

type sessionPayload struct {
	SessionID         string     `json:"session_id"`
	OrganizationID    string     `json:"organization_id"`
	UserID            string     `json:"user_id"`
	MFAVerified       bool       `json:"mfa_verified"`
	RevokedAt         *time.Time `json:"revoked_at"`
	IdleExpiresAt     time.Time  `json:"idle_expires_at"`
	AbsoluteExpiresAt time.Time  `json:"absolute_expires_at"`
	StepUpExpiresAt   *time.Time `json:"step_up_expires_at"`
	Role              string     `json:"role"`
	Status            string     `json:"status"`
}

func (r *sessionRepository) CreateSession(ctx context.Context, p CreateSessionParams) error {
	return r.db.WithOrg(ctx, p.OrganizationID, func(tx *sql.Tx) error {
		return sqlc.New(tx).CreateSession(ctx, sqlc.CreateSessionParams{
			ID:                p.ID,
			OrganizationID:    p.OrganizationID,
			UserID:            p.UserID,
			TokenHash:         p.TokenHash,
			MfaVerified:       p.MFAVerified,
			Ip:                p.IP,
			UserAgent:         p.UserAgent,
			IdleExpiresAt:     p.IdleExpiresAt,
			AbsoluteExpiresAt: p.AbsoluteExpiresAt,
		})
	})
}

func (r *sessionRepository) LookupSession(ctx context.Context, tokenHash []byte) (SessionRecord, error) {
	var rec SessionRecord
	err := r.db.WithPreAuth(ctx, func(tx *sql.Tx) error {
		raw, err := sqlc.New(tx).LookupSession(ctx, tokenHash)
		if err != nil {
			return err
		}
		if len(raw) == 0 || string(raw) == "null" {
			return ErrUserNotFound
		}
		var p sessionPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return err
		}
		rec = SessionRecord{ //nolint:gosimple
			SessionID:         p.SessionID,
			OrganizationID:    p.OrganizationID,
			UserID:            p.UserID,
			MFAVerified:       p.MFAVerified,
			RevokedAt:         p.RevokedAt,
			IdleExpiresAt:     p.IdleExpiresAt,
			AbsoluteExpiresAt: p.AbsoluteExpiresAt,
			StepUpExpiresAt:   p.StepUpExpiresAt,
			Role:              p.Role,
			Status:            p.Status,
		}
		return nil
	})
	if err != nil {
		return SessionRecord{}, err
	}
	return rec, nil
}

func (r *sessionRepository) TouchSession(ctx context.Context, p TouchSessionParams) error {
	return r.db.WithOrg(ctx, p.OrganizationID, func(tx *sql.Tx) error {
		return sqlc.New(tx).TouchSession(ctx, sqlc.TouchSessionParams{
			Now:            p.Now,
			IdleExpiresAt:  p.IdleExpiresAt,
			OrganizationID: p.OrganizationID,
			ID:             p.SessionID,
		})
	})
}

func (r *sessionRepository) RevokeSession(ctx context.Context, p RevokeSessionParams) error {
	return r.db.WithOrg(ctx, p.OrganizationID, func(tx *sql.Tx) error {
		return sqlc.New(tx).RevokeSession(ctx, sqlc.RevokeSessionParams{
			Now:            p.Now,
			OrganizationID: p.OrganizationID,
			TokenHash:      p.TokenHash,
		})
	})
}

func (r *sessionRepository) MarkSessionMFAVerified(ctx context.Context, organizationID, sessionID string, now time.Time) error { //nolint:lll
	return r.db.WithOrg(ctx, organizationID, func(tx *sql.Tx) error {
		_, err := sqlc.New(tx).MarkSessionMFAVerified(ctx, sqlc.MarkSessionMFAVerifiedParams{
			OrganizationID: organizationID,
			ID:             sessionID,
			Now:            now,
		})
		return err
	})
}

func (r *sessionRepository) SetSessionStepUp(ctx context.Context, organizationID, sessionID string, until, now time.Time) error { //nolint:lll
	return r.db.WithOrg(ctx, organizationID, func(tx *sql.Tx) error {
		_, err := sqlc.New(tx).SetSessionStepUp(ctx, sqlc.SetSessionStepUpParams{
			OrganizationID: organizationID,
			ID:             sessionID,
			Now:            now,
			Until:          until,
		})
		return err
	})
}
