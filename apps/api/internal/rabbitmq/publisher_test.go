package rabbitmq_test

import (
	"context"
	"os"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/HrithikSawant/shopflow-api/internal/rabbitmq"
)

func TestPublishFailsFastWhenBrokerIsDown(t *testing.T) {
	p := rabbitmq.NewPublisher("amqp://guest:guest@127.0.0.1:1/")
	defer p.Close()

	start := time.Now()
	err := p.Publish(context.Background(), []byte(`{}`))

	assert.Error(t, err)
	assert.Less(t, time.Since(start), 12*time.Second)
}

// Integration test: TEST_RABBITMQ_URL=amqp://guest:guest@localhost:5672/
func TestPublishDeliversMessage(t *testing.T) {
	url := os.Getenv("TEST_RABBITMQ_URL")
	if url == "" {
		t.Skip("TEST_RABBITMQ_URL not set")
	}

	conn, err := amqp.Dial(url)
	require.NoError(t, err)
	defer conn.Close()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	_, err = ch.QueueDeclare(rabbitmq.QueueName, true, false, false, false, nil)
	require.NoError(t, err)
	_, err = ch.QueuePurge(rabbitmq.QueueName, false)
	require.NoError(t, err)

	p := rabbitmq.NewPublisher(url)
	defer p.Close()

	require.NoError(t, p.Publish(context.Background(), []byte(`{"hello":"world"}`)))

	msg, ok, err := ch.Get(rabbitmq.QueueName, true)
	require.NoError(t, err)
	require.True(t, ok, "message should be in the queue")
	assert.JSONEq(t, `{"hello":"world"}`, string(msg.Body))
	assert.Equal(t, amqp.Persistent, msg.DeliveryMode)
}
