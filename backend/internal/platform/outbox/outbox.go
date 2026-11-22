package outbox

import (
	"context"
	"database/sql"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/platform/db"
	"github.com/TaushifReza/email-platform/backend/internal/platform/db/sqlc"
)

// Message is a queued system email.
type Message struct {
	ID            string
	OrganizationID string
	UserID        string
	Kind          string
	Recipient     string
	Subject       string
	TextBody      string
	HTMLBody      string
	NextAttemptAt time.Time
}

// Store persists system emails.
type Store struct {
	db *db.DB
}

// NewStore creates an outbox store.
func NewStore(database *db.DB) *Store {
	return &Store{db: database}
}

// EnqueueTx inserts a message inside an existing tenant transaction.
func (s *Store) EnqueueTx(ctx context.Context, tx *sql.Tx, m Message) error {
	if m.NextAttemptAt.IsZero() {
		m.NextAttemptAt = time.Now().UTC()
	}
	userID := sql.NullString{}
	if m.UserID != "" {
		userID = sql.NullString{String: m.UserID, Valid: true}
	}
	html := sql.NullString{}
	if m.HTMLBody != "" {
		html = sql.NullString{String: m.HTMLBody, Valid: true}
	}
	return sqlc.New(tx).EnqueueSystemEmail(ctx, sqlc.EnqueueSystemEmailParams{
		ID:             m.ID,
		OrganizationID: m.OrganizationID,
		UserID:         userID,
		Kind:           m.Kind,
		RecipientEmail: m.Recipient,
		Subject:        m.Subject,
		BodyText:       m.TextBody,
		BodyHtml:       html,
		NextAttemptAt:  m.NextAttemptAt,
	})
}

// Claimed is a claimed outbox row.
type Claimed struct {
	ID             string
	OrganizationID string
	UserID         string
	Kind           string
	Recipient      string
	Subject        string
	TextBody       string
	HTMLBody       string
	AttemptCount   int32
}

// Claim claims up to limit messages for sending.
func (s *Store) Claim(ctx context.Context, limit int, lease time.Time) ([]Claimed, error) {
	var out []Claimed
	err := s.db.WithSystem(ctx, func(tx *sql.Tx) error {
		rows, err := sqlc.New(tx).ClaimSystemEmails(ctx, sqlc.ClaimSystemEmailsParams{
			RowLimit:        int32(limit), // #nosec G115
			LeaseExpiresAt:  lease,
		})
		if err != nil {
			return err
		}
		for _, r := range rows {
			c := Claimed{
				ID:             r.ID,
				OrganizationID: r.OrganizationID,
				Kind:           r.Kind,
				Recipient:      r.RecipientEmail,
				Subject:        r.Subject,
				TextBody:       r.BodyText,
				AttemptCount:   r.AttemptCount,
			}
			if r.UserID.Valid {
				c.UserID = r.UserID.String
			}
			if r.BodyHtml.Valid {
				c.HTMLBody = r.BodyHtml.String
			}
			out = append(out, c)
		}
		return nil
	})
	return out, err
}

// MarkSent marks a message as sent.
func (s *Store) MarkSent(ctx context.Context, id string) error {
	return s.db.WithSystem(ctx, func(tx *sql.Tx) error {
		return sqlc.New(tx).MarkSystemEmailSent(ctx, id)
	})
}

// MarkFailed records a failure and schedules retry or marks failed.
func (s *Store) MarkFailed(ctx context.Context, id string, attempts, maxAttempts int, next time.Time, errMsg string) error {
	return s.db.WithSystem(ctx, func(tx *sql.Tx) error {
		return sqlc.New(tx).MarkSystemEmailFailed(ctx, sqlc.MarkSystemEmailFailedParams{
			ID:            id,
			AttemptCount:  int32(attempts), // #nosec G115
			MaxAttempts:   int32(maxAttempts), // #nosec G115
			NextAttemptAt: next,
			LastError:     errMsg,
		})
	})
}