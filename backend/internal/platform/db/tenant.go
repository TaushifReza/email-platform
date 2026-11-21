package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/TaushifReza/email-platform/backend/internal/platform/db/sqlc"
)

// ErrNoOrganization is returned when a tenant transaction is requested
// without an organization. This fails closed rather than leaving app.org_id
// unset.
var ErrNoOrganization = errors.New("db: organization id is required")

// WithOrg runs fn inside a transaction scoped to orgID. It switches to the
// tenant application role and sets app.org_id for the transaction, so every
// query in fn is constrained by row-level security.
//
// Calls must not be nested; fn receives the transaction and must use it for
// all database work.
func (d *DB) WithOrg(ctx context.Context, orgID string, fn func(*sql.Tx) error) error {
	if orgID == "" {
		return ErrNoOrganization
	}

	tx, err := d.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			// A rollback after a failed commit may report ErrTxDone.
			_ = tx.Rollback() //nolint:errcheck // Best-effort rollback.
		}
	}()

	if err := sqlc.New(tx).SetAppRole(ctx); err != nil {
		return err
	}
	if err := sqlc.New(tx).SetOrgID(ctx, orgID); err != nil {
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
