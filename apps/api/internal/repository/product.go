package repository

import (
	"context"

	"github.com/HrithikSawant/shopflow-api/internal/models"
	"github.com/jackc/pgx/v5"
)

func GetProductByID(
	db *pgx.Conn,
	productID int,
) (*models.Product, error) {

	var p models.Product

	err := db.QueryRow(
		context.Background(),
		`
		SELECT
			id,
			name,
			description,
			price,
			inventory_count
		FROM products
		WHERE id = $1
		`,
		productID,
	).Scan(
		&p.ID,
		&p.Name,
		&p.Description,
		&p.Price,
		&p.Inventory,
	)

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func ListProducts(db *pgx.Conn) ([]models.Product, error) {
	rows, err := db.Query(
		context.Background(),
		`SELECT id,name,description,price,inventory_count FROM products`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	products := make([]models.Product, 0)

	for rows.Next() {
		var p models.Product

		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Price,
			&p.Inventory,
		)

		if err != nil {
			return nil, err
		}

		products = append(products, p)
	}

	return products, nil
}
