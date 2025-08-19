package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"image-processor/internal/domain"
	"time"

	"github.com/kxddry/wbf/kafka"
	"github.com/kxddry/wbf/retry"
)

type Broker struct {
	strat    retry.Strategy
	producer *kafka.Producer
}

func NewBroker(brokers []string, topic string, strat retry.Strategy) *Broker {
	producer := kafka.NewProducer(brokers, topic)
	return &Broker{producer: producer, strat: strat}
}

type task struct {
	FileName  string    `json:"file_name"`
	CreatedAt time.Time `json:"created_at"`
}

func (b *Broker) SendTask(ctx context.Context, t *domain.Task) error {
	const op = "broker.kafka.SendTask"
	tt := task{t.FileName, t.CreatedAt}
	data, err := json.Marshal(tt)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := b.producer.SendWithRetry(ctx, b.strat, []byte(t.FileName), data); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
