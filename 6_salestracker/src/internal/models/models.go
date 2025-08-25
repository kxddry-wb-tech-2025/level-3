package models

import (
	"time"
)

// Item is the model for an item.
type Item struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Price       float64   `json:"price"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	Category    string    `json:"category"`
}

// Analytics is the model for analytics.
type Analytics struct {
	Sum          float64 `json:"sum"`
	Count        int64   `json:"count"`
	Average      float64 `json:"average"`
	Median       float64 `json:"median"`
	Percentile90 float64 `json:"percentile_90"`
}
