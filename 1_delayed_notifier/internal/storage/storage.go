package storage

import "errors"

var (
	ErrInvalidNotification = errors.New("invalid notification")
	ErrUnknownZSet         = errors.New("unknown zset kind")
)
