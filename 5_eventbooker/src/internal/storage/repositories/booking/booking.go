package booking

import (
	"eventbooker/src/internal/domain"

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

func (r *BookingRepository) Book(eventID, userID string) (domain.Booking, error) {
	panic("not implemented")
}

func (r *BookingRepository) GetBooking(bookingID string) (domain.Booking, error) {
	panic("not implemented")
}

func (r *BookingRepository) Confirm(bookingID string) (string, error) {
	panic("not implemented")
}
