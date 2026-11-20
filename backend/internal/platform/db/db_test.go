package db

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
)

func TestNewPool(t *testing.T) {
	ctx := context.Background()
	pool, err := New(ctx, testDSN())
	require.NoError(t, err)
	require.NotNil(t, pool)
	defer pool.Close()

	require.NoError(t, pool.Ping(ctx))
}

func TestMigrateIsIdempotent(t *testing.T) {
	ctx := context.Background()

	sqlDB, err := sql.Open("pgx", testDSN())
	require.NoError(t, err)
	defer sqlDB.Close()

	require.NoError(t, Migrate(ctx, sqlDB))
	require.NoError(t, Migrate(ctx, sqlDB))

	var version int64
	err = sqlDB.QueryRowContext(ctx, "SELECT COALESCE(MAX(version_id), 0) FROM goose_db_version").Scan(&version)
	require.NoError(t, err)
	require.GreaterOrEqual(t, version, int64(1))
}

func testDSN() string {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}
	return "postgres://postgres:password@localhost:5432/email_platform_dev?sslmode=disable"
}
