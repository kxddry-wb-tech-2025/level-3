package models

import "errors"

var (
	// ErrItemNotFound is returned when an item is not found.
	ErrItemNotFound = errors.New("item not found")
	// ErrInvalidDate is returned when a date is invalid.
	ErrInvalidDate = errors.New("invalid date")
)
