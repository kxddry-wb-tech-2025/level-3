package domain

import "time"

type ShortenedURL struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	ShortCode string    `json:"short_code"`
	CreatedAt time.Time `json:"created_at"`
}

type Click struct {
	ID        string    `json:"id"`
	ShortCode string    `json:"short_code"`
	UserAgent string    `json:"user_agent"`
	IP        string    `json:"ip"`
	Referer   string    `json:"referer"`
	Country   string    `json:"country,omitempty"`
	City      string    `json:"city,omitempty"`
	Device    string    `json:"device,omitempty"`
	OS        string    `json:"os,omitempty"`
	Browser   string    `json:"browser,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

