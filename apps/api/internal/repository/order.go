package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

// createOrderSQL runs as ONE atomic statement:
//   - reserves stock (the UPDATE only matches when enough stock is left, and
//     row locking makes concurrent orders safe from overselling),
//   - inserts the order with total = unit price * quantity (computed by
//     Postgres NUMERIC, no float rounding),
//   - inserts the order item with the UNIT price,
//   - writes an outbox event, published to RabbitMQ later by the relay.
//
// If the product is missing or stock is too low, nothing is written and no
// row is returned.
const createOrderSQL = `
WITH reserved AS (
    UPDATE products
    SET inventory_count = inventory_count - $2::int
    WHERE id = $1::bigint AND inventory_count >= $2::int
    RETURNING id, price
), new_order AS (
    INSERT INTO orders (status, total_amount)
    SELECT 'PENDING', price * $2::int FROM reserved
    RETURNING id
), new_item AS (
    INSERT INTO order_items (order_id, product_id, quantity, price)
    SELECT new_order.id, reserved.id, $2::int, reserved.price
    FROM new_order, reserved
), new_event AS (
    INSERT INTO outbox_events (order_id, payload)
    SELECT new_order.id, jsonb_build_object(
        'event', 'ORDER_CREATED',
        'orderId', new_order.id,
        'productId', reserved.id,
        'quantity', $2::int)
    FROM new_order, reserved
)
SELECT id FROM new_order`

// CreateOrder creates an order for one product and returns its id.
// It returns ErrProductNotFound or ErrInsufficientStock for expected failures.
func CreateOrder(db *pgxpool.Pool, productID, quantity int) (int64, error) {
	ctx := context.Background()

	var orderID int64

	err := db.QueryRow(ctx, createOrderSQL, productID, quantity).Scan(&orderID)
	if err == nil {
		return orderID, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	var exists bool

	if err := db.QueryRow(
		ctx,
		`SELECT EXISTS (SELECT 1 FROM products WHERE id = $1)`,
		productID,
	).Scan(&exists); err != nil {
		return 0, err
	}

	if !exists {
		return 0, ErrProductNotFound
	}

	return 0, ErrInsufficientStock
}
