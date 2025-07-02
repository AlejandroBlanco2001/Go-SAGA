package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"saga-pattern/internal/database/models"

	"github.com/joho/godotenv"
)

// Hardcoded DB config — you can replace this with env vars or config file later
const (
	port     = 5432
	user     = "myuser"
	password = "somerandompassword"
)

type DBConfig struct {
	host     string
	port     int
	user     string
	password string
	name     string
}

func (c *DBConfig) getDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", c.user, c.password, c.host, c.port, c.name)
}

// NewDatabase creates and returns a *bun.DB instance
func NewDatabase(log *zap.Logger) (*bun.DB, error) {
	err := godotenv.Load()

	if err != nil {
		log.Error("Error loading .env file", zap.Error(err))
	}

	database := os.Getenv("DATABASE_NAME")
	host := os.Getenv("HOST")

	cfg := DBConfig{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		name:     database,
	}

	dsn := cfg.getDSN()

	log.Info("Connecting to DB", zap.String("dsn", dsn))
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	maxAttempts := 10

	// change this to a exponential backoff
	for i := 0; i < maxAttempts; i++ {
		err := sqldb.Ping()
		if err == nil {
			log.Info("Succesfully connected and pinged database")
			break
		}

		log.Warn("Failed to ping database, retrying...", zap.Error(err), zap.Int("attempt", i+1))
		time.Sleep(2 * time.Second)
		if i == maxAttempts-1 {
			return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxAttempts, err)
		}
	}

	db := bun.NewDB(sqldb, pgdialect.New())

	return db, nil
}

func runMigrations(ctx context.Context, db *bun.DB, log *zap.Logger) error {
	log.Info("Running migrations")

	_, err := db.NewCreateTable().Model((*models.Order)(nil)).IfNotExists().Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create Orders table: %w", err)
	}

	_, err = db.NewCreateTable().Model((*models.Inventory)(nil)).IfNotExists().Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create Inventory table: %w", err)
	}

	log.Info("Migrations completed")

	return nil
}

// Module provides the *bun.DB instance for use in other fx components
var Module = fx.Module("database",
	fx.Provide(NewDatabase),
	fx.Invoke(func(db *bun.DB, log *zap.Logger) error {
		return runMigrations(context.Background(), db, log)
	}),
)
