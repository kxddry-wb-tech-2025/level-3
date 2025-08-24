package usecase

import (
	"context"
	"errors"
	"eventbooker/src/internal/domain"
	"time"

	"github.com/kxddry/wbf/zlog"
)

// Confirm is the set of actions required to run the confirmation process.
// It is responsible for confirming a booking and decrementing the event's available capacity.
func (u *Usecase) Confirm(ctx context.Context, eventID, bookingID string) domain.ConfirmResponse {
	var status string
	var notificationID string
	err := u.storage.Do(ctx, func(ctx context.Context, tx Tx) error {
		// get event
		_, err := tx.GetEvent(ctx, eventID)
		if err != nil {
			return err
		}
		// get booking
		booking, err := tx.GetBooking(ctx, bookingID)
		if err != nil {
			return err
		}
		// check if booking belongs to the event
		if booking.EventID != eventID {
			return errors.New("booking does not belong to the event")
		}

		// check if booking payment deadline has passed
		if booking.PaymentDeadline.Before(time.Now()) {
			return errors.New("booking payment deadline has passed")
		}

		// check if booking is already confirmed or expired
		switch booking.Status {
		case domain.BookingStatusConfirmed:
			return errors.New("booking is already confirmed")
		case domain.BookingStatusExpired:
			return errors.New("booking is expired")
		case domain.BookingStatusPending:
		default:
			zlog.Logger.Err(errors.New("invalid booking status")).Msg("invalid booking status")
			panic(errors.New("invalid booking status in the database"))
		}
		// confirm booking
		status, err = tx.Confirm(ctx, bookingID)
		if err != nil {
			return err
		}

		notificationID, _ = tx.GetNotificationID(ctx, bookingID)

		return nil
	})
	if err != nil {
		return domain.ConfirmResponse{
			Error: err.Error(),
		}
	}

	// cancel delayed notification
	if err := u.nf.CancelNotification(ctx, notificationID); err != nil {
		zlog.Logger.Err(err).Msg("failed to cancel delayed notification")
	}

	return domain.ConfirmResponse{
		Status: status,
	}
}
