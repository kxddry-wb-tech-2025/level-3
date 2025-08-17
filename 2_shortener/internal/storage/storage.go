package storage

import (
	"errors"
)

// ErrNotFound is the error for the not found links.
var (
	ErrNotFound = errors.New("not found")
)
