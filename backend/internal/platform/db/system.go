package db

import (
	"context"
	"database/sql"

	"github.com/TaushifReza/email-platform/backend/internal/platform/db/sqlc"
)

// WithSystem runs fn as the system role, spanning organizations.
func (d *DB) WithSystem(ctx context.Context, fn func(*sql.Tx) error) error {
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
	if err := sqlc.New(tx).SetSystemRole(ctx); err != nil {
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
