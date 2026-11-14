package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func CreateOrder(db *pgx.Conn, totalAmount float64) (int64, error) {
	var orderID int64

	err := db.QueryRow(
		context.Background(),
		`
		INSERT INTO orders (status, total_amount)
		VALUES ('PENDING', $1)
		RETURNING id
		`,
		totalAmount,
	).Scan(&orderID)

	return orderID, err
}
