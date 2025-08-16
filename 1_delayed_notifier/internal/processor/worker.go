package worker

import (
	"context"
	"delayed-notifier/internal/models"
	"fmt"
	"time"

	"github.com/kxddry/wbf/retry"
	"github.com/rs/zerolog"
)

// Broker can be something like Kafka streams or RabbitMQ. It's used for working with delayed notifications.
type Broker interface {
	Publish(ctx context.Context, n models.Notification) error
	Consume(ctx context.Context, queueName string) (<-chan models.QueuePayload, error)
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

	msgs, err := np.b.Consume(ctx, queueName)
	if err != nil {
		return fmt.Errorf("failed to regiter consumer for %s: %v", channel, err)
	}

	np.log.Info().Msg("started consumer for channel " + channel)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case m, ok := <-msgs:
			if !ok {
				return fmt.Errorf("channel closed")
			}
			if err := np.process(ctx, m.Notify); err != nil {
				np.log.Err(err).Msg("failed to process message")
				m.Discard()
			} else {
				m.Commit()
			}
		}
	}
}

func (np *NotificationProcessor) process(ctx context.Context, note *models.Notification) error {

	if status, err := np.store.Get(ctx, note.ID); err == nil && status.Status == models.StatusFailed {
		return nil
	}

	ch := note.Channel
	var err error

	if err = np.store.Set(ctx, models.NotificationStatus{
		ID:        note.ID,
		Status:    models.StatusReceived,
		UpdatedAt: time.Now(),
	}); err != nil {
		np.log.Err(err).Msg("failed to change notification status")
		return fmt.Errorf("failed to change notification status to received")
	}

	switch ch {
	case "telegram":
		err = np.tg.Send(ctx, *note)
	default:
		return fmt.Errorf("unknown channel %s", ch)
	}

	if err != nil {
		if err = np.store.Set(ctx, models.NotificationStatus{
			ID:        note.ID,
			Status:    models.StatusFailed,
			UpdatedAt: time.Now(),
		}); err != nil {
			np.log.Err(err).Msg("failed to change notification status to failed")
			return fmt.Errorf("failed to change notification status to failed")
		}

		if err = retry.Do(func() error {
			note.Attempt++
			return np.b.Publish(ctx, *note)
		}, retry.Strategy{
			Attempts: 3,
			Delay:    2,
			Backoff:  2,
		}); err != nil {
			np.log.Err(err).Msg("PERMANENT FAILURE !!!")
		}
	}
	return nil
}
