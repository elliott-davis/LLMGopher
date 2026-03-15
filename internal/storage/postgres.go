package storage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/ed007183/llmgopher/pkg/config"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// NewPostgresDB opens a connection pool to PostgreSQL using pgx via database/sql.
func NewPostgresDB(ctx context.Context, cfg config.PostgresConfig, logger *slog.Logger) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	logger.Info("connected to PostgreSQL",
		"max_open_conns", cfg.MaxOpenConns,
		"max_idle_conns", cfg.MaxIdleConns,
	)

	return db, nil
}

// Migrate applies all embedded goose migrations before startup continues.
func Migrate(ctx context.Context, db *sql.DB) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("apply goose migrations: %w", err)
	}

	return nil
}
