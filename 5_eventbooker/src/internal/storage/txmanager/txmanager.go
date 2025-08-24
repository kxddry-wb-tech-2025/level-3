package txmanager

import (
	"context"
	"database/sql"
	"eventbooker/src/internal/domain"
	"eventbooker/src/internal/domain/usecase"
	"eventbooker/src/internal/storage"
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

func New(masterDSN string, slaveDSNs ...string) (*TxManager, error) {
	db, err := dbpg.New(masterDSN, slaveDSNs, &dbpg.Options{
		MaxOpenConns: 100,
	})
	if err != nil {
		return nil, err
	}
	eventRepo, err := events.New(masterDSN, slaveDSNs...)
	if err != nil {
		return nil, err
	}
	bookingRepo, err := booking.New(masterDSN, slaveDSNs...)
	if err != nil {
		return nil, err
	}
	notificationRepo, err := notifications.New(masterDSN, slaveDSNs...)
	if err != nil {
		return nil, err
	}
	repos := &Repositories{
		EventRepository:        eventRepo,
		BookingRepository:      bookingRepo,
		NotificationRepository: notificationRepo,
	}
	return &TxManager{db: db, repos: repos}, db.Master.Ping()
}

func (m *TxManager) Close() error {
	for _, slave := range m.db.Slaves {
		_ = slave.Close()
	}
	_ = m.repos.EventRepository.Close()
	_ = m.repos.BookingRepository.Close()
	_ = m.repos.NotificationRepository.Close()
	return m.db.Master.Close()
}

type Repositories struct {
	EventRepository        *events.EventRepository
	BookingRepository      *booking.BookingRepository
	NotificationRepository *notifications.NotificationRepository
}

func NewTxManager(db *dbpg.DB, repos *Repositories) *TxManager {
	return &TxManager{db: db, repos: repos}
}

// tx implements the domain Tx interface by delegating to repositories over a real DB transaction.
type tx struct {
	ctx   context.Context
	repos *Repositories
	tx    *sql.Tx
}

func (t *tx) CreateEvent(ctx context.Context, event domain.CreateEventRequest) (string, error) {
	return t.repos.EventRepository.CreateEvent(storage.WithTx(ctx, t.tx), event)
}

func (t *tx) UpdateEvent(ctx context.Context, event domain.Event) error {
	return t.repos.EventRepository.UpdateEvent(storage.WithTx(ctx, t.tx), event)
}

func (t *tx) GetEvent(ctx context.Context, eventID string) (domain.Event, error) {
	return t.repos.EventRepository.GetEvent(storage.WithTx(ctx, t.tx), eventID)
}

func (t *tx) Book(ctx context.Context, eventID, userID string, paymentDeadline time.Time) (string, error) {
	return t.repos.BookingRepository.Book(storage.WithTx(ctx, t.tx), eventID, userID, paymentDeadline, false)
}

func (t *tx) GetBooking(ctx context.Context, bookingID string) (domain.Booking, error) {
	return t.repos.BookingRepository.GetBooking(storage.WithTx(ctx, t.tx), bookingID)
}

func (t *tx) Confirm(ctx context.Context, bookingID string) (string, error) {
	return t.repos.BookingRepository.Confirm(storage.WithTx(ctx, t.tx), bookingID)
}

func (t *tx) BookingSetDecremented(ctx context.Context, bookingID string, decremented bool) error {
	return t.repos.BookingRepository.BookingSetDecremented(storage.WithTx(ctx, t.tx), bookingID, decremented)
}

func (t *tx) BookingSetStatus(ctx context.Context, bookingID string, status string) error {
	return t.repos.BookingRepository.BookingSetStatus(storage.WithTx(ctx, t.tx), bookingID, status)
}

func (t *tx) AddNotification(ctx context.Context, notif domain.DelayedNotification) error {
	return t.repos.NotificationRepository.AddNotification(storage.WithTx(ctx, t.tx), notif)
}

func (t *tx) GetNotificationID(ctx context.Context, bookingID string) (string, error) {
	return t.repos.NotificationRepository.GetNotificationID(storage.WithTx(ctx, t.tx), bookingID)
}

func (t *tx) Commit() error   { return t.tx.Commit() }
func (t *tx) Rollback() error { return t.tx.Rollback() }

// Do runs a function within a DB transaction, committing on success and rolling back on error.
func (m *TxManager) Do(ctx context.Context, fn func(ctx context.Context, tx usecase.Tx) error) error {
	// Start transaction on master connection
	sqlTx, err := m.db.Master.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	t := &tx{ctx: ctx, repos: m.repos, tx: sqlTx}
	ctxWithTx := storage.WithTx(ctx, sqlTx)
	if err := fn(ctxWithTx, t); err != nil {
		_ = t.Rollback()
		return err
	}
	return t.Commit()
}
