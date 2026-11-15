package rabbitmq_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/HrithikSawant/shopflow-worker/internal/rabbitmq"
)

func TestOrderCreatedJSONContract(t *testing.T) {
	// This is exactly what the API's outbox writes. If it changes, the worker must change too.
	payload := `{"event":"ORDER_CREATED","orderId":7,"productId":2,"quantity":3}`

	var e rabbitmq.OrderCreatedEvent
	assert.NoError(t, json.Unmarshal([]byte(payload), &e))

	assert.Equal(t, rabbitmq.OrderCreatedEvent{
		Event: "ORDER_CREATED", OrderID: 7, ProductID: 2, Quantity: 3,
	}, e)
	assert.NoError(t, e.Validate())
}

func TestValidate(t *testing.T) {
	ok := rabbitmq.OrderCreatedEvent{Event: "ORDER_CREATED", OrderID: 1, ProductID: 1, Quantity: 1}

	assert.NoError(t, ok.Validate())

	bad := map[string]rabbitmq.OrderCreatedEvent{
		"wrong event":   {Event: "X", OrderID: 1, ProductID: 1, Quantity: 1},
		"no order id":   {Event: "ORDER_CREATED", ProductID: 1, Quantity: 1},
		"no product id": {Event: "ORDER_CREATED", OrderID: 1, Quantity: 1},
		"zero quantity": {Event: "ORDER_CREATED", OrderID: 1, ProductID: 1},
		"empty":         {},
	}

	for name, e := range bad {
		t.Run(name, func(t *testing.T) { assert.Error(t, e.Validate()) })
	}
}
