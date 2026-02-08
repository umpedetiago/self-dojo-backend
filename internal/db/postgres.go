package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("db connect: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db ping: %w", err)
	}

	return pool, nil
}

// ConnectWithRetry tenta conectar ao banco com retries (útil quando o Postgres ainda está subindo).
func ConnectWithRetry(ctx context.Context, databaseURL string, maxAttempts int, interval time.Duration) (*pgxpool.Pool, error) {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		pool, err := Connect(ctx, databaseURL)
		if err == nil {
			return pool, nil
		}
		lastErr = err
		if attempt < maxAttempts {
			log.Printf("db: tentativa %d/%d falhou, aguardando %v: %v", attempt, maxAttempts, interval, err)
			time.Sleep(interval)
		}
	}
	return nil, fmt.Errorf("db connect after %d attempts: %w", maxAttempts, lastErr)
}

func EnsureSchema(ctx context.Context, pool *pgxpool.Pool) error {
	var dataType string
	err := pool.QueryRow(ctx, `
SELECT data_type FROM information_schema.columns
WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'id'
`).Scan(&dataType)
	if err == nil && (dataType == "bigint" || dataType == "integer") {
		log.Printf("db: migrando users (id %s -> UUID), recriando tabela", dataType)
		_, _ = pool.Exec(ctx, `DROP TABLE IF EXISTS users`)
	}

	_, err = pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`)
	if err != nil {
		return fmt.Errorf("ensure schema: %w", err)
	}
	return nil
}
