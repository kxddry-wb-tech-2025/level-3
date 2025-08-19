package domain

import "time"

const (
	// StatusPending is the status that is returned when the task is pending.	
	StatusPending   = "pending"
	// StatusRunning is the status that is returned when the task is running.
	StatusRunning = "running"
	// StatusCompleted is the status that is returned when the task is completed.
	StatusCompleted = "completed"
	// StatusFailed is the status that is returned when the task failed.
	StatusFailed    = "failed"
)

// Task is the main task struct that contains the task's id, status, and creation time.
type Task struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Image is the main image struct that contains the image's url and status.
type Image struct {
	URL    string `json:"url,omitempty"`
	Status string `json:"status"`
}
