package usecase

import (
	"context"
	"eventbooker/src/internal/domain"
)

// GetEvent is the set of actions required to run the get event process.
// It is responsible for getting an event and returning the event details.
func (u *Usecase) GetEvent(ctx context.Context, eventID string) domain.EventDetailsResponse {
	var event domain.Event
	err := u.storage.Do(ctx, func(ctx context.Context, tx Tx) error {
		var err error
		event, err = tx.GetEvent(ctx, eventID)
		if err != nil {
			u.log.Error().Err(err).Msg("failed to get event")
			return err
		}
		return nil
	})
	if err != nil {
		u.log.Error().Err(err).Msg("failed to get event")
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
