package rabbitmq_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/HrithikSawant/shopflow-worker/internal/rabbitmq"
)

func TestOrderCreatedMarshal(t *testing.T) {
	event := rabbitmq.OrderCreatedEvent{
		Event:     "ORDER_CREATED",
		OrderID:   1,
		ProductID: 2,
		Quantity:  3,
	}

	data, err := json.Marshal(event)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}