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
	err := u.storage.Do(context.Background(), func(ctx context.Context, tx Tx) error {
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
		return nil
	})
	if err != nil {
		return domain.BookResponse{
			Error: err.Error(),
		}
	}

	notif := domain.DelayedNotification{
		SendAt:     &paymentDeadline,
		TelegramID: userID,
		EventID:    eventID,
		BookingID:  bookingID,
	}

	if err := u.nfs.SendDelayed(notif); err != nil {
		zlog.Logger.Err(err).Msg("failed to send delayed notification")
	}

	return domain.BookResponse{
		ID:              bookingID,
		Status:          domain.BookingStatusPending,
		PaymentDeadline: paymentDeadline,
	}
}
