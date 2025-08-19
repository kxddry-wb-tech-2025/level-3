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

type Producer struct {
	strat    retry.Strategy
	producer *kafka.Producer
}

func NewProducer(brokers []string, topic string, strat retry.Strategy) *Producer {
	return &Producer{
		strat:    strat,
		producer: kafka.NewProducer(brokers, topic),
	}
}

type task struct {
	FileName  string    `json:"file_name"`
	CreatedAt time.Time `json:"created_at"`
}

func (p *Producer) SendTask(ctx context.Context, t *domain.Task) error {
	const op = "broker.kafka.SendTask"
	tt := task{t.FileName, t.CreatedAt}
	data, err := json.Marshal(tt)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := p.producer.SendWithRetry(ctx, p.strat, []byte(t.FileName), data); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}
