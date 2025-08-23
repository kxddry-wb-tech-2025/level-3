package usecase

import (
	"context"
	"errors"
	"eventbooker/src/internal/domain"
	"time"

	"github.com/kxddry/wbf/zlog"
)

// Book is the set of actions required to run the booking process.
// It is responsible for booking an event, sending a delayed notification, and decrementing the event's available capacity.
func (u *Usecase) Book(ctx context.Context, eventID string, userID string) domain.BookResponse {
	var bookingID string
	var paymentDeadline time.Time
	// begin transaction
	err := u.storage.Do(ctx, func(ctx context.Context, tx Tx) error {
		// get event
		event, err := tx.GetEvent(ctx, eventID)
		if err != nil {
			return err
		}
		// check if event is full
		if event.Available <= 0 {
			return errors.New("event is full")
		}
		// check if event is in the past
		if event.Date.Before(time.Now()) {
			return errors.New("event is in the past")
		}
		// set payment deadline
		paymentDeadline = time.Now().Add(event.PaymentTTL)
		// book event
		bookingID, err = tx.Book(ctx, eventID, userID, paymentDeadline)
		if err != nil {
			return err
		}

		notif := domain.DelayedNotification{
			SendAt:     &paymentDeadline,
			TelegramID: userID,
			EventID:    eventID,
			BookingID:  bookingID,
		}

		// send delayed notification
		if err := u.nf.SendDelayed(notif); err != nil {
			zlog.Logger.Err(err).Msg("failed to send delayed notification")
		} else {
			// decrement event available capacity if notification was sent
			event.Available--
			if err = tx.BookingSetDecremented(ctx, bookingID, true); err != nil {
				return err
			}
			// we only update the event afterwards because we prioritize the event availability over the booking decrement.
			if err = tx.UpdateEvent(ctx, event); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return domain.BookResponse{
			Error: err.Error(),
		}
	}

	return domain.BookResponse{
		ID:              bookingID,
		Status:          domain.BookingStatusPending,
		PaymentDeadline: paymentDeadline,
	}
}
