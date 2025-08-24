package client

import "time"

// SendNotificationRequest is the request body for sending a notification.
type SendNotificationRequest struct {
	SendAt    *time.Time `json:"send_at"`
	Channel   string     `json:"channel"`
	Recipient string     `json:"recipient"`
	Message   string     `json:"message"`
}

// SendNotificationResponse is the response body for sending a notification.
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

// CancelNotificationRequest is the request body for canceling a notification.
type CancelNotificationRequest struct {
	ID string `json:"id"`
}

// CancelSuccess is the status code for a successful cancellation.
const (
	CancelSuccess       = 204 // No Content
	CancelNotFound      = 404 // Not Found
	CancelInternalError = 500 // Internal Server Error
)

// SendSuccess is the status code for a successful send.
const (
	SendSuccess       = 202 // Accepted
	SendInternalError = 500 // Internal Server Error
	SendBadRequest    = 400 // Bad Request
)
