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
	"github.com/kxddry/wbf/zlog"
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
	s := &Storage{db: db}

	go func() {
		for range time.NewTicker(15 * time.Minute).C {
			if err := s.updateSlavesURLs(ctx); err != nil {
				zlog.Logger.Error().Err(err).Msg("failed to update slaves")
			}
		}
	}()
	return s, nil
}

func (s *Storage) Close() error {
	for _, db := range s.db.Slaves {
		_ = db.Close()
	}
	return s.db.Master.Close()
}

func (s *Storage) updateSlavesURLs(ctx context.Context) error {
	query := `
		SELECT url, short_code, created_at 
		FROM shortened_urls 
		WHERE created_at > NOW() - INTERVAL '1 hour'
		ORDER BY created_at DESC
	`

	rows, err := s.db.Master.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Collect all URLs to replicate
	type urlRecord struct {
		url       string
		shortCode string
		createdAt time.Time
	}

	var records []urlRecord
	for rows.Next() {
		var record urlRecord
		if err := rows.Scan(&record.url, &record.shortCode, &record.createdAt); err != nil {
			return err
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	// Replicate to each slave
	for i, slave := range s.db.Slaves {
		for _, record := range records {
			insertQuery := `
				INSERT INTO shortened_urls (url, short_code, created_at)
				VALUES ($1, $2, $3)
				ON CONFLICT (short_code) DO NOTHING
			`
			_, err := slave.ExecContext(ctx, insertQuery, record.url, record.shortCode, record.createdAt)
			if err != nil {
				zlog.Logger.Error().
					Err(err).
					Int("slave_index", i).
					Str("short_code", record.shortCode).
					Msg("failed to replicate URL to slave")
				continue
			}
		}
	}

	zlog.Logger.Info().
		Int("records_count", len(records)).
		Int("slaves_count", len(s.db.Slaves)).
		Msg("completed slave update")

	return nil
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
