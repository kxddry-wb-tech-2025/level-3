package usecase

import (
	"context"
	"errors"
	"eventbooker/src/internal/domain"
	"eventbooker/src/internal/storage"

	"github.com/kxddry/wbf/zlog"
)

func (u *Usecase) Confirm(eventID, bookingID string) domain.ConfirmResponse {
	var status string
	err := u.storage.Do(context.Background(), func(ctx context.Context, tx Tx) error {
		event, err := tx.GetEvent(ctx, eventID)
		if err != nil {
			if errors.Is(err, storage.ErrEventNotFound) {
				return errors.New("event not found")
			}
			return err
		}
		booking, err := tx.GetBooking(ctx, bookingID)
		if err != nil {
			return err
		}
		if booking.EventID != eventID {
			return errors.New("booking does not belong to the event")
		}
		switch booking.Status {
		case domain.BookingStatusCancelled:
			return errors.New("booking is cancelled")
		case domain.BookingStatusConfirmed:
			return errors.New("booking is already confirmed")
		case domain.BookingStatusExpired:
			return errors.New("booking is expired")
		case domain.BookingStatusPending:
		default:
			zlog.Logger.Err(errors.New("invalid booking status")).Msg("invalid booking status")
			panic(errors.New("invalid booking status in the database"))
		}
		status, err = tx.Confirm(ctx, bookingID)
		if err != nil {
			return err
		}

		// decrement the available capacity on confirmation
		// I know its better to use it during booking, but I'm too lazy to change the architecture completely
		event.Available--
		if err = tx.UpdateEvent(ctx, event); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return domain.ConfirmResponse{
			Error: err.Error(),
		}
	}

	if err := u.nfs.CancelDelayed(bookingID); err != nil {
		zlog.Logger.Err(err).Msg("failed to cancel delayed notification")
	}

	return domain.ConfirmResponse{
		Status: status,
	}
}
