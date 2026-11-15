// Package outbox publishes events stored in the outbox_events table.
//
// Orders and their events are written in one database statement, so an order
// can never exist without its event. The relay then delivers events to
// RabbitMQ with at-least-once semantics: if the process dies between publish
// and commit, the event is published again, so consumers must be idempotent.
package outbox

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Publisher interface {
	Publish(ctx context.Context, body []byte) error
}

const batchSize = 50

// Run polls the outbox until ctx is cancelled.
func Run(ctx context.Context, pool *pgxpool.Pool, pub Publisher, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for {
				n, err := Drain(ctx, pool, pub, batchSize)
				if err != nil {
					log.Printf("outbox relay: %v", err)
					break
				}
				if n < batchSize {
					break
				}
			}
		}
	}
}

// Drain publishes up to limit unpublished events, in order, and marks the
// successfully published ones. It returns how many were published. On a
// publish error it keeps the progress made so far and returns the error.
func Drain(ctx context.Context, pool *pgxpool.Pool, pub Publisher, limit int) (int, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
        SELECT id, payload
        FROM outbox_events
        WHERE published_at IS NULL
        ORDER BY id
        LIMIT $1
        FOR UPDATE SKIP LOCKED`, limit)
	if err != nil {
		return 0, err
	}

	type event struct {
		id      int64
		payload []byte
	}

	var events []event

	for rows.Next() {
		var e event
		if err := rows.Scan(&e.id, &e.payload); err != nil {
			rows.Close()
			return 0, err
		}
		events = append(events, e)
	}
	rows.Close()

	if err := rows.Err(); err != nil {
		return 0, err
	}

	var (
		done   []int64
		pubErr error
	)

	for _, e := range events {
		pctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		err := pub.Publish(pctx, e.payload)
		cancel()

		if err != nil {
			pubErr = err
			break
		}

		done = append(done, e.id)
	}

	if len(done) > 0 {
		if _, err := tx.Exec(ctx,
			`UPDATE outbox_events SET published_at = now() WHERE id = ANY($1)`, done); err != nil {
			return 0, err
		}
		if err := tx.Commit(ctx); err != nil {
			return 0, err
		}
	}

	return len(done), pubErr
}
