package repository_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/HrithikSawant/shopflow-api/internal/repository"
)

// Integration tests: need a migrated Postgres.
//   make up && make test-integration   (sets TEST_DATABASE_URL from .env)
// They are skipped when the variable is not set.

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	pool, err := pgxpool.New(context.Background(), url)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	return pool
}

func newProduct(t *testing.T, pool *pgxpool.Pool, stock int) int {
	t.Helper()

	var id int
	err := pool.QueryRow(context.Background(),
		`INSERT INTO products (name, price, inventory_count)
         VALUES ($1, 19.99, $2) RETURNING id`,
		fmt.Sprintf("test-product-%d", time.Now().UnixNano()), stock,
	).Scan(&id)
	require.NoError(t, err)

	return id
}

func inventory(t *testing.T, pool *pgxpool.Pool, productID int) int {
	t.Helper()

	var n int
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT inventory_count FROM products WHERE id=$1`, productID).Scan(&n))

	return n
}

func TestCreateOrderSuccess(t *testing.T) {
	pool := testPool(t)
	productID := newProduct(t, pool, 5)

	orderID, err := repository.CreateOrder(pool, productID, 2)
	require.NoError(t, err)
	assert.Greater(t, orderID, int64(0))

	ctx := context.Background()

	var total, unit string
	var qty int
	require.NoError(t, pool.QueryRow(ctx, `
        SELECT o.total_amount::text, i.price::text, i.quantity
        FROM orders o JOIN order_items i ON i.order_id = o.id
        WHERE o.id = $1`, orderID).Scan(&total, &unit, &qty))

	assert.Equal(t, "39.98", total, "order total = unit price * quantity")
	assert.Equal(t, "19.99", unit, "order item stores the UNIT price")
	assert.Equal(t, 2, qty)

	assert.Equal(t, 3, inventory(t, pool, productID), "stock is decremented")

	var event string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT payload->>'event' FROM outbox_events WHERE order_id = $1`, orderID).Scan(&event))
	assert.Equal(t, "ORDER_CREATED", event, "event is written to the outbox")
}

func TestCreateOrderProductNotFound(t *testing.T) {
	pool := testPool(t)

	_, err := repository.CreateOrder(pool, 2_000_000_000, 1)

	assert.ErrorIs(t, err, repository.ErrProductNotFound)
}

func TestCreateOrderInsufficientStock(t *testing.T) {
	pool := testPool(t)
	productID := newProduct(t, pool, 1)

	_, err := repository.CreateOrder(pool, productID, 2)

	assert.ErrorIs(t, err, repository.ErrInsufficientStock)
	assert.Equal(t, 1, inventory(t, pool, productID), "stock untouched")

	var orders int
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT count(*) FROM order_items WHERE product_id=$1`, productID).Scan(&orders))
	assert.Equal(t, 0, orders, "no order was created")
}

func TestCreateOrderNeverOversells(t *testing.T) {
	pool := testPool(t)
	productID := newProduct(t, pool, 5)

	const attempts = 20

	var (
		wg         sync.WaitGroup
		mu         sync.Mutex
		succeeded  int
		outOfStock int
	)

	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, err := repository.CreateOrder(pool, productID, 1)

			mu.Lock()
			defer mu.Unlock()

			switch err {
			case nil:
				succeeded++
			case repository.ErrInsufficientStock:
				outOfStock++
			default:
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}

	wg.Wait()

	assert.Equal(t, 5, succeeded)
	assert.Equal(t, attempts-5, outOfStock)
	assert.Equal(t, 0, inventory(t, pool, productID))
}

func TestListProductsToleratesMissingDescription(t *testing.T) {
	pool := testPool(t)

	// description is a nullable column; a product without one must not break the listing
	productID := newProduct(t, pool, 1)

	products, err := repository.ListProducts(pool)
	require.NoError(t, err)

	found := false

	for _, p := range products {
		if p.ID == int64(productID) {
			found = true
			assert.Equal(t, "", p.Description)
		}
	}

	assert.True(t, found)
}
