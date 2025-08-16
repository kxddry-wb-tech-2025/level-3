package worker

import (
	"context"
	"delayed-notifier/internal/models"

	"github.com/rs/zerolog"
)

// Broker can be something like Kafka streams or RabbitMQ. It's used for working with delayed notifications.
type Broker interface {
	PublishDelayed(ctx context.Context, n models.Notification) error
	GetReady(ctx context.Context) (<-chan models.Notification, error)
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
