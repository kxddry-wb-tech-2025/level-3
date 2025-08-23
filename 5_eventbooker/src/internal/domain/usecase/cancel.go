package usecase

import (
	"context"
	"errors"
	"eventbooker/src/internal/domain"

	"github.com/kxddry/wbf/zlog"
)

// HandleCancellations is the set of actions required to run the cancellation process.
// It is responsible for cancelling a booking and incrementing the event's available capacity.
func (u *Usecase) HandleCancellations(ctx context.Context) error {
	events, err := u.cs.Cancellations(ctx)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-events:
				if !ok {
					return
				}
				if err := u.cancelBooking(ctx, event.BookingID); err != nil {
					zlog.Logger.Err(err).Msg("failed to cancel booking")
				}
			}
		}
	}()
	return nil
}

// cancelBooking is the set of actions required to cancel a booking.
func (u *Usecase) cancelBooking(ctx context.Context, bookingID string) error {
	return u.storage.Do(ctx, func(ctx context.Context, tx Tx) error {
		// get booking
		booking, err := tx.GetBooking(ctx, bookingID)
		if err != nil {
			return err
		}
		// check if booking is expired
		if booking.Status == domain.BookingStatusExpired {
			return errors.New("booking is already expired")
		}
		// check if booking is confirmed
		if booking.Status == domain.BookingStatusConfirmed {
			return errors.New("booking is confirmed")
		}

		// if the booking was decremented, we need to increment the available capacity back
		if booking.Decremented {
			event, err := tx.GetEvent(ctx, booking.EventID)
			if err != nil {
				return err
			}
			// increment event available capacity
			event.Available++
			if err = tx.UpdateEvent(ctx, event); err != nil {
				return err
			}
		}

		// set booking decremented to false
		if err = tx.BookingSetDecremented(ctx, bookingID, false); err != nil {
			return err
		}

		// set booking status to expired
		if err = tx.BookingSetStatus(ctx, bookingID, domain.BookingStatusExpired); err != nil {
			return err
		}
		return nil
	})
}
