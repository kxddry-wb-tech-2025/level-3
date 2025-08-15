package rabbitmq

import (
	"context"
	"delayed-notifier/internal/models"
	"encoding/json"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

const (
	exchangeReady = "delayed_notifier.ready"
	queueReady    = "delayed_notifier.ready"
	queueDelayed  = "delayed_notifier.delayed"
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
	if err := r.ch.ExchangeDeclare(exchangeReady, "fanout", true, false, false, false, nil); err != nil {
		return err
	}

	readyArgs := amqp.Table{"x-queue-type": "classic"}
	if _, err := r.ch.QueueDeclare(queueReady, true, false, false, false, readyArgs); err != nil {
		return err
	}
	if err := r.ch.QueueBind(queueReady, "", exchangeReady, false, nil); err != nil {
		return err
	}
	// Очередь delayed с DLX на ready exchange
	args := amqp.Table{
		"x-dead-letter-exchange": exchangeReady,
		"x-queue-type":           "classic",
		"x-delayed-type":         "direct",
	}
	if _, err := r.ch.QueueDeclare(queueDelayed, true, false, false, false, args); err != nil {
		return err
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

// PublishDelayed publishes a models.Notification to a delayed queue
func (r *RabbitMQ) PublishDelayed(ctx context.Context, n models.Notification) error {
	body, _ := json.Marshal(n)
	var exp string
	if n.SendAt.Before(time.Now()) {
		exp = "0"
	} else {
		exp = strconv.FormatInt(time.Until(n.SendAt).Milliseconds(), 10)
	}
	r.log.Debug().Str("notification_id", n.ID).Str("expiration_ms", exp).Msg("Publishing delayed notification to RabbitMQ")
	return r.ch.PublishWithContext(ctx, "", queueDelayed, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
		Expiration:   exp,
		Headers: amqp.Table{
			"x-attempt": n.Attempt,
			"x-delay":   exp,
		},
		Timestamp: time.Now(),
		MessageId: n.ID,
		Type:      "notification",
	})
}

// GetReady returns a channel of ready-to-go notifications.
func (r *RabbitMQ) GetReady(ctx context.Context) (<-chan models.Notification, error) {
	ch, err := r.ch.Consume(queueReady, "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	out := make(chan models.Notification)

	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-ch:
				if !ok {
					return
				}
				var n models.Notification
				if err := json.Unmarshal(d.Body, &n); err != nil {
					_ = d.Nack(false, false) // bad payload
					continue
				}
				out <- n
				_ = d.Ack(false)
			}
		}
	}()

	return out, nil
}
