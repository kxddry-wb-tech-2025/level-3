package client

import "time"

type SendNotificationRequest struct {
	SendAt    *time.Time `json:"send_at"`
	Channel   string     `json:"channel"`
	Recipient string     `json:"recipient"`
	Message   string     `json:"message"`
}

type SendNotificationResponse struct {
	ID         string    `json:"id"`
	Channel    string    `json:"channel"`
	Recipient  string    `json:"recipient"`
	Message    string    `json:"message"`
	SendAt     time.Time `json:"send_at"`
	Status     string    `json:"status"`
	RetryCount int       `json:"retry_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CancelNotificationRequest struct {
	ID string `json:"id"`
}

const (
	CancelSuccess       = 204 // No Content
	CancelNotFound      = 404 // Not Found
	CancelInternalError = 500 // Internal Server Error
)

const (
	SendSuccess       = 202 // Accepted
	SendInternalError = 500 // Internal Server Error
	SendBadRequest    = 400 // Bad Request
)
