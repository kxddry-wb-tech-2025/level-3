package worker

import (
	"context"
	"encoding/json"
	"time"

	"delayed-notifier/internal/models"

	"github.com/kxddry/wbf/retry"
	"github.com/kxddry/wbf/zlog"
)

// ConsumerQueue abstracts the queue for the consumer.
type ConsumerQueue interface {
	Consume(ctx context.Context) (<-chan models.Delivery, error)
}

// Sender sends a notification via a channel.
type Sender interface {
	Send(ctx context.Context, n *models.Notification) error
}

// Consumer processes messages from the queue and delivers notifications.
type Consumer struct {
	store  storageAccess
	q      ConsumerQueue
	sender Sender
}

// storageAccess is the subset of storage methods used by the consumer.
type storageAccess interface {
	GetNotification(ctx context.Context, id string) (*models.Notification, error)
	SaveNotification(ctx context.Context, n *models.Notification) error
	AddToRetry(ctx context.Context, id string, when time.Time) error
}

// NewConsumer constructs a Consumer.
func NewConsumer(store storageAccess, q ConsumerQueue, s Sender) *Consumer {
	return &Consumer{store: store, q: q, sender: s}
}

// Run starts consumption loop until ctx is cancelled.
func (c *Consumer) Run(ctx context.Context) {
	msgs, err := c.q.Consume(ctx)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("consumer: failed to start consuming")
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-msgs:
			if !ok {
				return
			}
			c.processDelivery(ctx, d)
		}
	}
}

func (c *Consumer) processDelivery(ctx context.Context, d models.Delivery) {
	var n models.Notification
	if err := json.Unmarshal(d.Body(), &n); err != nil {
		_ = d.Nack(false)
		zlog.Logger.Error().Err(err).Msg("consumer: bad payload")
		return
	}
	if n.Status == models.StatusCancelled {
		_ = d.Ack()
		return
	}
	// Send via sender with short retry strategy; schedule long retry if still failing
	short := retry.Strategy{Attempts: 3, Delay: 10 * time.Millisecond, Backoff: 2}
	if err := retry.Do(func() error { return c.sender.Send(ctx, &n) }, short); err != nil {
		zlog.Logger.Error().Err(err).Str("id", n.ID).Msg("consumer: send failed")
		n.Status = models.StatusRetrying
		n.RetryCount++
		next := time.Now().Add(computeBackoff(n.RetryCount)).UTC()
		n.NextAttemptAt = &next
		n.LastError = err.Error()
		n.UpdatedAt = time.Now().UTC()
		_ = c.store.SaveNotification(ctx, &n)
		_ = c.store.AddToRetry(ctx, n.ID, next)
		_ = d.Ack()
		return
	}
	// Success
	n.Status = models.StatusSent
	n.UpdatedAt = time.Now().UTC()
	_ = c.store.SaveNotification(ctx, &n)
	_ = d.Ack()
}

func computeBackoff(retry int) time.Duration {
	if retry <= 0 {
		return 2 * time.Second
	}
	max := 6 * time.Hour
	base := 2 * time.Second
	d := base << (retry - 1)
	if d > max {
		return max
	}
	return d
}
