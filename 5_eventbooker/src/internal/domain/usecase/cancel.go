package usecase

import (
	"context"
	"errors"
	"eventbooker/src/internal/domain"

	"github.com/kxddry/wbf/zlog"
)

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

func (u *Usecase) cancelBooking(ctx context.Context, bookingID string) error {
	return u.storage.Do(ctx, func(ctx context.Context, tx Tx) error {
		booking, err := tx.GetBooking(ctx, bookingID)
		if err != nil {
			return err
		}
		if booking.Status == domain.BookingStatusCancelled {
			return errors.New("booking is already cancelled")
		}
		if booking.Status == domain.BookingStatusExpired {
			return errors.New("booking is expired")
		}
		if booking.Status == domain.BookingStatusConfirmed {
			return errors.New("booking is confirmed")
		}

		// if the booking was decremented, we need to increment the available capacity back.
		if booking.Decremented {
			event, err := tx.GetEvent(ctx, booking.EventID)
			if err != nil {
				return err
			}
			event.Available++
			if err = tx.UpdateEvent(ctx, event); err != nil {
				return err
			}
		}

		if err = tx.BookingSetDecremented(ctx, bookingID, false); err != nil {
			return err
		}

		if err = tx.BookingSetStatus(ctx, bookingID, domain.BookingStatusCancelled); err != nil {
			return err
		}
		return nil
	})
}
