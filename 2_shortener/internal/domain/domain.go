package domain

import "time"

// ShortenedURL is the struct for the shortened URL.
type ShortenedURL struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	ShortCode string    `json:"short_code"`
	CreatedAt time.Time `json:"created_at"`
}

// Click is the struct for the click.
type Click struct {
	ID        string    `json:"id"`
	ShortCode string    `json:"short_code"`
	UserAgent string    `json:"user_agent"`
	IP        string    `json:"ip"`
	Referer   string    `json:"referer"`
	Timestamp time.Time `json:"timestamp"`
}

// ShortenRequest is the struct for the shorten request.
type ShortenRequest struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

// AnalyticsResponse is the struct for the analytics response.
type AnalyticsResponse struct {
	ShortCode     string           `json:"short_code"`
	From          *time.Time       `json:"from,omitempty"`
	To            *time.Time       `json:"to,omitempty"`
	TotalClicks   int64            `json:"total_clicks"`
	UniqueClicks  int64            `json:"unique_clicks"`
	ClicksByDay   map[string]int64 `json:"clicks_by_day,omitempty"`
	ClicksByMonth map[string]int64 `json:"clicks_by_month,omitempty"`
	TopUserAgents map[string]int64 `json:"top_user_agents,omitempty"`
	TopReferers   map[string]int64 `json:"top_referers,omitempty"`
	TopIPs        map[string]int64 `json:"top_ips,omitempty"`
}

// MinUsageForCache is the minimum number of clicks required to cache a URL.
const MinUsageForCache = 10