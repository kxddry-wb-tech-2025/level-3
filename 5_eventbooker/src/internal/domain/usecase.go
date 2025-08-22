package domain

import (
	"context"
	"errors"
	"eventbooker/src/internal/storage"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kxddry/wbf/zlog"
)

type EventRepository interface {
	CreateEvent(event CreateEventRequest) (string, error)
	GetEvent(eventID string) (EventDetailsResponse, error)
}

type BookingRepository interface {
	Book(eventID, userID string, paymentDeadline time.Time) (string, error)
	GetBooking(bookingID string) (Booking, error)
	Confirm(bookingID string) (string, error)
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
	SendDelayed(notif DelayedNotification) error
	CancelDelayed(bookingID string) error
}

// BookingCache is the interface for the booking cache.
type BookingCache interface {
	Add(key string, value Booking, ttl time.Duration) error
	Get(key string) (Booking, error)
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

func (u *Usecase) CreateEvent(event CreateEventRequest) CreateEventResponse {
	if !event.Date.After(time.Now()) {
		return CreateEventResponse{
			Error: "event date must be in the future",
		}
	}

	var id string
	err := u.storage.Do(context.Background(), func(ctx context.Context, tx Tx) error {
		var err error
		id, err = tx.CreateEvent(event)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return CreateEventResponse{
			Error: err.Error(),
		}
	}

	return CreateEventResponse{
		ID: id,
	}
}

func (u *Usecase) Book(eventID string, userID string) BookResponse {
	var bookingID string
	var paymentDeadline time.Time
	err := u.storage.Do(context.Background(), func(ctx context.Context, tx Tx) error {
		event, err := tx.GetEvent(eventID)
		if err != nil {
			return err
		}
		if event.Available <= 0 {
			return errors.New("event is full")
		}
		if event.Date.Before(time.Now()) {
			return errors.New("event is in the past")
		}
		paymentDeadline = time.Now().Add(event.PaymentTTL)
		bookingID, err = tx.Book(eventID, userID, paymentDeadline)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return BookResponse{
			Error: err.Error(),
		}
	}

	notif := DelayedNotification{
		SendAt:     &paymentDeadline,
		TelegramID: userID,
		EventID:    eventID,
		BookingID:  bookingID,
	}

	if err := u.nfs.SendDelayed(notif); err != nil {
		zlog.Logger.Err(err).Msg("failed to send delayed notification")
	}

	return BookResponse{
		ID:              bookingID,
		Status:          BookingStatusPending,
		PaymentDeadline: paymentDeadline,
	}
}

func (u *Usecase) Confirm(eventID, bookingID string) ConfirmResponse {
	var status string
	err := u.storage.Do(context.Background(), func(ctx context.Context, tx Tx) error {
		booking, err := tx.GetBooking(bookingID)
		if err != nil {
			return err
		}
		if booking.EventID != eventID {
			return errors.New("booking does not belong to the event")
		}
		switch booking.Status {
		case BookingStatusCancelled:
			return errors.New("booking is cancelled")
		case BookingStatusConfirmed:
			return errors.New("booking is already confirmed")
		case BookingStatusExpired:
			return errors.New("booking is expired")
		case BookingStatusPending:
		default:
			zlog.Logger.Err(errors.New("invalid booking status")).Msg("invalid booking status")
			panic(errors.New("invalid booking status in the database"))
		}
		status, err = tx.Confirm(bookingID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return ConfirmResponse{
			Error: err.Error(),
		}
	}

	if err := u.nfs.CancelDelayed(bookingID); err != nil {
		zlog.Logger.Err(err).Msg("failed to cancel delayed notification")
	}

	return ConfirmResponse{
		Status: status,
	}
}

func (u *Usecase) GetEvent(eventID string) EventDetailsResponse {
	var event EventDetailsResponse
	err := u.storage.Do(context.Background(), func(ctx context.Context, tx Tx) error {
		var err error
		event, err = tx.GetEvent(eventID)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return errors.New("event not found")
			}
			return err
		}
		return nil
	})
	if err != nil {
		return EventDetailsResponse{
			Error: err.Error(),
		}
	}

	return event
}
