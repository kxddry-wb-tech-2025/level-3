package usecase

import (
	"context"
	"time"

	"eventbooker/src/internal/domain"

	"github.com/go-playground/validator/v10"
)

type EventRepository interface {
	CreateEvent(ctx context.Context, event domain.CreateEventRequest) (string, error)
	UpdateEvent(ctx context.Context, event domain.Event) error
	GetEvent(ctx context.Context, eventID string) (domain.Event, error)
}

type BookingRepository interface {
	Book(ctx context.Context, eventID, userID string, paymentDeadline time.Time) (string, error)
	GetBooking(ctx context.Context, bookingID string) (domain.Booking, error)
	Confirm(ctx context.Context, bookingID string) (string, error)
}

type NotificationRepository interface {
	Add(ctx context.Context, notif domain.DelayedNotification) error
}

type Tx interface {
	EventRepository
	BookingRepository
	Commit() error
	Rollback() error
}

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error
}

type NotificationService interface {
	SendDelayed(notif domain.DelayedNotification) error
	CancelDelayed(bookingID string) error
}

// BookingCache is the interface for the booking cache.
type BookingCache interface {
	Add(key string, value domain.Booking, ttl time.Duration) error
	Get(key string) (domain.Booking, error)
}

// Usecase is the usecase for the event booking service.
type Usecase struct {
	nfs       NotificationService
	validator *validator.Validate
	storage   TxManager
}

func NewUsecase(nfs NotificationService, storage TxManager) *Usecase {
	return &Usecase{
		nfs:       nfs,
		validator: validator.New(),
		storage:   storage,
	}
}
