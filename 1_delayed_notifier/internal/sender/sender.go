package sender

import (
	"context"
	"delayed-notifier/internal/models"
	"errors"
)

type Sender interface {
	Send(ctx context.Context, n *models.Notification) error
}

var ErrUnsupportedChannel = errors.New("unsupported channel")
