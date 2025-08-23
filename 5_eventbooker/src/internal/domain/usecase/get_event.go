package usecase

import (
	"context"
	"errors"
	"eventbooker/src/internal/domain"
	"eventbooker/src/internal/storage"
)

func (u *Usecase) GetEvent(ctx context.Context, eventID string) domain.EventDetailsResponse {
	var event domain.Event
	err := u.storage.Do(context.Background(), func(ctx context.Context, tx Tx) error {
		var err error
		event, err = tx.GetEvent(ctx, eventID)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return errors.New("event not found")
			}
			return err
		}
		return nil
	})
	if err != nil {
		return domain.EventDetailsResponse{
			Error: err.Error(),
		}
	}

	return domain.EventDetailsResponse{
		Name:       event.Name,
		Capacity:   event.Capacity,
		Available:  event.Available,
		Date:       event.Date,
		PaymentTTL: event.PaymentTTL,
	}
}
