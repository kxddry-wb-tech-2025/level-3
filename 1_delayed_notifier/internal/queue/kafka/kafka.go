package kafka

import (
	"context"
	"delayed-notifier/internal/models"
	"fmt"
	"time"

	"github.com/kxddry/wbf/zlog"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	prod *kafka.Writer
}

func New(ctx context.Context, brokers []string, topic string) (*Producer, error) {
	prod := kafka.NewWriter(kafka.WriterConfig{
		Brokers:           brokers,
		Topic:             topic,
		Dialer:            kafka.DefaultDialer,
		Balancer:          new(kafka.LeastBytes),
		MaxAttempts:       10,
		QueueCapacity:     100,
		BatchSize:         1,
		BatchBytes:        1024,
		BatchTimeout:      10 * time.Millisecond,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		RebalanceInterval: 10 * time.Second,
		IdleConnTimeout:   10 * time.Second,
		RequiredAcks:      -1,
		Async:             false,
		CompressionCodec:  kafka.Snappy.Codec(),
		Logger: kafka.LoggerFunc(func(msg string, args ...any) {
			log := zlog.Logger.With().Str("kafka", msg).Logger()
			log.Info().Msgf(msg, args...)
		}),
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...any) {
			log := zlog.Logger.With().Str("kafka", msg).Logger()
			log.Error().Msgf(msg, args...)
		}),
	})

	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return nil, fmt.Errorf("failed to dial kafka: %w", err)
	}
	defer conn.Close()

	_, err = conn.ReadPartitions(topic)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metadata for topic %q: %w", topic, err)
	}

	return &Producer{prod: prod}, nil
}

func (p *Producer) Publish(ctx context.Context, n models.NotificationKafka) error {
	msg := kafka.Message{
		Key:   []byte(n.NotificationID),
		Value: []byte(n.Message),
	}
	return p.prod.WriteMessages(ctx, msg)
}
