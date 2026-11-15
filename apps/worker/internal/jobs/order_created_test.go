package jobs_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/HrithikSawant/shopflow-worker/internal/jobs"
	"github.com/HrithikSawant/shopflow-worker/internal/rabbitmq"
)

func TestHandleOrderCreated(t *testing.T) {
	err := jobs.HandleOrderCreated(rabbitmq.OrderCreatedEvent{
		Event:     "ORDER_CREATED",
		OrderID:   1,
		ProductID: 1,
		Quantity:  2,
	})

	assert.NoError(t, err)
}
