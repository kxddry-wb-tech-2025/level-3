package models

import "time"

type PostRequest struct {
	Title       string    `json:"title" validate:"required"`
	Price       int       `json:"price" validate:"required,min=0"`
	Description string    `json:"description" validate:"required"`
	Date        time.Time `json:"date" validate:"required"`
	Category    string    `json:"category" validate:"required"`
}

type PostResponse Item
