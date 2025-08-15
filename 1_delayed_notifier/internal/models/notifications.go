package models

import "time"

const (
	StatusCreated  = "created"
	StatusSent     = "sent"
	StatusReceived = "received"
	StatusCanceled = "canceled"
	StatusFailed   = "failed"
)

const (
	ChannelEmail    = "email"
	ChannelTelegram = "telegram"
)

type NotificationCreate struct {
	SendAt    time.Time `json:"send_at"`
	Channel   string    `json:"channel"`
	Recipient string    `json:"recipient"`
	Message   string    `json:"message"`
}

type Notification struct {
	ID        string    `json:"id"`
	SendAt    time.Time `json:"send_at"`
	Channel   string    `json:"channel"`
	Recipient string    `json:"recipient"`
	Message   string    `json:"message"`
	Attempt   int       `json:"attempt"`
}

type NotificationStatus struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}
