package events

import (
	"context"
	"database/sql"
	"errors"
	"eventbooker/src/internal/domain"
	"eventbooker/src/internal/storage"

	"github.com/kxddry/wbf/dbpg"
)

type EventRepository struct {
	db *dbpg.DB
}

func New(masterDSN string, slaveDSNs ...string) (*EventRepository, error) {
	db, err := dbpg.New(masterDSN, slaveDSNs, nil)
	if err != nil {
		return nil, err
	}
	return &EventRepository{db: db}, nil
}

/*
CREATE TABLE events (
	pk BIGSERIAL PRIMARY KEY,
	id UUID NOT NULL UNIQUE,
	name VARCHAR(255) NOT NULL,
	capacity BIGINT NOT NULL,
	date TIMESTAMPTZ NOT NULL,
	payment_ttl INTERVAL NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
)
*/

func (r *EventRepository) CreateEvent(ctx context.Context, event domain.CreateEventRequest) (string, error) {
	var id string
	err := r.db.Master.QueryRowContext(ctx, `INSERT INTO events (name, capacity, date, payment_ttl) VALUES ($1, $2, $3, $4) RETURNING id`, event.Name, event.Capacity, event.Date, event.PaymentTTL).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *EventRepository) UpdateEvent(ctx context.Context, event domain.Event) error {
	res, err := r.db.ExecContext(ctx, `UPDATE events SET name = $1, capacity = $2, date = $3, payment_ttl = $4 WHERE id = $5`, event.Name, event.Capacity, event.Date, event.PaymentTTL, event.ID)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return storage.ErrEventNotFound
	}
	return nil
}

func (r *EventRepository) GetEvent(ctx context.Context, id string) (domain.Event, error) {
	var event domain.Event
	res, err := r.db.QueryContext(ctx, `SELECT name, capacity, date, payment_ttl FROM events WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Event{}, storage.ErrEventNotFound
		}
		return domain.Event{}, err
	}
	defer res.Close()
	if !res.Next() {
		return domain.Event{}, storage.ErrEventNotFound
	}
	err = res.Scan(&event.Name, &event.Capacity, &event.Date, &event.PaymentTTL)
	if err != nil {
		return domain.Event{}, err
	}
	event.ID = id
	return event, nil
}
