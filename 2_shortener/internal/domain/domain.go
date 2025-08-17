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
	Timestamp time.Time `json:"timestamp"`
}

type ShortenRequest struct {
	URL string `json:"url" validate:"required,url"`
}

// AnalyticsResponse aggregates analytics for a short code over an optional time range.
type AnalyticsResponse struct {
	ShortCode     string           `json:"short_code"`
	From          time.Time       `json:"from,omitempty"`
	To            time.Time       `json:"to,omitempty"`
	TotalClicks   int64            `json:"total_clicks"`
	UniqueClicks  int64            `json:"unique_clicks"`
	ClicksByDay   map[string]int64 `json:"clicks_by_day,omitempty"`
	ClicksByMonth map[string]int64 `json:"clicks_by_month,omitempty"`
	TopUserAgent  map[string]int64 `json:"top_user_agent,omitempty"`
}
