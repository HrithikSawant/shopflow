package jobs_test

import (
	"testing"

	"github.com/HrithikSawant/shopflow-worker/internal/jobs"
	"github.com/HrithikSawant/shopflow-worker/internal/rabbitmq"
)

func TestHandleOrderCreated(t *testing.T) {

	event := rabbitmq.OrderCreatedEvent{
		Event:     "ORDER_CREATED",
		OrderID:   1,
		ProductID: 1,
		Quantity:  2,
	}

	jobs.HandleOrderCreated(event)
}