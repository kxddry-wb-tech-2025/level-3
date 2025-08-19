package postgres

import (
	"context"
	"fmt"
	"image-processor/internal/domain"
	"image-processor/internal/helpers"
	"image-processor/internal/storage"
	"time"

	"github.com/kxddry/wbf/dbpg"
	"github.com/kxddry/wbf/zlog"
)

// Storage is the main storage struct that contains the database connection.
type Storage struct {
	db *dbpg.DB
}

func (s *Storage) execAndCheck(ctx context.Context, query string, args ...any) error {

	// I will not be using ExecWithRetry because its implementation is shit. If it gets sql.ErrNoRows, it retries.
	// So instead, we'll use ExecContext and handle the error ourselves.
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return storage.ErrNotFound
	}

	return nil
}

func New(master string, slaves ...string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := dbpg.New(master, slaves, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := db.Master.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	for _, s := range s.db.Slaves {
		_ = s.Close()
	}
	return s.db.Master.Close()
}

func (s *Storage) AddFile(ctx context.Context, file *domain.File) error {
	const op = "storage.postgres.AddFile"

	id, ext := helpers.Split(file.Name)

	if err := s.execAndCheck(ctx, "INSERT INTO images (id, ext, created_at, status) VALUES ($1, $2, $3, $4)", id, ext, time.Now(), domain.StatusPending); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) UpdateStatus(ctx context.Context, fileName string, status string) error {
	const op = "storage.postgres.UpdateStatus"

	id, ext := helpers.Split(fileName)

	if err := s.execAndCheck(ctx, "UPDATE images SET status = $1 WHERE id = $2 AND ext = $3", status, id, ext); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) AddNewID(ctx context.Context, fileName string, newID string) error {
	const op = "storage.postgres.AddNewID"

	id, ext := helpers.Split(fileName)

	if err := s.execAndCheck(ctx, "UPDATE images SET new_id = $1 WHERE id = $2 AND ext = $3", newID, id, ext); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetStatus(ctx context.Context, id string) (string, error) {
	const op = "storage.postgres.GetStatus"

	var status string
	rows, err := s.db.QueryContext(ctx, "SELECT status FROM images WHERE id = $1", id)
	if err != nil {
		zlog.Logger.Err(err).Msg("failed to get status")
		return "", fmt.Errorf("%s: %w", op, err)
	}

	defer rows.Close()

	if !rows.Next() {
		return "", fmt.Errorf("%s: %w", op, storage.ErrNotFound)
	}

	err = rows.Scan(&status)

	if err != nil {
		zlog.Logger.Err(err).Msg("failed to scan status")
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return status, nil
}

func (s *Storage) GetFileName(ctx context.Context, id string) (string, error) {
	const op = "storage.postgres.GetFileName"

	var ext string
	rows, err := s.db.QueryContext(ctx, "SELECT ext FROM images WHERE id = $1", id)
	if err != nil {
		zlog.Logger.Err(err).Msg("failed to get file name")
		return "", fmt.Errorf("%s: %w", op, err)
	}

	defer rows.Close()

	if !rows.Next() {
		return "", fmt.Errorf("%s: %w", op, storage.ErrNotFound)
	}

	err = rows.Scan(&ext)
	if err != nil {
		zlog.Logger.Err(err).Msg("failed to scan file name")
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id + "." + ext, nil
}

func (s *Storage) DeleteFile(ctx context.Context, id string) error {
	const op = "storage.postgres.DeleteFile"

	if err := s.execAndCheck(ctx, "DELETE FROM images WHERE id = $1", id); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
