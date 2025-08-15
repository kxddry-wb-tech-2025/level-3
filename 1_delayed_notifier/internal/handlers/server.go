package handlers

import (
	"context"
	"delayed-notifier/internal/models"
)

// Publisher is an interface that can publish a delayed notification, i.e. to Kafka or RabbitMQ
type Publisher interface {
	PublishDelayed(ctx context.Context, n models.Notification) error
}

// StatusStorage is an interface for a storage that can get status quickly
type StatusStorage interface {
	Cancel(ctx context.Context, id string) error
	Set(ctx context.Context, st models.NotificationStatus) error
	Get(ctx context.Context, id string) (*models.NotificationStatus, error)
}

// Server is a struct that contains both of these interfaces.
type Server struct {
	pub   Publisher
	store StatusStorage
}

// NewServer creates a Server.
func NewServer(pub Publisher, s StatusStorage) *Server {
	return &Server{
		pub:   pub,
		store: s,
	}
}
