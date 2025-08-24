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

func (r *EventRepository) Close() error {
	for _, slave := range r.db.Slaves {
		_ = slave.Close()
	}
	return r.db.Master.Close()
}

/*
CREATE TABLE events (
	pk BIGSERIAL PRIMARY KEY,
	id UUID NOT NULL UNIQUE,
	name VARCHAR(255) NOT NULL,
	capacity BIGINT NOT NULL,
	date TIMESTAMPTZ NOT NULL,
	payment_ttl INTEGER NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
)
*/

func (r *EventRepository) CreateEvent(ctx context.Context, event domain.CreateEventRequest) (string, error) {
	var id string
	if tx, ok := storage.TxFromContext(ctx); ok {
		err := tx.QueryRowContext(ctx, `INSERT INTO events (name, capacity, available, date, payment_ttl) VALUES ($1, $2, $3, $4, $5) RETURNING id`, event.Name, event.Capacity, event.Capacity, event.Date, event.PaymentTTL).Scan(&id)
		if err != nil {
			return "", err
		}
		return id, nil
	}
	err := r.db.Master.QueryRowContext(ctx, `INSERT INTO events (name, capacity, available, date, payment_ttl) VALUES ($1, $2, $3, $4, $5) RETURNING id`, event.Name, event.Capacity, event.Capacity, event.Date, event.PaymentTTL).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *EventRepository) UpdateEvent(ctx context.Context, event domain.Event) error {
	if tx, ok := storage.TxFromContext(ctx); ok {
		res, err := tx.ExecContext(ctx, `UPDATE events SET name = $1, capacity = $2, available = $3, date = $4, payment_ttl = $5 WHERE id = $6`, event.Name, event.Capacity, event.Available, event.Date, event.PaymentTTL, event.ID)
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
	res, err := r.db.ExecContext(ctx, `UPDATE events SET name = $1, capacity = $2, available = $3, date = $4, payment_ttl = $5 WHERE id = $6`, event.Name, event.Capacity, event.Available, event.Date, event.PaymentTTL, event.ID)
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
	var rows *sql.Rows
	var err error
	if tx, ok := storage.TxFromContext(ctx); ok {
		rows, err = tx.QueryContext(ctx, `SELECT name, capacity, available, date, payment_ttl FROM events WHERE id = $1`, id)
	} else {
		rows, err = r.db.QueryContext(ctx, `SELECT name, capacity, available, date, payment_ttl FROM events WHERE id = $1`, id)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Event{}, storage.ErrEventNotFound
		}
		return domain.Event{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return domain.Event{}, storage.ErrEventNotFound
	}
	var event domain.Event
	if err = rows.Scan(&event.Name, &event.Capacity, &event.Available, &event.Date, &event.PaymentTTL); err != nil {
		return domain.Event{}, err
	}
	event.ID = id
	return event, nil
}
