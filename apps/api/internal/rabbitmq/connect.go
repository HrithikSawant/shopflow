package rabbitmq

import (
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Connect(rabbitURL string) (*amqp.Connection, error) {
	var (
		conn *amqp.Connection
		err  error
	)

	for i := 1; i <= 30; i++ {
		conn, err = amqp.Dial(rabbitURL)

		if err == nil {
			log.Println("rabbitmq connected")
			return conn, nil
		}

		log.Printf(
			"waiting for rabbitmq (%d/30): %v",
			i,
			err,
		)

		time.Sleep(time.Second)
	}

	return nil, err
}
