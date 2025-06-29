package database

import (
	"context"
	"database/sql"
	"saga-pattern/internal/database/models"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

// NewMockDatabase creates an in-memory SQLite Bun DB for tests
func NewMockDatabase(t *testing.T) *bun.DB {
	t.Helper()

	sqldb, err := sql.Open(sqliteshim.ShimName, "file::memory:?cache=shared")
	if err != nil {
		panic(err)
	}

	require.NoError(t, err)

	db := bun.NewDB(sqldb, sqlitedialect.New())

	// Run auto-migrations
	_, err = db.NewCreateTable().
		Model((*models.Order)(nil)).
		IfNotExists().
		Exec(context.Background())
	require.NoError(t, err)

	return db
}
