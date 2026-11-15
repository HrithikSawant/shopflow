package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	QueueName       = "orders"
	DeadLetterQueue = "orders.dlq"
)

type OrderCreatedEvent struct {
	Event     string `json:"event"`
	OrderID   int64  `json:"orderId"`
	ProductID int    `json:"productId"`
	Quantity  int    `json:"quantity"`
}

// Validate rejects events that can never be processed successfully.
func (e OrderCreatedEvent) Validate() error {
	switch {
	case e.Event != "ORDER_CREATED":
		return fmt.Errorf("unexpected event type %q", e.Event)
	case e.OrderID <= 0:
		return errors.New("orderId must be positive")
	case e.ProductID <= 0:
		return errors.New("productId must be positive")
	case e.Quantity <= 0:
		return errors.New("quantity must be positive")
	}

	return nil
}

// Consume processes messages until ctx is cancelled (returns nil) or the
// connection is lost (returns an error, so the process can be restarted).
//
// Delivery semantics (at-least-once):
//   - a message is acknowledged only AFTER the handler succeeded;
//   - a handler error requeues the message once; if it fails again it is
//     moved to the orders.dlq queue;
//   - malformed or invalid messages go straight to orders.dlq.
//
// Because redelivery is possible, handlers must be idempotent.
func Consume(ctx context.Context, rabbitURL string, handler func(OrderCreatedEvent) error) error {
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

	for _, name := range []string{QueueName, DeadLetterQueue} {
		if _, err := ch.QueueDeclare(name, true, false, false, false, nil); err != nil {
			return err
		}
	}

	if err := ch.Qos(10, 0, false); err != nil {
		return err
	}

	msgs, err := ch.Consume(QueueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	log.Println("worker waiting for messages...")

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return errors.New("rabbitmq: connection lost")
			}
			process(ctx, ch, msg, handler)
		}
	}
}

func process(ctx context.Context, ch *amqp.Channel, msg amqp.Delivery, handler func(OrderCreatedEvent) error) {
	var event OrderCreatedEvent

	if err := json.Unmarshal(msg.Body, &event); err != nil {
		deadLetter(ctx, ch, msg, "unparseable message: "+err.Error())
		return
	}

	if err := event.Validate(); err != nil {
		deadLetter(ctx, ch, msg, "invalid event: "+err.Error())
		return
	}

	if err := handler(event); err != nil {
		if !msg.Redelivered {
			log.Printf("handler failed for order %d, requeueing once: %v", event.OrderID, err)
			_ = msg.Nack(false, true)
			return
		}

		deadLetter(ctx, ch, msg, "handler failed twice: "+err.Error())
		return
	}

	_ = msg.Ack(false)
}

// deadLetter copies the message to the DLQ and only then acknowledges it. If
// the copy fails the message is requeued, so it is never silently lost.
func deadLetter(ctx context.Context, ch *amqp.Channel, msg amqp.Delivery, reason string) {
	log.Printf("dead-lettering message: %s", reason)

	err := ch.PublishWithContext(ctx, "", DeadLetterQueue, false, false, amqp.Publishing{
		ContentType:  msg.ContentType,
		DeliveryMode: amqp.Persistent,
		Headers:      amqp.Table{"x-error": reason},
		Body:         msg.Body,
	})
	if err != nil {
		log.Printf("could not dead-letter message, requeueing: %v", err)
		_ = msg.Nack(false, true)
		return
	}

	_ = msg.Ack(false)
}
