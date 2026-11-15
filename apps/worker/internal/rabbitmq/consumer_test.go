package rabbitmq_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/HrithikSawant/shopflow-worker/internal/rabbitmq"
)

// Integration test: runs when TEST_RABBITMQ_URL is set (`make test-integration`).
func TestConsumeAckRetryAndDeadLetter(t *testing.T) {
	url := os.Getenv("TEST_RABBITMQ_URL")
	if url == "" {
		t.Skip("TEST_RABBITMQ_URL not set")
	}

	requireDedicatedVhost(t, url)

	conn, err := amqp.Dial(url)
	require.NoError(t, err)
	defer conn.Close()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	for _, q := range []string{rabbitmq.QueueName, rabbitmq.DeadLetterQueue} {
		_, err = ch.QueueDeclare(q, true, false, false, false, nil)
		require.NoError(t, err)
		_, err = ch.QueuePurge(q, false)
		require.NoError(t, err)
	}

	publish := func(body string) {
		require.NoError(t, ch.PublishWithContext(context.Background(), "", rabbitmq.QueueName, false, false,
			amqp.Publishing{ContentType: "application/json", DeliveryMode: amqp.Persistent, Body: []byte(body)}))
	}

	publish(`not json`)                                                         // -> DLQ
	publish(`{"event":"ORDER_CREATED","orderId":0,"productId":1,"quantity":1}`) // invalid -> DLQ
	publish(`{"event":"ORDER_CREATED","orderId":1,"productId":1,"quantity":1}`) // ok
	publish(`{"event":"ORDER_CREATED","orderId":2,"productId":1,"quantity":1}`) // always fails -> retried once -> DLQ

	var (
		mu    sync.Mutex
		calls = map[int64]int{}
	)

	handler := func(e rabbitmq.OrderCreatedEvent) error {
		mu.Lock()
		defer mu.Unlock()

		calls[e.OrderID]++

		if e.OrderID == 2 {
			return errors.New("boom")
		}

		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)

	go func() { done <- rabbitmq.Consume(ctx, url, handler) }()

	require.Eventually(t, func() bool {
		q, err := ch.QueueInspect(rabbitmq.DeadLetterQueue)
		return err == nil && q.Messages == 3
	}, 10*time.Second, 100*time.Millisecond, "3 messages should end up in the DLQ")

	cancel()
	require.NoError(t, <-done)

	mu.Lock()
	defer mu.Unlock()

	assert.Equal(t, 1, calls[1], "good message handled exactly once")
	assert.Equal(t, 2, calls[2], "failing message tried twice, then dead-lettered")
	assert.Zero(t, calls[0], "invalid message never reaches the handler")

	q, err := ch.QueueInspect(rabbitmq.QueueName)
	require.NoError(t, err)
	assert.Equal(t, 0, q.Messages, "main queue is drained")
}

// requireDedicatedVhost protects real queues: these tests purge and consume
// the "orders" queues, so they must never run on the default vhost where a
// live API / worker may be using them.
func requireDedicatedVhost(t *testing.T, url string) {
	t.Helper()

	uri, err := amqp.ParseURI(url)
	require.NoError(t, err)

	if uri.Vhost == "/" {
		t.Fatal("refusing to purge queues on the default vhost; use a dedicated test vhost (make test-integration sets one up)")
	}
}
