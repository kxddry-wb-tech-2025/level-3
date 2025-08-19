package storage

import "errors"

// ErrNotFound is the error that is returned when a file is not found.
var (
	ErrNotFound = errors.New("not found")
)
