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

// Strategy is the strategy for the database.
var Strategy = retry.Strategy{
	Attempts: 3,
	Delay:    1 * time.Second,
	Backoff:  2,
}

// maxGenerateAttempts is the maximum number of attempts to generate a unique short code.
const maxGenerateAttempts = 10

// Storage is the storage for the URLs.
type Storage struct {
	db *dbpg.DB
}

// New creates a new Storage instance.
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

// Close closes the Storage instance.
func (s *Storage) Close() error {
	for _, db := range s.db.Slaves {
		_ = db.Close()
	}
	return s.db.Master.Close()
}

// SaveURL inserts a new URL with a generated short code.
// If withAlias is true, the URL is inserted with the given alias.
// If withAlias is false, the URL is inserted with a generated short code.
// If the alias already exists, an error is returned.
// If the short code cannot be generated after maxGenerateAttempts, an error is returned.
func (s *Storage) SaveURL(ctx context.Context, url string, withAlias bool, alias string) (string, error) {
	// handle edge cases
	if alias == "" {
		withAlias = false
	}
	const insertQuery = `
		INSERT INTO shortened_urls (url, short_code, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (short_code) DO NOTHING
	`

	if withAlias {
		res, err := s.db.ExecWithRetry(ctx, Strategy, insertQuery, url, alias, time.Now().UTC())
		if err != nil {
			return "", err
		}
		if rows, err := res.RowsAffected(); err == nil && rows == 1 {
			return alias, nil
		}
		return "", errors.New("alias already exists")
	}
	// we are using a cycle here because we want to generate a new short code if the first one is already taken
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
