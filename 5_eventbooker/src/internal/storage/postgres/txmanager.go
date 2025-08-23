package postgres

import (
	"context"
	"eventbooker/src/internal/domain"
	"eventbooker/src/internal/domain/usecase"
	"eventbooker/src/internal/storage/repositories/booking"
	"eventbooker/src/internal/storage/repositories/events"
	"eventbooker/src/internal/storage/repositories/notifications"
	"time"

	"github.com/kxddry/wbf/dbpg"
)

type TxManager struct {
	db    *dbpg.DB
	repos *Repositories
}

type Repositories struct {
	EventRepository        *events.EventRepository
	BookingRepository      *booking.BookingRepository
	NotificationRepository *notifications.NotificationRepository
}

func NewTxManager(db *dbpg.DB, repos *Repositories) *TxManager {
	return &TxManager{db: db, repos: repos}
}

// tx implements the domain Tx interface by delegating to repositories.
type tx struct {
	ctx   context.Context
	repos *Repositories
}

func (t *tx) CreateEvent(ctx context.Context, event domain.CreateEventRequest) (string, error) {
	return t.repos.EventRepository.CreateEvent(ctx, event)
}

func (t *tx) UpdateEvent(ctx context.Context, event domain.Event) error {
	return t.repos.EventRepository.UpdateEvent(ctx, event)
}

func (t *tx) GetEvent(ctx context.Context, eventID string) (domain.Event, error) {
	return t.repos.EventRepository.GetEvent(ctx, eventID)
}

func (t *tx) Book(ctx context.Context, eventID, userID string, paymentDeadline time.Time) (string, error) {
	// default decremented to false; it will be updated after notification success
	return t.repos.BookingRepository.Book(ctx, eventID, userID, paymentDeadline, false)
}

func (t *tx) GetBooking(ctx context.Context, bookingID string) (domain.Booking, error) {
	return t.repos.BookingRepository.GetBooking(ctx, bookingID)
}

func (t *tx) Confirm(ctx context.Context, bookingID string) (string, error) {
	return t.repos.BookingRepository.Confirm(ctx, bookingID)
}

func (t *tx) BookingSetDecremented(ctx context.Context, bookingID string, decremented bool) error {
	return t.repos.BookingRepository.BookingSetDecremented(ctx, bookingID, decremented)
}

func (t *tx) BookingSetStatus(ctx context.Context, bookingID string, status string) error {
	return t.repos.BookingRepository.BookingSetStatus(ctx, bookingID, status)
}

// Commit is a no-op since repositories manage their own queries directly.
func (t *tx) Commit() error { return nil }

// Rollback is a no-op since repositories manage their own queries directly.
func (t *tx) Rollback() error { return nil }

// Do creates a transactional context and passes a tx facade to the callback.
// Note: current repositories operate directly on dbpg.DB without explicit sql.Tx,
// so we provide a logical transaction wrapper. If dbpg adds real transaction support,
// wire it here.
func (m *TxManager) Do(ctx context.Context, fn func(ctx context.Context, tx usecase.Tx) error) error {
	// Provide a tx facade that delegates to repos
	t := &tx{ctx: ctx, repos: m.repos}
	if err := fn(ctx, t); err != nil {
		_ = t.Rollback()
		return err
	}
	return t.Commit()
}
