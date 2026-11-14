package jobs

import (
	"log"

	"github.com/HrithikSawant/shopflow-worker/internal/rabbitmq"
)

func HandleOrderCreated(event rabbitmq.OrderCreatedEvent) {
	log.Printf("received event: %s", event.Event)

	log.Printf("generating invoice for order %d", event.OrderID)

	log.Printf("updating inventory for order %d", event.OrderID)

	log.Printf("sending confirmation email for order %d", event.OrderID)
}
