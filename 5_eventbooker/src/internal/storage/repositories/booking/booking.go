package booking

import (
	"context"
	"database/sql"
	"errors"
	"eventbooker/src/internal/domain"
	"eventbooker/src/internal/storage"
	"time"

	"github.com/kxddry/wbf/dbpg"
)

type BookingRepository struct {
	db *dbpg.DB
}

func New(masterDSN string, slaveDSNs ...string) (*BookingRepository, error) {
	db, err := dbpg.New(masterDSN, slaveDSNs, nil)
	if err != nil {
		return nil, err
	}
	return &BookingRepository{db: db}, nil
}

/*

CREATE TABLE bookings (
	id UUID PRIMARY KEY,
	event_id UUID NOT NULL,
	user_id UUID NOT NULL,
	decremented BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	payment_deadline TIMESTAMP NOT NULL,
)
*/

func (r *BookingRepository) Book(ctx context.Context, eventID, userID string, paymentDeadline time.Time, decremented bool) (string, error) {
	var id string
	if tx, ok := storage.TxFromContext(ctx); ok {
		err := tx.QueryRowContext(ctx, `INSERT INTO bookings (event_id, user_id, payment_deadline, decremented) VALUES ($1, $2, $3, $4) RETURNING id`, eventID, userID, paymentDeadline, decremented).Scan(&id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return "", storage.ErrEventNotFound
			}
			return "", err
		}
		return id, nil
	}
	err := r.db.Master.QueryRowContext(ctx, `INSERT INTO bookings (event_id, user_id, payment_deadline, decremented) VALUES ($1, $2, $3, $4) RETURNING id`, eventID, userID, paymentDeadline, decremented).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrEventNotFound
		}
		return "", err
	}
	return id, nil
}

func (r *BookingRepository) BookingSetDecremented(ctx context.Context, bookingID string, decremented bool) error {
	if tx, ok := storage.TxFromContext(ctx); ok {
		res, err := tx.ExecContext(ctx, `UPDATE bookings SET decremented = $1 WHERE id = $2`, decremented, bookingID)
		if err != nil {
			return err
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return storage.ErrBookingNotFound
		}
		return nil
	}
	res, err := r.db.ExecContext(ctx, `UPDATE bookings SET decremented = $1 WHERE id = $2`, decremented, bookingID)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return storage.ErrBookingNotFound
	}
	return nil
}

func (r *BookingRepository) BookingSetStatus(ctx context.Context, bookingID string, status string) error {
	if tx, ok := storage.TxFromContext(ctx); ok {
		res, err := tx.ExecContext(ctx, `UPDATE bookings SET status = $1 WHERE id = $2`, status, bookingID)
		if err != nil {
			return err
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return storage.ErrBookingNotFound
		}
		return nil
	}

	res, err := r.db.ExecContext(ctx, `UPDATE bookings SET status = $1 WHERE id = $2`, status, bookingID)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return storage.ErrBookingNotFound
	}
	return nil
}

func (r *BookingRepository) GetBooking(ctx context.Context, bookingID string) (domain.Booking, error) {
	var b domain.Booking
	if tx, ok := storage.TxFromContext(ctx); ok {
		err := tx.QueryRowContext(ctx, `SELECT event_id, user_id, payment_deadline FROM bookings WHERE id = $1`, bookingID).Scan(&b.EventID, &b.UserID, &b.PaymentDeadline)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.Booking{}, storage.ErrBookingNotFound
			}
			return domain.Booking{}, err
		}
		b.ID = bookingID
		return b, nil
	}
	err := r.db.Master.QueryRowContext(ctx, `SELECT event_id, user_id, payment_deadline FROM bookings WHERE id = $1`, bookingID).Scan(&b.EventID, &b.UserID, &b.PaymentDeadline)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Booking{}, storage.ErrBookingNotFound
		}
		return domain.Booking{}, err
	}
	b.ID = bookingID
	return b, nil
}

func (r *BookingRepository) Confirm(ctx context.Context, bookingID string) (string, error) {
	var status string
	if tx, ok := storage.TxFromContext(ctx); ok {
		res, err := tx.ExecContext(ctx, `UPDATE bookings SET status = $1 WHERE id = $2`, domain.BookingStatusConfirmed, bookingID)
		if err != nil {
			return "", err
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return "", err
		}
		if rowsAffected == 0 {
			return "", storage.ErrBookingNotFound
		}
		return status, nil
	}
	res, err := r.db.ExecContext(ctx, `UPDATE bookings SET status = $1 WHERE id = $2`, domain.BookingStatusConfirmed, bookingID)
	if err != nil {
		return "", err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return "", err
	}
	if rowsAffected == 0 {
		return "", storage.ErrBookingNotFound
	}
	return status, nil
}
