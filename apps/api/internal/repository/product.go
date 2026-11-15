package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/HrithikSawant/shopflow-api/internal/models"
)

func ListProducts(db *pgxpool.Pool) ([]models.Product, error) {
	rows, err := db.Query(
		context.Background(),
		`SELECT id,name,description,price,inventory_count FROM products ORDER BY id`,
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

	return products, rows.Err()
}
