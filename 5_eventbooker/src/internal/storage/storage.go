package storage

import (
	"context"
	"database/sql"
	"errors"
)

var (
	// ErrNotFound is returned when a resource is not found.
	ErrNotFound = errors.New("not found")
	// ErrEventNotFound is returned when an event is not found.
	ErrEventNotFound = errors.New("event not found")
	// ErrBookingNotFound is returned when a booking is not found.
	ErrBookingNotFound = errors.New("booking not found")
)

// txKey is an unexported context key for transactions.
type txKey struct{}

// WithTx attaches a sql.Tx to the context.
func WithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// TxFromContext retrieves a sql.Tx from context if present.
func TxFromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	return tx, ok
}
