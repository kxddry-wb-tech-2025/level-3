package sender

import (
	"context"
	"delayed-notifier/internal/models"
	"errors"
)

// Sender delivers a notification via a particular channel.
type Sender interface {
	Send(ctx context.Context, n *models.Notification) error
}

// ErrUnsupportedChannel is returned when a channel is not supported by a sender.
var ErrUnsupportedChannel = errors.New("unsupported channel")
