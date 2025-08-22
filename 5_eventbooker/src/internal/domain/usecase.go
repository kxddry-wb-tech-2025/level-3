package domain

import (
	"errors"
	"eventbooker/src/internal/storage"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kxddry/wbf/zlog"
)

type Repository interface {
	CreateEvent(event CreateEventRequest) (string, error)
	Book(eventID string, userID string) (Booking, error)
	Confirm(bookingID string) (string, error)
	GetEvent(eventID string) (EventDetailsResponse, error)
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
	nfs        NotificationService
	validator  *validator.Validate
	storage    Repository
	bookingTTL time.Duration
}

func NewUsecase(nfs NotificationService, storage Repository, bookingTTL time.Duration) *Usecase {
	return &Usecase{
		nfs:        nfs,
		validator:  validator.New(),
		storage:    storage,
		bookingTTL: bookingTTL,
	}
}

func (u *Usecase) CreateEvent(event CreateEventRequest) CreateEventResponse {
	if err := u.validator.Struct(event); err != nil {
		return CreateEventResponse{
			Error: err.Error(),
		}
	}

	if !event.Date.After(time.Now()) {
		return CreateEventResponse{
			Error: "event date must be in the future",
		}
	}

	id, err := u.storage.CreateEvent(event)
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
	event, err := u.storage.GetEvent(eventID)
	// 3 possible failure scenarios:
	// 1. event not found
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return BookResponse{
				Error: "event not found",
			}
		}
		return BookResponse{
			Error: err.Error(),
		}
	}

	// 2. event is full
	if event.Available <= 0 {
		return BookResponse{
			Error: "event is full",
		}
	}

	// 3. event is in the past
	if event.Date.Before(time.Now()) {
		return BookResponse{
			Error: "event is in the past",
		}
	}

	booking, err := u.storage.Book(eventID, userID)
	if err != nil {
		return BookResponse{
			Error: err.Error(),
		}
	}

	notif := DelayedNotification{
		SendAt:     &booking.PaymentDeadline,
		TelegramID: userID,
		EventID:    eventID,
		BookingID:  booking.ID,
	}

	if err := u.nfs.SendDelayed(notif); err != nil {
		zlog.Logger.Err(err).Msg("failed to send delayed notification")
	}

	return BookResponse{
		ID:              booking.ID,
		Status:          BookingStatusPending,
		PaymentDeadline: booking.PaymentDeadline,
	}
}
