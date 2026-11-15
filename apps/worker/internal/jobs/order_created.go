package jobs

import (
	"log"

	"github.com/HrithikSawant/shopflow-worker/internal/rabbitmq"
)

// HandleOrderCreated must be idempotent: the same event can be delivered
// more than once (at-least-once delivery).
func HandleOrderCreated(event rabbitmq.OrderCreatedEvent) error {
	log.Printf("received event: %s", event.Event)

	log.Printf("order %d: generating invoice", event.OrderID)

	// Stock was already reserved by the API when the order was created.
	log.Printf("order %d: %d x product %d reserved", event.OrderID, event.Quantity, event.ProductID)

	log.Printf("order %d: sending confirmation email", event.OrderID)

	return nil
}
