package postgres

import (
	"context"
	"errors"
	"shortener/internal/storage"
	"time"

	"github.com/google/uuid"
	"github.com/kxddry/wbf/dbpg"
	"github.com/kxddry/wbf/retry"
)

var Strategy = retry.Strategy{
	Attempts: 3,
	Delay:    1 * time.Second,
	Backoff:  2,
}

type Storage struct {
	db *dbpg.DB
}

func New(ctx context.Context, master string, slaves []string) (*Storage, error) {
	db, err := dbpg.New(master, slaves, &dbpg.Options{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 1 * time.Hour,
	})
	if err != nil {
		return nil, err
	}
	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	for _, db := range s.db.Slaves {
		_ = db.Close()
	}
	return s.db.Master.Close()
}

// SaveURL inserts a new URL with a generated short code.
// Handles collisions by re-generating codes and retrying with ExecWithRetry.
// Returns the generated short code or an error if the operation fails.
func (s *Storage) SaveURL(ctx context.Context, url string) (string, error) {
	const insertQuery = `
		INSERT INTO shortened_urls (url, short_code, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (short_code) DO NOTHING
	`

	const maxGenerateAttempts = 10
	for range maxGenerateAttempts {
		shortCode := uuid.New().String()[:6]
		now := time.Now().UTC()

		res, err := s.db.ExecWithRetry(ctx, Strategy, insertQuery, url, shortCode, now)
		if err != nil {
			return "", err
		}
		if rows, err := res.RowsAffected(); err == nil && rows == 1 {
			return shortCode, nil
		}
	}
	return "", errors.New("could not generate unique short code after multiple attempts")
}

// GetURL retrieves the original URL for a given short code.
// Handles retries and error propagation.
func (s *Storage) GetURL(ctx context.Context, shortCode string) (string, error) {
	const query = `
		SELECT url FROM shortened_urls WHERE short_code = $1
	`

	rows, err := s.db.QueryWithRetry(ctx, Strategy, query, shortCode)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	if !rows.Next() {
		return "", storage.ErrNotFound
	}

	var url string
	if err := rows.Scan(&url); err != nil {
		return "", err
	}
	return url, nil
}
