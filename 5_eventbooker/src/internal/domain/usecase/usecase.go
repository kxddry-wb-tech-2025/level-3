package usecase

import (
	"context"
	"time"

	"eventbooker/src/internal/domain"
)

// EventRepository is the interface for the event repository.
type EventRepository interface {
	CreateEvent(ctx context.Context, event domain.CreateEventRequest) (string, error)
	UpdateEvent(ctx context.Context, event domain.Event) error
	GetEvent(ctx context.Context, eventID string) (domain.Event, error)
}

// BookingRepository is the interface for the booking repository.
type BookingRepository interface {
	Book(ctx context.Context, eventID, userID string, paymentDeadline time.Time) (string, error)
	GetBooking(ctx context.Context, bookingID string) (domain.Booking, error)
	Confirm(ctx context.Context, bookingID string) (string, error)
	BookingSetDecremented(ctx context.Context, bookingID string, decremented bool) error
	BookingSetStatus(ctx context.Context, bookingID string, status string) error
}

// NotificationRepository is the interface for the notification repository.
type NotificationRepository interface {
	AddNotification(ctx context.Context, notif domain.DelayedNotification) error
	GetNotificationID(ctx context.Context, bookingID string) (string, error)
}

// Tx is the interface for the transactions.
type Tx interface {
	EventRepository
	BookingRepository
	NotificationRepository
	Commit() error
	Rollback() error
}

// TxManager is the interface for the transaction manager.
type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error
}

// NotificationService is the interface for the notification service.
type NotificationService interface {
	SendNotification(notif domain.DelayedNotification) (string, error)
	CancelNotification(notificationID string) error
}

// CancellationService is the interface for the cancellation service.
type CancellationService interface {
	Cancellations(ctx context.Context) (<-chan domain.CancelBookingEvent, error)
}

// Usecase is the usecase for the event booking service.
type Usecase struct {
	cs      CancellationService
	nf      NotificationService
	storage TxManager
}

// New is the constructor for the Usecase.
// It is responsible for creating a new Usecase and handling cancellations.
func New(ctx context.Context, cs CancellationService, nf NotificationService, storage TxManager) (*Usecase, error) {
	uc := &Usecase{
		cs:      cs,
		nf:      nf,
		storage: storage,
	}
	// handle cancellations
	err := uc.HandleCancellations(ctx)
	if err != nil {
		return nil, err
	}
	return uc, nil
}
