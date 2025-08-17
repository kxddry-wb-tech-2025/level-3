package storage

import (
	"context"
	"errors"
	"shortener/internal/domain"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
)

type URLStorage interface {
	SaveURL(ctx context.Context, url string) (string, error)
	GetURL(ctx context.Context, shortCode string) (string, error)
}

type ClickStorage interface {
	SaveClick(ctx context.Context, click domain.Click) error
	GetClicks(ctx context.Context, shortCode string, limit, offset int) ([]domain.Click, error)
	ClickCount(ctx context.Context, shortCode string) (int64, error)
	UniqueClickCount(ctx context.Context, shortCode string) (int64, error)
	ClicksByDay(ctx context.Context, shortCode string, start, end time.Time) (map[string]int64, error)
	ClicksByMonth(ctx context.Context, shortCode string, start, end time.Time) (map[string]int64, error)
	ClicksByUserAgent(ctx context.Context, shortCode string, start, end time.Time, limit int) (map[string]int64, error)
}