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
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	payment_deadline TIMESTAMP NOT NULL,
)
*/

func (r *BookingRepository) Book(ctx context.Context, eventID, userID string, paymentDeadline time.Time) (string, error) {
	var id string
	err := r.db.Master.QueryRowContext(ctx, `INSERT INTO bookings (event_id, user_id, payment_deadline) VALUES ($1, $2, $3) RETURNING id`, eventID, userID, paymentDeadline).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrEventNotFound
		}
		return "", err
	}

	return id, nil
}

func (r *BookingRepository) GetBooking(ctx context.Context, bookingID string) (domain.Booking, error) {
	panic("not implemented")
}

func (r *BookingRepository) Confirm(ctx context.Context, bookingID string) (string, error) {
	panic("not implemented")
}
