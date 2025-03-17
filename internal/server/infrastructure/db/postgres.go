package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewDB(ctx context.Context, log *slog.Logger, dsn string) (*sql.DB, error) {
	op := "db.NewDB"

	conn, err := sql.Open("pgx", dsn)

	log.DebugContext(ctx, "connecting to database...")

	if err != nil {
		return nil, fmt.Errorf("%s unable to connect to database: %w", op, err)
	}

	log.DebugContext(ctx, "pingging the database...")
	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return conn, nil
}
