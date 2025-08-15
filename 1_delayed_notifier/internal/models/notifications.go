package models

import "time"

const (
	StatusCreated  = "created"  // StatusCreated is used for a newly created notification
	StatusSent     = "sent"     // StatusSent is used for a notification that has finished its life cycle and has been sent to someone
	StatusReceived = "received" // StatusReceived is used for a notification that was received by the Telegram bot but not yet sent.
	StatusCanceled = "canceled" // StatusCanceled is used for a notification that has been canceled by the user.
	StatusFailed   = "failed"   // StatusFailed is used for a notification that has not been delivered due to technical reasons.
)

const (
	ChannelTelegram = "telegram" // ChannelTelegram is used as a way of communication to send notifications.
	// Currently, only telegram is supported.
)

// NotificationCreate is used as a struct for POST requests to disallow users set their own IDs and statuses and etc.
type NotificationCreate struct {
	SendAt    time.Time `json:"send_at"`
	Channel   string    `json:"channel"`
	Recipient string    `json:"recipient"`
	Message   string    `json:"message"`
}

// Notification is used as a struct for notifications.
type Notification struct {
	ID        string    `json:"id"`
	SendAt    time.Time `json:"send_at"`
	Channel   string    `json:"channel"`
	Recipient string    `json:"recipient"`
	Message   string    `json:"message"`
	Attempt   int       `json:"attempt"`
}

// Notification status is used as a status to store in the handlers.StatusStorage and worker.StatusStore templates.
type NotificationStatus struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}
