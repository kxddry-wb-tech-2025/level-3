package redis

import (
	"context"
	"delayed-notifier/internal/models"
	"delayed-notifier/internal/storage"
	"encoding/json"
	"time"

	"github.com/wb-go/wbf/redis"
)

type Storage struct {
	rdb *redis.Client
}

func New(addr, password string, db int) *Storage {
	return &Storage{
		rdb: redis.New(addr, password, db),
	}
}

func (s *Storage) key(id string) string { return "notif:" + id }

func (s *Storage) Set(ctx context.Context, st models.NotificationStatus) error {
	data, _ := json.Marshal(st)
	return s.rdb.SetWithRetry(ctx, storage.Strategy, s.key(st.ID), data)
}

func (s *Storage) Cancel(ctx context.Context, id string) error {
	return s.Set(ctx, models.NotificationStatus{
		ID:        id,
		Status:    models.StatusCanceled,
		UpdatedAt: time.Time{},
	})
}

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
