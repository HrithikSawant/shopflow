package rabbitmq

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishOrderCreated(
	rabbitURL string,
	orderID int64,
	productID int,
	quantity int,
) error {

	conn, err := Connect(rabbitURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

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

	event := OrderCreatedEvent{
		Event:     "ORDER_CREATED",
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  quantity,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(
		context.Background(),
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}
