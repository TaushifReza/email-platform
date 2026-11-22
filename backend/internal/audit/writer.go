package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/platform/db/sqlc"
)

type Entry struct {
	ID             string
	OrganizationID string
	ActorType      string // user, operator, system
	ActorID        string
	ActorRole      string
	Action         string
	TargetType     string
	TargetID       string
	Outcome        string // success, failure
	SourceIP       string
	Client         string
	Details        map[string]any
	At             time.Time
}

type Recorder interface {
	Append(ctx context.Context, tx *sql.Tx, e Entry) error
}

type Writer struct{}

func NewWriter() *Writer { return &Writer{} }

func (w *Writer) Append(ctx context.Context, tx *sql.Tx, e Entry) error {
	return Append(ctx, tx, e)
}

func Append(ctx context.Context, tx *sql.Tx, e Entry) error {
	if e.ID == "" {
		id, err := newID()
		if err != nil {
			return err
		}
		e.ID = id
	}
	if e.At.IsZero() {
		e.At = time.Now().UTC()
	}
	if e.Outcome == "" {
		e.Outcome = "success"
	}
	details := e.Details
	if details == nil {
		details = map[string]any{}
	}
	b, err := json.Marshal(details)
	if err != nil {
		return err
	}
	return sqlc.New(tx).AppendAuditEntry(ctx, sqlc.AppendAuditEntryParams{
		ID:             e.ID,
		OrganizationID: nullString(e.OrganizationID),
		ActorType:      e.ActorType,
		ActorID:        nullString(e.ActorID),
		ActorRole:      nullString(e.ActorRole),
		Action:         e.Action,
		TargetType:     nullString(e.TargetType),
		TargetID:       nullString(e.TargetID),
		Outcome:        e.Outcome,
		SourceIp:       nullString(e.SourceIP),
		Client:         nullString(e.Client),
		Details:        json.RawMessage(b),
		CreatedAt:      e.At,
	})
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
