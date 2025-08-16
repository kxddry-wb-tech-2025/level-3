package storage

import "errors"

var (
	// ErrInvalidNotification is returned when a notification is invalid.
	ErrInvalidNotification = errors.New("invalid notification")
	// ErrUnknownZSet is returned when a zset kind is unknown.
	ErrUnknownZSet = errors.New("unknown zset kind")
)
