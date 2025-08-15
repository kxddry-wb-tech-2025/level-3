package redis

import (
	"context"
	"delayed-notifier/internal/models"
	"delayed-notifier/internal/storage"
	"encoding/json"
	"time"

	"github.com/kxddry/wbf/redis"
)

// Storage implements the StatusStorage interfaces for storing statuses
type Storage struct {
	rdb *redis.Client
}

// New creates a Storage
func New(ctx context.Context, addr, password string, db int) (*Storage, error) {
	s := new(Storage)
	s.rdb = redis.New(addr, password, db)

	if err := s.rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Storage) key(id string) string { return "notif:" + id }

// Set sets a new notification to Redis
func (s *Storage) Set(ctx context.Context, st models.NotificationStatus) error {
	data, _ := json.Marshal(st)
	return s.rdb.SetWithRetry(ctx, storage.Strategy, s.key(st.ID), data)
}

// Cancel cancels a notification
func (s *Storage) Cancel(ctx context.Context, id string) error {
	return s.Set(ctx, models.NotificationStatus{
		ID:        id,
		Status:    models.StatusCanceled,
		UpdatedAt: time.Time{},
	})
}

// Get returns a notification.
func (s *Storage) Get(ctx context.Context, id string) (*models.NotificationStatus, error) {
	val, err := s.rdb.GetWithRetry(ctx, storage.Strategy, s.key(id))
	if err != nil {
		return nil, err
	}
	var st models.NotificationStatus
	if err := json.Unmarshal([]byte(val), &st); err != nil {
		return nil, err
	}
	return &st, nil
}
