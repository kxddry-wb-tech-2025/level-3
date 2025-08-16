package models

import (
	"time"
)

// NotificationStatus represents the current processing state of a notification.
type NotificationStatus string

// Notification status values.
const (
	StatusScheduled NotificationStatus = "scheduled"
	StatusQueued    NotificationStatus = "queued"
	StatusSent      NotificationStatus = "sent"
	StatusFailed    NotificationStatus = "failed"
	StatusRetrying  NotificationStatus = "retrying"
	StatusCancelled NotificationStatus = "cancelled"
)

// Notification is a persisted unit of work for delivering a message to a recipient via a channel at a given time.
type Notification struct {
	ID            string             `json:"id"`
	Channel       string             `json:"channel"`
	Recipient     string             `json:"recipient"`
	Message       string             `json:"message"`
	SendAt        time.Time          `json:"send_at"`
	Status        NotificationStatus `json:"status"`
	RetryCount    int                `json:"retry_count"`
	NextAttemptAt *time.Time         `json:"next_attempt_at,omitempty"`
	LastError     string             `json:"last_error,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}
