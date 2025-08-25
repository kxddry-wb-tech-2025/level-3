package models

import (
	"errors"
	"time"
)

type Item struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Price       int       `json:"price"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

var (
	ErrItemNotFound = errors.New("item not found")
)
