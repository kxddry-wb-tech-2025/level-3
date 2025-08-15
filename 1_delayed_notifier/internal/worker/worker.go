package worker

import (
	"context"
	"delayed-notifier/internal/models"
	"delayed-notifier/internal/storage"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/wb-go/wbf/retry"
)

type Broker interface {
	PublishDelayed(ctx context.Context, n models.Notification) error
	GetReady(ctx context.Context) (<-chan models.Notification, error)
}

type StatusStore interface {
	Set(ctx context.Context, st models.NotificationStatus) error
	Get(ctx context.Context, id string) (*models.NotificationStatus, error)
}

type Sender interface {
	Send(ctx context.Context, n models.Notification) error
	Name() string
}

type Worker struct {
	b       Broker
	store   StatusStore
	senders map[string]Sender
	log     *zerolog.Logger
}

func New(b Broker, st StatusStore, log *zerolog.Logger, senders []Sender) *Worker {
	m := map[string]Sender{}
	for _, s := range senders {
		m[s.Name()] = s
	}
	return &Worker{
		b:       b,
		store:   st,
		senders: m,
		log:     log,
	}
}

func (w *Worker) Start(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		msgs, err := w.b.GetReady(ctx)
		if err != nil {
			w.log.Error().Err(err).Msg("worker consume error")
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}

				w.handle(ctx, msg)
			}
		}
	}()
}

func (w *Worker) handle(ctx context.Context, msg models.Notification) {
	st, err := w.store.Get(ctx, msg.ID)
	if err == nil && st.Status == models.StatusCanceled {
		w.log.Info().Msg(st.ID + " canceled, dropping notification")
		_ = w.store.Set(ctx, models.NotificationStatus{ID: msg.ID, Status: st.Status, UpdatedAt: time.Now()})
		return
	}
	_ = w.store.Set(ctx, models.NotificationStatus{ID: msg.ID, Status: models.StatusReceived, UpdatedAt: time.Now()})

	sender, ok := w.senders[msg.Channel]
	if !ok {
		w.log.Error().Msg(msg.Channel + " is not registered")
		w.retry(ctx, msg)
	}
	if err = sender.Send(ctx, msg); err != nil {
		w.log.Error().Err(err).Msg(msg.Channel)
		w.retry(ctx, msg)
		return
	}

	_ = w.store.Set(ctx, models.NotificationStatus{ID: msg.ID, Status: models.StatusSent, UpdatedAt: time.Now()})
}

func (w *Worker) retry(ctx context.Context, msg models.Notification) {
	err := retry.Do(func() error {
		return w.b.PublishDelayed(ctx, msg)
	}, storage.Strategy)
	if err != nil {
		w.log.Error().Err(err).Msg(msg.Channel)
		err = retry.Do(func() error {
			return w.store.Set(ctx, models.NotificationStatus{ID: msg.ID, Status: models.StatusFailed, UpdatedAt: time.Now()})
		}, storage.Strategy)
		if err != nil {
			w.log.Error().Err(err).Msg(msg.Channel)
		}
	}
}
