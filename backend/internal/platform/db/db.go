package db

import (
	"context"
	"database/sql"
	"time"

	// Register the pgx driver for database/sql.
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB owns two database handles: the tenant application role and a dedicated
// system-role connection for cross-organization work. Query methods are
// intentionally not exposed: tenant SQL must run inside WithOrg so row-level
// security is always scoped.
type DB struct {
	sql    *sql.DB
	system *sql.DB
}

func openPool(ctx context.Context, dsn string) (*sql.DB, error) {
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(30 * time.Second)

	if err = sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return sqlDB, nil
}

// New creates a handle whose system work uses the same DSN. Callers with a
// separate system role should use NewWithSystem.
func New(ctx context.Context, dsn string) (*DB, error) {
	return NewWithSystem(ctx, dsn, dsn)
}

// NewWithSystem creates tenant and system handles. The system DSN must use a
// role with the operator/system grants; it must never be the tenant role.
func NewWithSystem(ctx context.Context, appDSN, systemDSN string) (*DB, error) {
	if systemDSN == "" {
		systemDSN = appDSN
	}
	app, err := openPool(ctx, appDSN)
	if err != nil {
		return nil, err
	}
	if systemDSN == appDSN {
		return &DB{sql: app, system: app}, nil
	}
	system, err := openPool(ctx, systemDSN)
	if err != nil {
		_ = app.Close()
		return nil, err
	}
	return &DB{sql: app, system: system}, nil
}

func (d *DB) systemDB() *sql.DB {
	if d.system != nil {
		return d.system
	}
	return d.sql
}

// Ping verifies the database is reachable.
func (d *DB) Ping(ctx context.Context) error {
	return d.sql.PingContext(ctx)
}

// Close releases the database handles.
func (d *DB) Close() error {
	if d.system != nil && d.system != d.sql {
		_ = d.system.Close()
	}
	return d.sql.Close()
}
