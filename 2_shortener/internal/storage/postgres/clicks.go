package postgres

import (
	"context"
	"time"

	"shortener/internal/domain"

	"github.com/google/uuid"
)

// SaveClick stores a click event for a short code.
func (s *Storage) SaveClick(ctx context.Context, c domain.Click) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	if c.Timestamp.IsZero() {
		c.Timestamp = time.Now().UTC()
	}

	const insert = `
		INSERT INTO clicks (
			id, short_code, user_agent, ip, referer, timestamp
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)`

	_, err := s.db.ExecWithRetry(
		ctx,
		Strategy,
		insert,
		c.ID, c.ShortCode, c.UserAgent, c.IP, c.Referer, c.Timestamp,
	)
	return err
}

// GetClicks returns a paginated list of clicks for the given short code.
func (s *Storage) GetClicks(ctx context.Context, shortCode string, limit, offset int) ([]domain.Click, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	const q = `
		SELECT id, short_code, user_agent, ip, referer, timestamp
		FROM clicks
		WHERE short_code = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryWithRetry(ctx, Strategy, q, shortCode, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Click
	for rows.Next() {
		var c domain.Click
		if err := rows.Scan(&c.ID, &c.ShortCode, &c.UserAgent, &c.IP, &c.Referer, &c.Timestamp); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ClickCount returns the total number of clicks for the given short code.
func (s *Storage) ClickCount(ctx context.Context, shortCode string) (int64, error) {
	const q = `SELECT COUNT(*) FROM clicks WHERE short_code = $1`
	r, err := s.db.QueryWithRetry(ctx, Strategy, q, shortCode)
	if err != nil {
		return 0, err
	}
	defer r.Close()
	if !r.Next() {
		return 0, nil
	}
	var count int64
	if err := r.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// UniqueClickCount returns the number of unique IPs for the given short code.
func (s *Storage) UniqueClickCount(ctx context.Context, shortCode string) (int64, error) {
	const q = `SELECT COUNT(DISTINCT ip) FROM clicks WHERE short_code = $1`
	r, err := s.db.QueryWithRetry(ctx, Strategy, q, shortCode)
	if err != nil {
		return 0, err
	}
	defer r.Close()
	if !r.Next() {
		return 0, nil
	}
	var count int64
	if err := r.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// ClicksByDay aggregates click counts by day for the given short code and optional time range.
func (s *Storage) ClicksByDay(ctx context.Context, shortCode string, start, end *time.Time) (map[string]int64, error) {
	base := `SELECT TO_CHAR(timestamp, 'YYYY-MM-DD') AS day, COUNT(*)
		FROM clicks WHERE short_code = $1`
	args := []any{shortCode}
	idx := 2
	if start != nil && !start.IsZero() {
		base += ` AND timestamp >= $` + itoa(idx)
		args = append(args, start)
		idx++
	}
	if end != nil && !end.IsZero() {
		base += ` AND timestamp <= $` + itoa(idx)
		args = append(args, end)
		idx++
	}
	base += ` GROUP BY day ORDER BY day`

	r, err := s.db.QueryWithRetry(ctx, Strategy, base, args...)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	res := make(map[string]int64)
	for r.Next() {
		var k string
		var v int64
		if err := r.Scan(&k, &v); err != nil {
			return nil, err
		}
		res[k] = v
	}
	return res, r.Err()
}

// ClicksByMonth aggregates click counts by month for the given short code and optional time range.
func (s *Storage) ClicksByMonth(ctx context.Context, shortCode string, start, end *time.Time) (map[string]int64, error) {
	base := `SELECT TO_CHAR(date_trunc('month', timestamp), 'YYYY-MM') AS month, COUNT(*)
		FROM clicks WHERE short_code = $1`
	args := []any{shortCode}
	idx := 2
	if start != nil && !start.IsZero() {
		base += ` AND timestamp >= $` + itoa(idx)
		args = append(args, start)
		idx++
	}
	if end != nil && !end.IsZero() {
		base += ` AND timestamp <= $` + itoa(idx)
		args = append(args, end)
		idx++
	}
	base += ` GROUP BY month ORDER BY month`

	r, err := s.db.QueryWithRetry(ctx, Strategy, base, args...)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	res := make(map[string]int64)
	for r.Next() {
		var k string
		var v int64
		if err := r.Scan(&k, &v); err != nil {
			return nil, err
		}
		res[k] = v
	}
	return res, r.Err()
}

// ClicksByUserAgent aggregates click counts by user agent; limited to top N.
func (s *Storage) ClicksByUserAgent(ctx context.Context, shortCode string, start, end *time.Time, limit int) (map[string]int64, error) {
	if limit <= 0 {
		limit = 10
	}
	base := `SELECT user_agent, COUNT(*) FROM clicks WHERE short_code = $1`
	args := []any{shortCode}
	idx := 2
	if start != nil && !start.IsZero() {
		base += ` AND timestamp >= $` + itoa(idx)
		args = append(args, start)
		idx++
	}
	if end != nil && !end.IsZero() {
		base += ` AND timestamp <= $` + itoa(idx)
		args = append(args, end)
		idx++
	}
	base += ` GROUP BY user_agent ORDER BY COUNT(*) DESC LIMIT $` + itoa(idx)
	args = append(args, limit)

	r, err := s.db.QueryWithRetry(ctx, Strategy, base, args...)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	res := make(map[string]int64)
	for r.Next() {
		var k string
		var v int64
		if err := r.Scan(&k, &v); err != nil {
			return nil, err
		}
		res[k] = v
	}
	return res, r.Err()
}

// Analytics builds a composite analytics response for a short code and optional time range.
func (s *Storage) Analytics(ctx context.Context, shortCode string, from, to *time.Time, topLimit int) (domain.AnalyticsResponse, error) {
	var start, end *time.Time
	if from != nil && !from.IsZero() {
		start = from
	}
	if to != nil && !to.IsZero() {
		end = to
	}

	total, err := s.ClickCount(ctx, shortCode)
	if err != nil {
		return domain.AnalyticsResponse{}, err
	}
	unique, err := s.UniqueClickCount(ctx, shortCode)
	if err != nil {
		return domain.AnalyticsResponse{}, err
	}
	byDay, err := s.ClicksByDay(ctx, shortCode, start, end)
	if err != nil {
		return domain.AnalyticsResponse{}, err
	}
	byMonth, err := s.ClicksByMonth(ctx, shortCode, start, end)
	if err != nil {
		return domain.AnalyticsResponse{}, err
	}
	ua, err := s.ClicksByUserAgent(ctx, shortCode, start, end, topLimit)
	if err != nil {
		return domain.AnalyticsResponse{}, err
	}

	return domain.AnalyticsResponse{
		ShortCode:     shortCode,
		TotalClicks:   total,
		UniqueClicks:  unique,
		ClicksByDay:   byDay,
		ClicksByMonth: byMonth,
		TopUserAgent:  ua,
		From:          start,
		To:            end,
	}, nil
}

// itoa converts an int to string without importing strconv to keep deps minimal here.
func itoa(i int) string {
	if i < 10 {
		return string('0' + byte(i))
	}
	return fmtInt(i)
}

func fmtInt(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
