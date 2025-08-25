package models

import (
	"time"
)

type Item struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Price       int       `json:"price"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

type Analytics struct {
	Sum          int     `json:"sum"`
	Count        int     `json:"count"`
	Average      float64 `json:"average"`
	Median       float64 `json:"median"`
	Percentile90 float64 `json:"percentile_90"`
}
