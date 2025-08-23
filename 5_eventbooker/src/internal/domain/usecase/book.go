package usecase

import (
	"context"
	"errors"
	"eventbooker/src/internal/domain"
	"time"

	"github.com/kxddry/wbf/zlog"
)

func (u *Usecase) Book(ctx context.Context, eventID string, userID string) domain.BookResponse {
	var bookingID string
	var paymentDeadline time.Time
	err := u.storage.Do(ctx, func(ctx context.Context, tx Tx) error {
		event, err := tx.GetEvent(ctx, eventID)
		if err != nil {
			return err
		}
		if event.Available <= 0 {
			return errors.New("event is full")
		}
		if event.Date.Before(time.Now()) {
			return errors.New("event is in the past")
		}
		paymentDeadline = time.Now().Add(event.PaymentTTL)
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

		if err := u.nf.SendDelayed(notif); err != nil {
			zlog.Logger.Err(err).Msg("failed to send delayed notification")
		} else {
			// but why here?
			// If the notification does not get sent, we cannot increment the available capacity later on.
			// So in that case, we should decrement it in the confirm usecase.
			// How do we do that? I don't know.
			// We should probably store a column in the database to track whether the notification was sent or not.
			// If it was not sent, we can decrement the available capacity on confirm.
			// If it was sent, we decrement it here.
			event.Available--
			// we set decremented to true here because we know that the notification was sent.
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
