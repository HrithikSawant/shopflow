package database

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

func Connect(databaseURL string) (*pgx.Conn, error) {
	var conn *pgx.Conn
	var err error

	for i := 1; i <= 30; i++ {
		conn, err = pgx.Connect(
			context.Background(),
			databaseURL,
		)

		if err == nil {
			log.Println("postgres connected")
			return conn, nil
		}

		log.Printf(
			"waiting for postgres (%d/30): %v",
			i,
			err,
		)

		time.Sleep(time.Second)
	}

	return nil, err
}
