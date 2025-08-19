package domain

import (
	"io"
	"time"
)

const (
	// StatusPending is the status that is returned when the task is pending.
	StatusPending = "pending"
	// StatusRunning is the status that is returned when the task is running.
	StatusRunning = "running"
	// StatusCompleted is the status that is returned when the task is completed.
	StatusCompleted = "completed"
	// StatusFailed is the status that is returned when the task failed.
	StatusFailed = "failed"
)

// Task is the main task struct that contains the task's id, status, and creation time.
type Task struct {
	FileName  string    `json:"file_name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Image is the main image struct that contains the image's url and status.
type Image struct {
	URL    string `json:"url,omitempty"`
	Status string `json:"status"`
}

// File is the main file struct that contains the file's name, data, size, and content type.
type File struct {
	Name        string
	Data        io.ReadSeekCloser
	Size        int64
	ContentType string
}
