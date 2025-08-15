package postgres

import (
	"context"
	"delayed-notifier/internal/models"
	"delayed-notifier/internal/storage"
	"time"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

var strat = retry.Strategy{Attempts: 3, Backoff: 2, Delay: 5 * time.Second}

type Storage struct {
	db *dbpg.DB
}

func NewStorage(db *dbpg.DB) *Storage {
	return &Storage{
		db: db,
	}
}

func (s *Storage) Add(ctx context.Context, notify *models.Notification) (int64, error) {
	res, err := s.db.ExecWithRetry(ctx, strat,
		`INSERT INTO notifications (user_id, channel_id, send_at, subject, body) VALUES ($1, $2, $3) RETURNING id`,
		notify.UserID, notify.ChannelID, notify.SendAt, notify.Payload.Subject, notify.Payload.Body)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Cancel cancels a notification
// StatusCanceled = 4 here, if it changes, have to change it here too.
func (s *Storage) Cancel(ctx context.Context, id int64) error {
	res, err := s.db.ExecWithRetry(ctx, strat, `UPDATE notifications SET status = 4 WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return storage.ErrNoSubscription
	}
	return nil
}

func (s *Storage) Get(ctx context.Context, id int64) (*models.Notification, error) {
	rows, err := s.db.QueryWithRetry(ctx, strat, `SELECT user_id, channel_id, send_at, subject, body FROM notifications WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	n := new(models.Notification)
	for rows.Next() {
		err = rows.Scan(&n.UserID, &n.ChannelID, &n.SendAt, &n.Payload.Subject, &n.Payload.Body)
		if err != nil {
			return nil, err
		}
		break
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return n, nil
}
