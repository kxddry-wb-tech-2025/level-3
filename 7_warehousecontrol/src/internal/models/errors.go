package models

import "errors"

var (
	ErrItemNotFound = errors.New("item not found")
	ErrUserNotFound = errors.New("user not found")
	ErrForbidden    = errors.New("forbidden")
)
