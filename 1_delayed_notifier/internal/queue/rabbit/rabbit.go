package rabbit

import (
	"context"
	"strconv"
	"time"

	"delayed-notifier/internal/models"

	amqp091 "github.com/rabbitmq/amqp091-go"
)

// RabbitConfig holds connection settings for RabbitMQ.
type RabbitConfig struct {
	Host      string
	Port      int
	Username  string
	Password  string
	QueueName string
}

// Rabbit is a RabbitMQ-backed implementation for publishing and consuming notifications.
type Rabbit struct {
	conn      *amqp091.Connection
	ch        *amqp091.Channel
	queue     amqp091.Queue
	queueName string
}

// amqpDelivery implements models.Delivery
type amqpDelivery struct {
	d *amqp091.Delivery
}

func (m amqpDelivery) Body() []byte            { return m.d.Body }
func (m amqpDelivery) Ack() error              { return m.d.Ack(false) }
func (m amqpDelivery) Nack(requeue bool) error { return m.d.Nack(false, requeue) }

// NewRabbit initializes a Rabbit connection and declares the queue.
func NewRabbit(cfg RabbitConfig) (*Rabbit, error) {
	url := "amqp://" + cfg.Username + ":" + cfg.Password + "@" + cfg.Host + ":" + strconv.Itoa(cfg.Port) + "/"
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	q, err := ch.QueueDeclare(
		cfg.QueueName,
		true,
		false,
		false,
		false,
		amqp091.Table{},
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}
	// Fair dispatch
	_ = ch.Qos(10, 0, false)
	return &Rabbit{conn: conn, ch: ch, queue: q, queueName: cfg.QueueName}, nil
}

// Publish sends a persistent JSON message to the configured queue.
func (r *Rabbit) Publish(ctx context.Context, body []byte) error {
	return r.ch.PublishWithContext(ctx,
		"",
		r.queueName,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
}

// Consume returns a channel of deliveries that implements models.Delivery.
func (r *Rabbit) Consume(ctx context.Context) (<-chan models.Delivery, error) {
	msgs, err := r.ch.Consume(
		r.queueName,
		"",
		false,
		false,
		false,
		false,
		amqp091.Table{},
	)
	if err != nil {
		return nil, err
	}
	out := make(chan models.Delivery)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-msgs:
				if !ok {
					return
				}
				dd := d
				out <- amqpDelivery{d: &dd}
			}
		}
	}()
	return out, nil
}

// Close terminates channel and connection.
func (r *Rabbit) Close() error {
	if r.ch != nil {
		r.ch.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
