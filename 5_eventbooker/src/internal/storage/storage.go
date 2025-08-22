package storage

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrEventNotFound   = errors.New("event not found")
	ErrBookingNotFound = errors.New("booking not found")
)
