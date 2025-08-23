package notifications

import (
	"context"
	"eventbooker/src/internal/domain"
	"eventbooker/src/internal/storage"

	"github.com/kxddry/wbf/dbpg"
)

// NotificationRepository is the repository for the notifications.
type NotificationRepository struct {
	db *dbpg.DB
}

// New is the constructor for the NotificationRepository.
// It is responsible for creating a new NotificationRepository.
func New(masterDSN string, slaveDSNs ...string) (*NotificationRepository, error) {
	db, err := dbpg.New(masterDSN, slaveDSNs, nil)
	if err != nil {
		return nil, err
	}
	return &NotificationRepository{db: db}, nil
}

/*
CREATE TABLE notifications (
	id UUID NOT NULL UNIQUE,
	user_id UUID NOT NULL,
	event_id UUID NOT NULL,
	booking_id UUID NOT NULL,
	send_at TIMESTAMPTZ NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
)
*/

// Add is the method for adding a notification to the database.
func (r *NotificationRepository) Add(ctx context.Context, notif domain.DelayedNotification) error {
	if tx, ok := storage.TxFromContext(ctx); ok {
		if err := tx.QueryRowContext(ctx, `INSERT INTO notifications (id, user_id, event_id, booking_id, send_at) VALUES ($1, $2, $3, $4, $5)`,
			notif.NotificationID, notif.TelegramID, notif.EventID, notif.BookingID, notif.SendAt).Err(); err != nil {
			return err
		}
		return nil
	}
	err := r.db.Master.QueryRowContext(ctx, `INSERT INTO notifications (id, user_id, event_id, booking_id, send_at) VALUES ($1, $2, $3, $4, $5)`,
		notif.NotificationID, notif.TelegramID, notif.EventID, notif.BookingID, notif.SendAt).Err()
	if err != nil {
		return err
	}
	return nil
}
