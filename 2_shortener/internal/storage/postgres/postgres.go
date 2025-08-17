package postgres

import (
	"context"
	"database/sql"
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

func New(db *dbpg.DB) (*Storage, error) {
	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	for _, db := range s.db.Slaves {
		_ = db.Close()
	}
	return s.db.Master.Close()
}

func (s *Storage) SaveURL(ctx context.Context, url string) (string, error) {
	tx, err := s.db.Master.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	shortCode := uuid.New().String()[:6]
	now := time.Now().UTC()

	query := `
		INSERT INTO shortened_urls (url, short_code, created_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	if err := tx.QueryRowContext(ctx, query, url, shortCode, now).Scan(&shortCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = retry.Do(func() error {
				shortCode = uuid.New().String()[:6]
				return tx.QueryRowContext(ctx, query, url, shortCode, now).Scan(&shortCode)
			}, retry.Strategy{
				Attempts: 3,
				Delay:    1 * time.Second,
				Backoff:  2,
			})
			if err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}
	return shortCode, nil
}

func (s *Storage) GetURL(ctx context.Context, shortCode string) (string, error) {
	query := `
		SELECT url FROM shortened_urls WHERE short_code = $1
	`
	var url string
	rows, err := s.db.QueryContext(ctx, query, shortCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrNotFound
		}
		return "", err
	}
	if !rows.Next() {
		return "", storage.ErrNotFound
	}
	if err := rows.Scan(&url); err != nil {
		return "", err
	}
	return url, nil
}
