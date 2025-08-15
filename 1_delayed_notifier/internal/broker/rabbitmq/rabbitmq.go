package rabbitmq

import amqp "github.com/rabbitmq/amqp091-go"

type RabbitMQ struct {
	conn *amqp.Connection
	q    *amqp.Queue
	ch   *amqp.Channel
}

func New(dsn, queueName string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(dsn)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	return &RabbitMQ{conn: conn, q: &q, ch: ch}, nil
}

func (r *RabbitMQ) Close() error {
	return r.conn.Close()
}

func (r *RabbitMQ) Publish(exchange string, msg []byte) error {
	return r.ch.Publish(exchange, r.q.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        msg,
	})
}
