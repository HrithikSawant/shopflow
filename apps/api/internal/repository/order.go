package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func CreateOrder(
	db *pgx.Conn,
	productID int,
	quantity int,
	totalAmount float64,
) (int64, error) {

	tx, err := db.Begin(context.Background())
	if err != nil {
		return 0, err
	}

	defer tx.Rollback(context.Background())

	var orderID int64

	err = tx.QueryRow(
		context.Background(),
		`
		INSERT INTO orders
		(status,total_amount)
		VALUES
		('PENDING',$1)
		RETURNING id
		`,
		totalAmount,
	).Scan(&orderID)

	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(
		context.Background(),
		`
		INSERT INTO order_items
		(order_id,product_id,quantity,price)
		VALUES
		($1,$2,$3,$4)
		`,
		orderID,
		productID,
		quantity,
		totalAmount,
	)

	if err != nil {
		return 0, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return 0, err
	}

	return orderID, nil
}