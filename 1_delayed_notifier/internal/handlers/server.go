package handlers

import (
	"context"
	"delayed-notifier/internal/models"
)

type Publisher interface {
	PublishDelayed(ctx context.Context, n models.Notification) error
}

type StatusStorage interface {
	Set(ctx context.Context, st models.NotificationStatus) error
	Get(ctx context.Context, id string) (*models.NotificationStatus, error)
}

type Server struct {
	pub   Publisher
	store StatusStorage
}

func NewServer(pub Publisher, s StatusStorage) *Server {
	return &Server{
		pub:   pub,
		store: s,
	}
}
