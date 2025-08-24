package usecase

import (
	"context"
	"errors"
	"eventbooker/src/internal/domain"
)

// HandleCancellations is the set of actions required to run the cancellation process.
// It is responsible for cancelling a booking and incrementing the event's available capacity.
func (u *Usecase) HandleCancellations(ctx context.Context) {
	events := u.cs.Messages(ctx)

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
					u.log.Err(err).Msg("failed to cancel booking")
				}
			}
		}
	}()
}

// cancelBooking is the set of actions required to cancel a booking.
func (u *Usecase) cancelBooking(ctx context.Context, bookingID string) error {
	return u.storage.Do(ctx, func(ctx context.Context, tx Tx) error {
		// get booking
		booking, err := tx.GetBooking(ctx, bookingID)
		if err != nil {
			u.log.Error().Err(err).Msg("failed to get booking")
			return err
		}
		// check if booking is expired
		if booking.Status == domain.BookingStatusExpired {
			u.log.Error().Msg("booking is already expired")
			return errors.New("booking is already expired")
		}
		// check if booking is confirmed
		if booking.Status == domain.BookingStatusConfirmed {
			u.log.Error().Msg("booking is confirmed")
			return errors.New("booking is confirmed")
		}

		// if the booking was decremented, we need to increment the available capacity back
		if booking.Decremented {
			event, err := tx.GetEvent(ctx, booking.EventID)
			if err != nil {
				u.log.Error().Err(err).Msg("failed to get event")
				return err
			}
			// increment event available capacity
			event.Available++
			if err = tx.UpdateEvent(ctx, event); err != nil {
				u.log.Error().Err(err).Msg("failed to update event")
				return err
			}
		}

		// set booking status to expired
		if err = tx.BookingSetStatus(ctx, bookingID, domain.BookingStatusExpired); err != nil {
			u.log.Error().Err(err).Msg("failed to set booking status")
			return err
		}
		// set booking decremented to false
		if err = tx.BookingSetDecremented(ctx, bookingID, false); err != nil {
			u.log.Error().Err(err).Msg("failed to set booking decremented")
			return err
		}

		return nil
	})
}
