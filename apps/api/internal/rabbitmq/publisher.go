package rabbitmq

import (
	"context"
	"errors"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const QueueName = "orders"

// Publisher keeps ONE long-lived connection and channel (with publisher
// confirms) and transparently reconnects when the broker drops it.
type Publisher struct {
	url string

	mu   sync.Mutex
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewPublisher(url string) *Publisher {
	return &Publisher{url: url}
}

func (p *Publisher) ensure() error {
	if p.conn != nil && !p.conn.IsClosed() && p.ch != nil && !p.ch.IsClosed() {
		return nil
	}

	p.closeLocked()

	conn, err := amqp.DialConfig(p.url, amqp.Config{
		Heartbeat: 10 * time.Second,
		Locale:    "en_US",
		Dial:      amqp.DefaultDial(5 * time.Second),
	})
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}

	if _, err := ch.QueueDeclare(QueueName, true, false, false, false, nil); err != nil {
		conn.Close()
		return err
	}

	if err := ch.Confirm(false); err != nil {
		conn.Close()
		return err
	}

	p.conn, p.ch = conn, ch

	return nil
}

// Publish sends body to the orders queue and waits for the broker's confirm.
func (p *Publisher) Publish(ctx context.Context, body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := p.ensure(); err != nil {
		return err
	}

	conf, err := p.ch.PublishWithDeferredConfirmWithContext(
		ctx, "", QueueName, false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		p.closeLocked()
		return err
	}

	ok, err := conf.WaitContext(ctx)
	if err != nil {
		p.closeLocked()
		return err
	}

	if !ok {
		return errors.New("rabbitmq: broker nacked the message")
	}

	return nil
}

func (p *Publisher) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closeLocked()
}

func (p *Publisher) closeLocked() {
	if p.ch != nil {
		_ = p.ch.Close()
	}
	if p.conn != nil {
		_ = p.conn.Close()
	}
	p.ch, p.conn = nil, nil
}
