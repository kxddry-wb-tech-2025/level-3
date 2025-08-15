package models

import "time"

const (
	StatusCreated = iota + 1
	StatusSent
	StatusReceived
	StatusCanceled
)

const (
	ChannelEmail = iota + 1
	ChannelTelegram
)

type Notification struct {
	ID        int64     `json:"id,omitempty"`
	UserID    int64     `json:"user_id"`
	ChannelID int64     `json:"channel_id"`
	SendAt    time.Time `json:"send_at"`
	Payload   payload   `json:"payload"`
}

type payload struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}
