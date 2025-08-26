package models

import "errors"

// Errors

var (
	// ErrItemNotFound is the error for when an item is not found.
	ErrItemNotFound = errors.New("item not found")
	// ErrUserNotFound is the error for when a user is not found.
	ErrUserNotFound = errors.New("user not found")
	// ErrForbidden is the error for when a user is forbidden from accessing a resource.
	ErrForbidden = errors.New("forbidden")
)
