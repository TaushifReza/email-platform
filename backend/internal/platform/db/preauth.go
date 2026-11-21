package db

import (
	"context"
	"database/sql"

	"github.com/TaushifReza/email-platform/backend/internal/platform/db/sqlc"
)

// WithPreAuth runs fn as email_app without app.org_id. It is only for the
// pre-authentication credential lookup; tenant queries here fail closed.
func (d *DB) WithPreAuth(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := d.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback() //nolint:errcheck
		}
	}()

	if err := sqlc.New(tx).SetAppRole(ctx); err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}
