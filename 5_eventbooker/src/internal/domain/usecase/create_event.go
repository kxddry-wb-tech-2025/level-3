package usecase

import (
	"context"
	"eventbooker/src/internal/domain"
	"time"
)

func (u *Usecase) CreateEvent(ctx context.Context, event domain.CreateEventRequest) domain.CreateEventResponse {
	if !event.Date.After(time.Now()) {
		return domain.CreateEventResponse{
			Error: "event date must be in the future",
		}
	}

	var id string
	err := u.storage.Do(context.Background(), func(ctx context.Context, tx Tx) error {
		var err error
		id, err = tx.CreateEvent(ctx, event)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return domain.CreateEventResponse{
			Error: err.Error(),
		}
	}

	return domain.CreateEventResponse{
		ID: id,
	}
}
