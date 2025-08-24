package domain

import "time"

// I know that applying omitempty to all fields is not a good practice, but it's better than returning an interface{}.

// CreateEventRequest is the request body for creating an event.
type CreateEventRequest struct {
	Date       time.Time `json:"date" validate:"required" format:"rfc3339"`
	Capacity   int64     `json:"capacity" validate:"required,min=1,max=8589934592"`
	Name       string    `json:"name" validate:"required,min=1,max=255"`
	PaymentTTL int       `json:"payment_ttl" validate:"required,min=1,max=2592000"`
}

// CreateEventResponse is the response body for creating an event.
type CreateEventResponse struct {
	ID    string `json:"id,omitempty" validate:"required,uuid"`
	Error string `json:"error,omitempty"`
}

// BookRequest is the request body for booking an event.
type BookRequest struct {
	UserID     string `json:"user_id" validate:"required,uuid"`
	TelegramID int    `json:"telegram_id" validate:"required,min=100000000,max=999999999"`
}

// BookResponse is the response body for booking an event.
type BookResponse struct {
	ID              string     `json:"id,omitempty"`
	Status          string     `json:"status,omitempty"`
	PaymentDeadline *time.Time `json:"payment_deadline,omitempty" format:"rfc3339"`
	Error           string     `json:"error,omitempty"`
}

// DelayedNotification is the value object for a delayed notification.
// It is used to send a notification to a user at a specific time.
type DelayedNotification struct {
	NotificationID string
	SendAt         *time.Time
	TelegramID     int
	UserID         string
	EventID        string
	BookingID      string
}

// ConfirmRequest is the request body for confirming a booking.
type ConfirmRequest struct {
	BookingID string `json:"booking_id,omitempty" validate:"required,uuid"`
}

// ConfirmResponse is the response body for confirming a booking.
type ConfirmResponse struct {
	Status string `json:"status,omitempty"`
	Error  string `json:"error,omitempty"`
}

// EventDetailsResponse is the response body for getting event details.
type EventDetailsResponse struct {
	Name       string     `json:"name,omitempty"`
	Available  int64      `json:"available"`
	Capacity   int64      `json:"capacity"`
	Date       *time.Time `json:"date,omitempty"`
	PaymentTTL int        `json:"payment_ttl,omitempty"`
	Error      string     `json:"error,omitempty"`
}

// Booking is the value object for a booking cache.
type Booking struct {
	ID              string     `json:"id,omitempty"`
	UserID          string     `json:"user_id,omitempty"`
	EventID         string     `json:"event_id,omitempty"`
	Status          string     `json:"status,omitempty"`
	PaymentDeadline *time.Time `json:"payment_deadline,omitempty" format:"rfc3339"`
	Decremented     bool       `json:"decremented,omitempty"`
}

// Event is the value object for an event.
type Event struct {
	ID         string     `json:"id,omitempty"`
	Name       string     `json:"name,omitempty"`
	Capacity   int64      `json:"capacity,omitempty"`
	Available  int64      `json:"available,omitempty"`
	Date       *time.Time `json:"date,omitempty" format:"rfc3339"`
	PaymentTTL int        `json:"payment_ttl,omitempty"`
}

// BookingStatus is the status of a booking.
const (
	BookingStatusPending   = "pending"
	BookingStatusConfirmed = "confirmed"
	BookingStatusExpired   = "expired"
)

// CancelBookingEvent is the value object for a cancel booking event.
type CancelBookingEvent struct {
	NotificationID string `json:"notification_id"`
	BookingID      string `json:"booking_id"`
	EventID        string `json:"event_id"`
}

// MessageCancelBookingTemplate is the template for the message to cancel a booking.
const (
	MessageCancelBookingTemplate = "Your booking %s for event %s has been canceled due to not paying in time. \nPlease contact support if you believe this is an error."
)
