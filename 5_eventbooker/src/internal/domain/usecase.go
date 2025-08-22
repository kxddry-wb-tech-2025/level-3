package domain

import (
	"time"

	"github.com/go-playground/validator/v10"
)

type Repository interface {
	CreateEvent(event *CreateEventRequest) (string, error)
	Book(eventID string, userID string) (string, error)
	Confirm(bookingID string) (string, error)
	GetEvent(eventID string) (*EventDetailsResponse, error)
}

type Usecase struct {
	validator  *validator.Validate
	storage    Repository
	bookingTTL time.Duration
}

func NewUsecase(storage Repository, bookingTTL time.Duration) *Usecase {
	return &Usecase{
		validator:  validator.New(),
		storage:    storage,
		bookingTTL: bookingTTL,
	}
}

func (u *Usecase) CreateEvent(event *CreateEventRequest) CreateEventResponse {
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
