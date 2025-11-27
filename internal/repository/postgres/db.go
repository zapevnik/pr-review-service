package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type Config struct {
	DSN           string
	MigrationsDir string

	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type DB struct {
	log *slog.Logger
	sql *sql.DB
}

func New(logg *slog.Logger, ctx context.Context, cfg Config) (*DB, error) {
	if cfg.MaxOpenConns <= 0 {
		cfg.MaxOpenConns = 10
	}
	if cfg.MaxIdleConns <= 0 {
		cfg.MaxIdleConns = 5
	}
	if cfg.ConnMaxLifetime <= 0 {
		cfg.ConnMaxLifetime = time.Hour
	}

	logg.Info("opening database connection", "dsn", cfg.DSN)

	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		logg.Error("failed to open database", "error", err)
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	logg.Info("pinging database")
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		logg.Error("database ping failed", "error", err)
		return nil, err
	}
	logg.Info("database ping successful")

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		_ = db.Close()
		logg.Error("postgres.WithInstance failed", "error", err)
		return nil, fmt.Errorf("postgres.WithInstance: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", cfg.MigrationsDir),
		"postgres", driver,
	)
	if err != nil {
		_ = db.Close()
		logg.Error("failed to create migrate instance", "error", err)
		return nil, fmt.Errorf("migrate.NewWithDatabaseInstance: %w", err)
	}

	logg.Info("running database migrations")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		_ = db.Close()
		logg.Error("migration up failed", "error", err)
		return nil, fmt.Errorf("migrate up failed: %w", err)
	}
	logg.Info("database migrations completed")

	return &DB{log: logg, sql: db}, nil
}

func (db *DB) Close() error {
	db.log.Info("closing database connection")
	return db.sql.Close()
}
