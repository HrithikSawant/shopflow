package database

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect opens a connection pool and waits (up to ~30s) for Postgres to accept connections.
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	var lastErr error

	for i := 1; i <= 30; i++ {
		pool, err := pgxpool.New(ctx, databaseURL)
		if err == nil {
			if err = pool.Ping(ctx); err == nil {
				log.Println("postgres connected")
				return pool, nil
			}
			pool.Close()
		}

		lastErr = err
		log.Printf("waiting for postgres (%d/30): %v", i, err)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}

	return nil, lastErr
}
