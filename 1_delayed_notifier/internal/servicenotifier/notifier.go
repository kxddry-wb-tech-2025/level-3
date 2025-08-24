// Package servicenotifier is a service that notifies other services about the attempt to send a notification.
package servicenotifier

import (
	"context"
	"delayed-notifier/internal/models"

	"github.com/kxddry/wbf/zlog"
)

type Producer interface {
	Publish(ctx context.Context, n models.NotificationKafka) error
}

type Notifier struct {
	prod Producer
}

func NewNotifier(producer Producer) *Notifier {
	return &Notifier{prod: producer}
}

func (n *Notifier) Notify(ctx context.Context, in <-chan models.NotificationKafka) {
	log := zlog.Logger.With().Str("component", "servicenotifier").Logger()
	for {
		select {
		case <-ctx.Done():
		case nk, ok := <-in:
			if !ok {
				return
			}
			if err := n.prod.Publish(ctx, nk); err != nil {
				log.Error().Err(err).Msg("failed to publish notification")
			} else {
				log.Debug().Any("notification", nk).Msg("published notification")
			}
		}
	}
}
