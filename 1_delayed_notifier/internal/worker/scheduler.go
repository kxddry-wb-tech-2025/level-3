package worker

import (
	"context"
	"encoding/json"
	"time"

	"delayed-notifier/internal/models"

	"github.com/kxddry/wbf/zlog"
)

// NotificationStore abstracts storage operations needed by the scheduler.
type NotificationStore interface {
	PopDue(ctx context.Context, which string, now time.Time, limit int64) ([]string, error)
	GetNotification(ctx context.Context, id string) (*models.Notification, error)
	SaveNotification(ctx context.Context, n *models.Notification) error
	EnqueueNow(ctx context.Context, id string) error
	AddToRetry(ctx context.Context, id string, when time.Time) error
}

// Publisher abstracts queue publishing.
type Publisher interface {
	Publish(ctx context.Context, body []byte) error
}

// Scheduler scans Redis zsets for due notifications and publishes them to RabbitMQ.
type Scheduler struct {
	store NotificationStore
	q     Publisher
}

func NewScheduler(store NotificationStore, q Publisher) *Scheduler {
	return &Scheduler{store: store, q: q}
}

func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			// process due
			s.publishDue(ctx, now)
			// process retry
			s.publishRetry(ctx, now)
		}
	}
}

func (s *Scheduler) publishDue(ctx context.Context, now time.Time) {
	ids, err := s.store.PopDue(ctx, "due", now, 100)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("scheduler: pop due")
		return
	}
	for _, id := range ids {
		n, err := s.store.GetNotification(ctx, id)
		if err != nil || n == nil {
			if err != nil {
				zlog.Logger.Error().Err(err).Str("id", id).Msg("scheduler: get notification")
			}
			continue
		}
		if n.Status == models.StatusCancelled {
			continue
		}
		n.Status = models.StatusQueued
		n.UpdatedAt = now.UTC()
		_ = s.store.SaveNotification(ctx, n)
		payload, _ := json.Marshal(n)
		if err := s.q.Publish(ctx, payload); err != nil {
			zlog.Logger.Error().Err(err).Str("id", id).Msg("scheduler: publish")
			// re-enqueue shortly to avoid losing it
			_ = s.store.EnqueueNow(ctx, id)
		}
	}
}

func (s *Scheduler) publishRetry(ctx context.Context, now time.Time) {
	ids, err := s.store.PopDue(ctx, "retry", now, 100)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("scheduler: pop retry")
		return
	}
	for _, id := range ids {
		n, err := s.store.GetNotification(ctx, id)
		if err != nil || n == nil {
			if err != nil {
				zlog.Logger.Error().Err(err).Str("id", id).Msg("scheduler: get notification (retry)")
			}
			continue
		}
		if n.Status == models.StatusCancelled {
			continue
		}
		n.Status = models.StatusQueued
		n.UpdatedAt = now.UTC()
		_ = s.store.SaveNotification(ctx, n)
		payload, _ := json.Marshal(n)
		if err := s.q.Publish(ctx, payload); err != nil {
			zlog.Logger.Error().Err(err).Str("id", id).Msg("scheduler: publish retry")
			_ = s.store.AddToRetry(ctx, id, now.Add(5*time.Second))
		}
	}
}
