package outbox_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/HrithikSawant/shopflow-api/internal/outbox"
	"github.com/HrithikSawant/shopflow-api/internal/repository"
)

type fakePublisher struct {
	mu     sync.Mutex
	fail   bool
	bodies []string
}

func (f *fakePublisher) Publish(_ context.Context, body []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.fail {
		return errors.New("broker down")
	}

	f.bodies = append(f.bodies, string(body))

	return nil
}

func setup(t *testing.T) (*pgxpool.Pool, int64) {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	pool, err := pgxpool.New(context.Background(), url)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	var productID int
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO products (name, price, inventory_count) VALUES ($1, 10, 10) RETURNING id`,
		fmt.Sprintf("outbox-test-%d", time.Now().UnixNano())).Scan(&productID))

	orderID, err := repository.CreateOrder(pool, productID, 1)
	require.NoError(t, err)

	return pool, orderID
}

func publishedAt(t *testing.T, pool *pgxpool.Pool, orderID int64) bool {
	t.Helper()

	var published bool
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT published_at IS NOT NULL FROM outbox_events WHERE order_id=$1`,
		orderID).Scan(&published))

	return published
}

func TestDrainPublishesAndMarksEvents(t *testing.T) {
	pool, orderID := setup(t)
	pub := &fakePublisher{}

	_, err := outbox.Drain(context.Background(), pool, pub, 1000)
	require.NoError(t, err)

	assert.True(t, publishedAt(t, pool, orderID))
	assert.Contains(t, fmt.Sprint(pub.bodies), fmt.Sprintf(`"orderId": %d`, orderID))

	// A second drain must not publish the same event again.
	before := len(pub.bodies)
	_, err = outbox.Drain(context.Background(), pool, pub, 1000)
	require.NoError(t, err)
	assert.Equal(t, before, len(pub.bodies))
}

func TestDrainKeepsEventsWhenBrokerIsDown(t *testing.T) {
	pool, orderID := setup(t)

	_, err := outbox.Drain(context.Background(), pool, &fakePublisher{fail: true}, 1000)

	assert.Error(t, err)
	assert.False(t, publishedAt(t, pool, orderID), "event must stay pending and be retried")

	// Broker is back: the pending event is delivered.
	_, err = outbox.Drain(context.Background(), pool, &fakePublisher{}, 1000)
	require.NoError(t, err)
	assert.True(t, publishedAt(t, pool, orderID))
}
