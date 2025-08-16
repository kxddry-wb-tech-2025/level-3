package worker

import (
	"context"
	"delayed-notifier/internal/models"
	"fmt"

	"github.com/rs/zerolog"
)

// Broker can be something like Kafka streams or RabbitMQ. It's used for working with delayed notifications.
type Broker interface {
	Publish(ctx context.Context, n models.Notification) error
	Consume(ctx context.Context) (<-chan models.QueuePayload, error)
}

// StatusStore is used as an interface to store and get statuses
type StatusStore interface {
	Set(ctx context.Context, st models.NotificationStatus) error
	Get(ctx context.Context, id string) (*models.NotificationStatus, error)
}

// Sender is something that can send notifications
type Sender interface {
	Send(ctx context.Context, n models.Notification) error
	Name() string
}

// NotificationProcessor is a structure for processing delayed notifications and sending them over
type NotificationProcessor struct {
	b     Broker
	store StatusStore
	tg    Sender
	log   *zerolog.Logger
}

func New(b Broker, st StatusStore, tg Sender, log *zerolog.Logger) *NotificationProcessor {
	return &NotificationProcessor{
		b:     b,
		store: st,
		tg:    tg,
		log:   log,
	}
}

func (np *NotificationProcessor) Start(ctx context.Context) {
	channels := []string{"telegram"} // only telegram is supported for now

	for _, ch := range channels {
		go func(channel string) {
			if err := np.consume(ctx, channel); err != nil {
				np.log.Error().Err(err).Msg("error consuming channel")
			}
		}(ch)
	}

}

func (np *NotificationProcessor) consume(ctx context.Context, channel string) error {
	queueName := fmt.Sprintf("delayed_notifier.%s", channel)

	msgs, err := np.b.Consume(ctx)
}
