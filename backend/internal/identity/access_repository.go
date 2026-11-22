package identity

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/platform/db"
	"github.com/TaushifReza/email-platform/backend/internal/platform/db/sqlc"
	"github.com/TaushifReza/email-platform/backend/internal/platform/outbox"
)

const (
	invitationExpiry = 72 * time.Hour
	resetExpiry      = 30 * time.Minute
	nullJSON         = "null"
)

type AccessRepository interface {
	IssueInvitation(ctx context.Context, orgID, invitedBy, email, role string, now time.Time,
		expiresAt time.Time, tokenHash []byte, invitationID string, msg outbox.Message) error
	LookupInvitation(ctx context.Context, tokenHash []byte) (invitationLookup, error)
	AcceptInvitation(ctx context.Context, orgID, userID, invitationID string, passwordHash string, now time.Time) error
	LookupResetTarget(ctx context.Context, email string) (resetTarget, error)
	RequestPasswordReset(ctx context.Context, orgID, userID string, now time.Time,
		tokenID string, tokenHash []byte, expiresAt time.Time, msg outbox.Message) error
	LookupPasswordReset(ctx context.Context, tokenHash []byte) (passwordResetLookup, error)
	ResetPassword(ctx context.Context, orgID, userID, tokenID string, passwordHash string, now time.Time) error
}

type invitationLookup struct {
	InvitationID    string
	OrganizationID  string
	UserID          string
	Email           string
	Role            string
	ExpiresAt       time.Time
	AcceptedAt      *time.Time
	RevokedAt       *time.Time
	InvitedByUserID string
}

type resetTarget struct {
	UserID         string
	OrganizationID string
	Email          string
	Status         string
}

type passwordResetLookup struct {
	TokenID        string
	OrganizationID string
	UserID         string
	ExpiresAt      time.Time
	UsedAt         *time.Time
}

type accessRepository struct {
	db     *db.DB
	outbox *outbox.Store
}

func NewAccessRepository(database *db.DB, o *outbox.Store) AccessRepository {
	return &accessRepository{db: database, outbox: o}
}

func (r *accessRepository) IssueInvitation(ctx context.Context, orgID, invitedBy, email, role string, _ time.Time,
	expiresAt time.Time, tokenHash []byte, invitationID string, msg outbox.Message,
) error {
	return r.db.WithOrg(ctx, orgID, func(tx *sql.Tx) error {
		if err := sqlc.New(tx).CreateInvitation(ctx, sqlc.CreateInvitationParams{
			ID:              invitationID,
			OrganizationID:  orgID,
			Email:           email,
			Role:            role,
			InvitedByUserID: invitedBy,
			TokenHash:       tokenHash,
			ExpiresAt:       expiresAt,
		}); err != nil {
			return err
		}
		return r.outbox.EnqueueTx(ctx, tx, msg)
	})
}

func (r *accessRepository) LookupInvitation(ctx context.Context, tokenHash []byte) (invitationLookup, error) {
	var out invitationLookup
	err := r.db.WithPreAuth(ctx, func(tx *sql.Tx) error {
		raw, err := sqlc.New(tx).LookupInvitationByToken(ctx, tokenHash)
		if err != nil {
			return err
		}
		if len(raw) == 0 || string(raw) == nullJSON {
			return ErrUserNotFound
		}
		var p struct {
			InvitationID    string     `json:"invitation_id"`
			OrganizationID  string     `json:"organization_id"`
			UserID          *string    `json:"user_id"`
			Email           string     `json:"email"`
			Role            string     `json:"role"`
			ExpiresAt       time.Time  `json:"expires_at"`
			AcceptedAt      *time.Time `json:"accepted_at"`
			RevokedAt       *time.Time `json:"revoked_at"`
			InvitedByUserID string     `json:"invited_by_user_id"`
		}
		if err := json.Unmarshal(raw, &p); err != nil {
			return err
		}
		out = invitationLookup{
			InvitationID:    p.InvitationID,
			OrganizationID:  p.OrganizationID,
			Email:           p.Email,
			Role:            p.Role,
			ExpiresAt:       p.ExpiresAt,
			AcceptedAt:      p.AcceptedAt,
			RevokedAt:       p.RevokedAt,
			InvitedByUserID: p.InvitedByUserID,
		}
		if p.UserID != nil {
			out.UserID = *p.UserID
		}
		return nil
	})
	if err != nil {
		return invitationLookup{}, err
	}
	return out, nil
}

func (r *accessRepository) AcceptInvitation(ctx context.Context, orgID, userID, invitationID string,
	passwordHash string, now time.Time,
) error {
	return r.db.WithOrg(ctx, orgID, func(tx *sql.Tx) error {
		rows, err := sqlc.New(tx).MarkInvitationAccepted(ctx, sqlc.MarkInvitationAcceptedParams{
			OrganizationID: orgID,
			UserID:         userID,
			AcceptedAt:     now,
			ID:             invitationID,
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return errors.New("identity: invitation no longer valid")
		}
		// User must have been created at invitation time; if not, nothing to activate.
		_, err = sqlc.New(tx).ActivateInvitedUser(ctx, sqlc.ActivateInvitedUserParams{
			OrganizationID:    orgID,
			ID:                userID,
			PasswordHash:      passwordHash,
			PasswordChangedAt: now,
		})
		if err != nil {
			return err
		}
		return nil
	})
}

func (r *accessRepository) LookupResetTarget(ctx context.Context, email string) (resetTarget, error) {
	var out resetTarget
	err := r.db.WithPreAuth(ctx, func(tx *sql.Tx) error {
		raw, err := sqlc.New(tx).LookupResetTarget(ctx, email)
		if err != nil {
			return err
		}
		if len(raw) == 0 || string(raw) == nullJSON {
			return ErrUserNotFound
		}
		var p struct {
			UserID         string `json:"user_id"`
			OrganizationID string `json:"organization_id"`
			Email          string `json:"email"`
			Status         string `json:"status"`
		}
		if err := json.Unmarshal(raw, &p); err != nil {
			return err
		}
		out = resetTarget{UserID: p.UserID, OrganizationID: p.OrganizationID, Email: p.Email, Status: p.Status}
		return nil
	})
	if err != nil {
		return resetTarget{}, err
	}
	return out, nil
}

func (r *accessRepository) RequestPasswordReset(ctx context.Context, orgID, userID string, _ time.Time,
	tokenID string, tokenHash []byte, expiresAt time.Time, msg outbox.Message,
) error {
	return r.db.WithOrg(ctx, orgID, func(tx *sql.Tx) error {
		if err := sqlc.New(tx).CreateAuthToken(ctx, sqlc.CreateAuthTokenParams{
			ID:             tokenID,
			OrganizationID: orgID,
			UserID:         userID,
			TokenHash:      tokenHash,
			ExpiresAt:      expiresAt,
		}); err != nil {
			return err
		}
		return r.outbox.EnqueueTx(ctx, tx, msg)
	})
}

func (r *accessRepository) LookupPasswordReset(ctx context.Context, tokenHash []byte) (passwordResetLookup, error) {
	var out passwordResetLookup
	err := r.db.WithPreAuth(ctx, func(tx *sql.Tx) error {
		raw, err := sqlc.New(tx).LookupPasswordResetByToken(ctx, tokenHash)
		if err != nil {
			return err
		}
		if len(raw) == 0 || string(raw) == nullJSON {
			return ErrUserNotFound
		}
		var p struct {
			TokenID        string     `json:"token_id"`
			OrganizationID string     `json:"organization_id"`
			UserID         string     `json:"user_id"`
			ExpiresAt      time.Time  `json:"expires_at"`
			UsedAt         *time.Time `json:"used_at"`
		}
		if err := json.Unmarshal(raw, &p); err != nil {
			return err
		}
		out = passwordResetLookup{
			TokenID:        p.TokenID,
			OrganizationID: p.OrganizationID,
			UserID:         p.UserID,
			ExpiresAt:      p.ExpiresAt,
			UsedAt:         p.UsedAt,
		}
		return nil
	})
	if err != nil {
		return passwordResetLookup{}, err
	}
	return out, nil
}

func (r *accessRepository) ResetPassword(ctx context.Context, orgID, userID, tokenID string,
	passwordHash string, now time.Time,
) error {
	return r.db.WithOrg(ctx, orgID, func(tx *sql.Tx) error {
		rows, err := sqlc.New(tx).ConsumePasswordResetToken(ctx, sqlc.ConsumePasswordResetTokenParams{
			OrganizationID: orgID,
			ID:             tokenID,
			UsedAt:         now,
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return errors.New("identity: reset token no longer valid")
		}
		rows, err = sqlc.New(tx).SetUserPassword(ctx, sqlc.SetUserPasswordParams{
			OrganizationID:    orgID,
			ID:                userID,
			PasswordHash:      passwordHash,
			PasswordChangedAt: now,
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return ErrUserNotFound
		}
		return sqlc.New(tx).RevokeAllUserSessions(ctx, sqlc.RevokeAllUserSessionsParams{
			OrganizationID: orgID,
			UserID:         userID,
			RevokedAt:      now,
		})
	})
}

// AccessService orchestrates invitations and password reset.
