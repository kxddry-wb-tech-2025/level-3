package usecase

import (
	"context"
	"eventbooker/src/internal/domain"
	"time"
)

// CreateEvent is the set of actions required to run the create event process.
// It is responsible for creating an event and returning the event ID.
func (u *Usecase) CreateEvent(ctx context.Context, event domain.CreateEventRequest) domain.CreateEventResponse {
	// check if event date is in the future
	if !event.Date.After(time.Now()) {
		return domain.CreateEventResponse{
			Error: "event date must be in the future",
		}
	}

	// create event
	var id string
	err := u.storage.Do(ctx, func(ctx context.Context, tx Tx) error {
		// create event
		var err error
		id, err = tx.CreateEvent(ctx, event)
		if err != nil {
			return err
		}
		return nil
	})

	// return error if event creation failed
	if err != nil {
		return domain.CreateEventResponse{
			Error: err.Error(),
		}
	}

	// return event ID
	return domain.CreateEventResponse{
		ID: id,
	}
}
