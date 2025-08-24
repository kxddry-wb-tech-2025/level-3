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
func (u *Usecase) Book(ctx context.Context, eventID, userID string, telegramID int) domain.BookResponse {
	var bookingID string
	var paymentDeadline time.Time
	// begin transaction
	if err := u.storage.Do(ctx, func(ctx context.Context, tx Tx) error {
		// get event
		event, err := tx.GetEvent(ctx, eventID)
		if err != nil {
			u.log.Error().Err(err).Msg("failed to get event")
			return err
		}
		// check if event is full
		if event.Available <= 0 {
			u.log.Error().Msg("event is full")
			return errors.New("event is full")
		}
		// check if event is in the past
		if event.Date != nil && event.Date.Before(time.Now().UTC()) {
			u.log.Error().Msg("event is in the past")
			return errors.New("event is in the past")
		}
		// set payment deadline
		paymentDeadline = time.Now().Add(time.Duration(event.PaymentTTL) * time.Second)
		// book event
		bookingID, err = tx.Book(ctx, eventID, userID, paymentDeadline)
		if err != nil {
			u.log.Error().Err(err).Msg("failed to book event")
			return err
		}

		notif := domain.DelayedNotification{
			SendAt:     &paymentDeadline,
			TelegramID: telegramID,
			UserID:     userID,
			EventID:    eventID,
			BookingID:  bookingID,
		}

		// send delayed notification
		if notificationID, err := u.nf.SendNotification(ctx, notif); err != nil {
			zlog.Logger.Err(err).Msg("failed to send delayed notification")
		} else {
			event.Available--
			if err = tx.UpdateEvent(ctx, event); err != nil {
				u.log.Error().Err(err).Msg("failed to update event")
				return err
			}
			if err = tx.BookingSetDecremented(ctx, bookingID, true); err != nil {
				u.log.Error().Err(err).Msg("failed to set booking decremented")
				return err
			}
			notif.NotificationID = notificationID
			if err = tx.AddNotification(ctx, notif); err != nil {
				u.log.Error().Err(err).Msg("failed to add notification")
				return err
			}
		}
		return nil
	}); err != nil {
		u.log.Error().Err(err).Msg("failed to book event")
		return domain.BookResponse{
			Error: err.Error(),
		}
	}

	return domain.BookResponse{
		ID:              bookingID,
		Status:          domain.BookingStatusPending,
		PaymentDeadline: &paymentDeadline,
	}
}
