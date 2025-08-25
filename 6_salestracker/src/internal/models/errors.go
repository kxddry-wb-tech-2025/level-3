package models

import "errors"

var (
	ErrItemNotFound = errors.New("item not found")
	ErrInvalidDate  = errors.New("invalid date")
)
