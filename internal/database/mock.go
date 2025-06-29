package database

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

// NewMockDatabase creates an in-memory SQLite Bun DB for tests
func NewMockDatabase(t *testing.T, models ...interface{}) *bun.DB {
	t.Helper()

	sqldb, err := sql.Open(sqliteshim.ShimName, ":memory:")
	require.NoError(t, err)

	db := bun.NewDB(sqldb, sqlitedialect.New())

	for _, m := range models {
		_, err := db.
			NewCreateTable().
			Model(m).
			IfNotExists().
			Exec(context.Background())
		require.NoError(t, err)
	}

	t.Cleanup(func() { _ = db.Close() })

	return db
}
