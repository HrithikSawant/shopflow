package repository

import (
	"context"

	"github.com/HrithikSawant/shopflow-api/internal/models"
	"github.com/jackc/pgx/v5"
)

func ListProducts(db *pgx.Conn) ([]models.Product, error) {
	rows, err := db.Query(
		context.Background(),
		`SELECT id,name,description,price,inventory_count FROM products`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var products []models.Product

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