package domain

import "time"

type Error struct {
	Error string `json:"error,omitempty"`
}

type CreateEventRequest struct {
	Date     time.Time `json:"date"`
	Capacity int       `json:"capacity" validate:"required,min=1"`
	Name     string    `json:"name" validate:"required,min=1"`
}

type CreateEventResponse struct {
	ID string `json:"id"`
}

type BookRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

type BookResponse struct {
	ID              string    `json:"id"`
	Status          string    `json:"status"`
	PaymentDeadline time.Time `json:"payment_deadline" format:"rfc3339"`
}

type ConfirmRequest struct {
	BookingID string `json:"booking_id" validate:"required,uuid"`
}

type ConfirmResponse struct {
	Status string `json:"status"`
}

type EventDetailsResponse struct {
	Name      string    `json:"name"`
	Available int       `json:"available"`
	Capacity  int       `json:"capacity"`
	Date      time.Time `json:"date"`
}
