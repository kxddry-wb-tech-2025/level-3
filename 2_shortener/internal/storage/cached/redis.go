package cached

import (
	"context"
	"errors"
	"fmt"
	"shortener/internal/storage"
	"time"

	// sorry, i am not using wb-go/wbf/redis because it is shit and does not support TTL
	"github.com/redis/go-redis/v9"
)

const (
	ttl     = 24 * time.Hour
	keyLink = "link:%s"
	keyHits = "hits:%s"
)

// Redis is an implementation of the CacheStorage interface.
type Redis struct {
	client *redis.Client
}

// New creates a new Redis instance.
func New(ctx context.Context, addr, password string, db int) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &Redis{client: client}, nil
}

// Close closes the Redis client.
func (r *Redis) Close() error {
	return r.client.Close()
}

// GetURL gets the URL from the Redis client.
func (r *Redis) GetURL(ctx context.Context, shortCode string) (string, error) {
	keyL := fmt.Sprintf(keyLink, shortCode)

	url, err := r.client.Get(ctx, keyL).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", storage.ErrNotFound
		}
		return "", err
	}
	go r.client.Incr(ctx, fmt.Sprintf(keyHits, shortCode))
	return url, nil
}

// SetURL sets the URL in the Redis client.
func (r *Redis) SetURL(ctx context.Context, shortCode, url string, usage int64) error {
	keyL := fmt.Sprintf(keyLink, shortCode)
	keyH := fmt.Sprintf(keyHits, shortCode)

	pipe := r.client.TxPipeline()
	pipe.Set(ctx, keyL, url, ttl)
	pipe.Set(ctx, keyH, usage, ttl)
	_, err := pipe.Exec(ctx)
	return err
}
