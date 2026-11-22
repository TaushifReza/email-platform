package db

import (
	"context"
	"database/sql"
)

// WithSystem runs fn on the dedicated system connection, spanning
// organizations. The tenant role is never elevated, which keeps the
// email_app RLS policies intact.
func (d *DB) WithSystem(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := d.systemDB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback() //nolint:errcheck
		}
	}()
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}
