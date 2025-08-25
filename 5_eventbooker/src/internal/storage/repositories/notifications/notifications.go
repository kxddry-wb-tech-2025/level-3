package notifications

import (
	"context"
	"eventbooker/src/internal/domain"
	"eventbooker/src/internal/storage"

	"github.com/kxddry/wbf/dbpg"
	"github.com/kxddry/wbf/zlog"
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

// Close closes the NotificationRepository.
func (r *NotificationRepository) Close() error {
	for _, slave := range r.db.Slaves {
		_ = slave.Close()
	}
	return r.db.Master.Close()
}

// AddNotification is the method for adding a notification to the database.
func (r *NotificationRepository) AddNotification(ctx context.Context, notif domain.DelayedNotification) error {
	log := zlog.Logger.With().Str("component", "notifications").Logger().With().Str("operation", "AddNotification").Logger()
	if tx, ok := storage.TxFromContext(ctx); ok {
		if err := tx.QueryRowContext(ctx, `INSERT INTO notifications (id, user_id, telegram_id, event_id, booking_id, send_at) VALUES ($1, $2, $3, $4, $5, $6)`,
			notif.NotificationID, notif.UserID, notif.TelegramID, notif.EventID, notif.BookingID, notif.SendAt).Err(); err != nil {
			log.Error().Err(err).Msg("failed to add notification")
			return err
		}
		return nil
	}
	log.Warn().Msg("no transaction found, using master connection")
	err := r.db.Master.QueryRowContext(ctx, `INSERT INTO notifications (id, user_id, telegram_id, event_id, booking_id, send_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		notif.NotificationID, notif.UserID, notif.TelegramID, notif.EventID, notif.BookingID, notif.SendAt).Err()
	if err != nil {
		log.Error().Err(err).Msg("failed to add notification")
		return err
	}
	return nil
}

// GetNotificationID is the method for getting a notification ID from the database by booking ID.
func (r *NotificationRepository) GetNotificationID(ctx context.Context, bookingID string) (string, error) {
	log := zlog.Logger.With().Str("component", "notifications").Logger().With().Str("operation", "GetNotificationID").Logger()
	var id string
	if tx, ok := storage.TxFromContext(ctx); ok {
		if err := tx.QueryRowContext(ctx, `SELECT id FROM notifications WHERE booking_id = $1`, bookingID).Scan(&id); err != nil {
			log.Error().Err(err).Msg("failed to get notification ID")
			return "", err
		}
		return id, nil
	}
	log.Warn().Msg("no transaction found, using master connection")
	err := r.db.Master.QueryRowContext(ctx, `SELECT id FROM notifications WHERE booking_id = $1`, bookingID).Scan(&id)
	if err != nil {
		log.Error().Err(err).Msg("failed to get notification ID")
		return "", err
	}
	return id, nil
}
