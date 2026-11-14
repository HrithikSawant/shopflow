package rabbitmq

import (
	"encoding/json"
	"log"
)

type OrderCreatedEvent struct {
	Event     string `json:"event"`
	OrderID   int64  `json:"orderId"`
	ProductID int    `json:"productId"`
	Quantity  int    `json:"quantity"`
}

func Consume(rabbitURL string, handler func(OrderCreatedEvent)) error {
	conn, err := Connect(rabbitURL)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	q, err := ch.QueueDeclare(
		"orders",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	log.Println("worker waiting for messages...")

	go func() {
		for msg := range msgs {
			var event OrderCreatedEvent

			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("failed to parse message: %v", err)
				continue
			}

			handler(event)
		}
	}()

	select {}
}
