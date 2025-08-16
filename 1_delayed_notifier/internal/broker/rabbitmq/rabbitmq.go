package rabbitmq

import (
	"context"
	"delayed-notifier/internal/models"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

const (
	exchangeDirect  = "delayed_notifier.direct"
	exchangeDelayed = "delayed_notifier.delayed"
)

// RabbitMQ implements the Broker interface and can register delayed notifications
type RabbitMQ struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	log  *zerolog.Logger
}

// New creates a RabbitMQ instance.
func New(url string, logger *zerolog.Logger) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	mq := &RabbitMQ{conn: conn, ch: ch, log: logger}
	if err := mq.setup(); err != nil {
		_ = conn.Close()
		_ = ch.Close()
		return nil, err
	}
	return mq, nil
}

func (r *RabbitMQ) setup() error {
	if err := r.ch.ExchangeDeclare(
		exchangeDelayed,
		"x-delayed-message",
		true, false, false, false,
		amqp.Table{
			"x-delayed-type": "direct",
		},
	); err != nil {
		return fmt.Errorf("failed to declare delayed exchange")
	}

	if err := r.ch.ExchangeDeclare(
		exchangeDirect,
		"direct",
		true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("failed to declare direct exchange")
	}

	channels := []string{"telegram"} // only telegram is supported for now

	for _, ch := range channels {
		queueName := fmt.Sprintf("notifications.%s", ch)

		if _, err := r.ch.QueueDeclare(
			queueName,
			true, false, false, false, nil,
		); err != nil {
			return fmt.Errorf("failed to declare queue %s: %v", queueName, err)
		}

		if err := r.ch.QueueBind(
			queueName, ch,
			exchangeDelayed, false, nil,
		); err != nil {
			return fmt.Errorf("failed to bind queue %s to delayed exchange: %v", queueName, err)
		}

		if err := r.ch.QueueBind(
			queueName, ch, exchangeDirect, false, nil,
		); err != nil {
			return fmt.Errorf("failed to bind queue %s to direct exchange: %v", queueName, err)
		}
	}
	return nil
}

// Close closes the RabbitMQ instance.
func (r *RabbitMQ) Close() {
	if r.ch != nil {
		_ = r.ch.Close()
	}
	if r.conn != nil {
		_ = r.conn.Close()
	}
}

// Publish publishes a models.Notification to
// either the delayed queue or the direct queue, based on the expiration factor.
func (r *RabbitMQ) Publish(ctx context.Context, n models.Notification) error {
	body, _ := json.Marshal(n)
	var (
		exchange string
		exp      int64
		headers  amqp.Table
	)
	if n.SendAt.Before(time.Now()) {
		headers = amqp.Table{}
		exchange = exchangeDirect
	} else {
		exp = time.Until(n.SendAt).Milliseconds()
		exchange = exchangeDelayed
		headers = amqp.Table{
			"x-delay": exp,
		}
	}
	if err := r.ch.Publish(
		exchange, n.Channel,
		false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Headers:      headers,
			Timestamp:    time.Now(),
			MessageId:    n.ID,
		},
	); err != nil {
		return fmt.Errorf("failed to publish delayed")
	}

	r.log.Debug().Str("notification_id", n.ID).Int64("expiration_ms", exp).Msg("Published notification to exchange " + exchange)
	return nil
}

func (r *RabbitMQ) Consume(ctx context.Context, consumerName, queueName string) (<-chan models.QueuePayload, error) {
	msgs, err := r.ch.Consume(
		queueName, consumerName,
		false, false, false, false, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register consumer: %v", err)
	}
	out := make(chan models.QueuePayload)

	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case m, ok := <-msgs:
				if !ok {
					return
				}

				var n models.Notification
				if err := json.Unmarshal(m.Body, &n); err != nil {
					continue
				}

				select {
				case <-ctx.Done():
				case out <- models.QueuePayload{
					Notify: &n,
					Commit: func() error {
						return m.Ack(false)
					},
					Discard: func() error {
						return m.Nack(false, false)
					},
				}:
				}
			}
		}
	}()

	return out, nil
}
