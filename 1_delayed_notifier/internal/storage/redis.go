package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"delayed-notifier/internal/models"

	"github.com/kxddry/wbf/zlog"
	"github.com/redis/go-redis/v9"
)

// RedisConfig contains connection parameters for Redis.
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// RedisStorage persists notifications and schedules using Redis.
type RedisStorage struct {
	client *redis.Client
}

const (
	keyNotificationObj = "notify:obj:%s"
	keyDueZSet         = "notify:due"
	keyRetryZSet       = "notify:retry"
)

// NewRedisStorage constructs a RedisStorage and pings the server.
func NewRedisStorage(ctx context.Context, cfg RedisConfig) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &RedisStorage{client: client}, nil
}

// Close shuts down the underlying Redis client.
func (s *RedisStorage) Close() error {
	return s.client.Close()
}

// SaveNotification updates the stored notification object.
func (s *RedisStorage) SaveNotification(ctx context.Context, n *models.Notification) error {
	if n == nil || n.ID == "" {
		return errors.New("invalid notification")
	}
	bytes, err := json.Marshal(n)
	if err != nil {
		return err
	}
	key := fmt.Sprintf(keyNotificationObj, n.ID)
	return s.client.Set(ctx, key, bytes, 0).Err()
}

// CreateNotification stores a new notification and schedules it in the due set.
func (s *RedisStorage) CreateNotification(ctx context.Context, n *models.Notification) error {
	if n == nil || n.ID == "" {
		return errors.New("invalid notification")
	}
	// Persist object
	if err := s.SaveNotification(ctx, n); err != nil {
		return err
	}
	// Schedule in due zset
	return s.client.ZAdd(ctx, keyDueZSet, redis.Z{
		Score:  float64(n.SendAt.Unix()),
		Member: n.ID,
	}).Err()
}

// GetNotification returns a notification by id or nil if not found.
func (s *RedisStorage) GetNotification(ctx context.Context, id string) (*models.Notification, error) {
	key := fmt.Sprintf(keyNotificationObj, id)
	val, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	var n models.Notification
	if err := json.Unmarshal(val, &n); err != nil {
		return nil, err
	}
	return &n, nil
}

// CancelNotification sets status to cancelled and removes it from scheduling sets.
func (s *RedisStorage) CancelNotification(ctx context.Context, id string) error {
	n, err := s.GetNotification(ctx, id)
	if err != nil {
		return err
	}
	if n == nil {
		return nil
	}
	n.Status = models.StatusCancelled
	n.UpdatedAt = time.Now().UTC()
	if err := s.SaveNotification(ctx, n); err != nil {
		return err
	}
	// Remove from scheduling sets
	pipe := s.client.TxPipeline()
	pipe.ZRem(ctx, keyDueZSet, id)
	pipe.ZRem(ctx, keyRetryZSet, id)
	_, err = pipe.Exec(ctx)
	return err
}

// EnqueueNow pushes the id to the due set with score of now.
func (s *RedisStorage) EnqueueNow(ctx context.Context, id string) error {
	return s.client.ZAdd(ctx, keyDueZSet, redis.Z{Score: float64(time.Now().Unix()), Member: id}).Err()
}

// AddToRetry schedules the id for retry at the given time.
func (s *RedisStorage) AddToRetry(ctx context.Context, id string, when time.Time) error {
	return s.client.ZAdd(ctx, keyRetryZSet, redis.Z{Score: float64(when.Unix()), Member: id}).Err()
}

// PopDue pops up to 'limit' ids due at or before 'now' from the given zset key.
func (s *RedisStorage) PopDue(ctx context.Context, which string, now time.Time, limit int64) ([]string, error) {
	var zsetKey string
	switch which {
	case "due":
		zsetKey = keyDueZSet
	case "retry":
		zsetKey = keyRetryZSet
	default:
		return nil, errors.New("unknown zset kind")
	}
	// Fetch due ids
	vals, err := s.client.ZRangeByScore(ctx, zsetKey, &redis.ZRangeBy{
		Min:    "-inf",
		Max:    fmt.Sprintf("%d", now.Unix()),
		Offset: 0,
		Count:  limit,
	}).Result()
	if err != nil {
		return nil, err
	}
	if len(vals) == 0 {
		return nil, nil
	}
	// Remove fetched ids
	if err := s.client.ZRem(ctx, zsetKey, anySlice(vals)...).Err(); err != nil {
		zlog.Logger.Warn().Err(err).Msg("failed to ZREM due ids; possible duplicates may occur")
	}
	return vals, nil
}

func anySlice[T any](in []T) []any {
	out := make([]any, 0, len(in))
	for _, v := range in {
		out = append(out, v)
	}
	return out
}
