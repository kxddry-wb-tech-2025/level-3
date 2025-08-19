package kafka

import (
	"context"
	"encoding/json"
	"image-processor/internal/domain"

	"github.com/kxddry/wbf/kafka"
	"github.com/kxddry/wbf/retry"
	"github.com/kxddry/wbf/zlog"
	kafka_go "github.com/segmentio/kafka-go"
)

// Consumer is the struct that contains the consumer and the strategy.
type Consumer struct {
	strat    retry.Strategy
	consumer *kafka.Consumer
}

// NewConsumer creates a new consumer with the given brokers, topic, groupID, and strategy.
func NewConsumer(brokers []string, topic, groupID string, strat retry.Strategy) *Consumer {
	return &Consumer{
		strat:    strat,
		consumer: kafka.NewConsumer(brokers, topic, groupID),
	}
}

// StartConsuming starts consuming messages from the Kafka topic.
func (c *Consumer) StartConsuming(ctx context.Context, out chan<- *domain.KafkaMessage) {
	const op = "broker.kafka.Consumer.StartConsuming"
	in := make(chan kafka_go.Message)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-in:
				if !ok {
					return
				}
				var t domain.Task
				if err := json.Unmarshal(msg.Value, &t); err != nil {
					zlog.Logger.Err(err).Str("op", op).Msg("failed to unmarshal task")
					continue
				}

				km := domain.KafkaMessage{
					Task: t,
					Commit: func() error {
						return c.consumer.Commit(ctx, msg)
					},
				}

				t.Status = string(msg.Key)
				select {
				case out <- &km:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	c.consumer.StartConsuming(ctx, in, c.strat)
}

// Close closes the consumer.
func (c *Consumer) Close() error {
	return c.consumer.Close()
}
