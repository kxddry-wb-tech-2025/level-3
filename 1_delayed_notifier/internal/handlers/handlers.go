package handlers

import (
	"context"
	"delayed-notifier/internal/models"
)

type Storage interface {
	Add(ctx context.Context, notify *models.Notification) (int64, error)
	Update(ctx context.Context, id int64, status int) error
	Get(ctx context.Context, id int64) (*models.Notification, error)
}
