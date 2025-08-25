package storage

import (
	"context"
	"database/sql"
	"errors"
)

var (
	// ErrNotFound is returned when a resource is not found.
	ErrNotFound = errors.New("not found")
)

// txKey is an unexported context key for transactions.
type txKey struct{}

// WithTx attaches a sql.Tx to the context.
//
// It returns a new context with the transaction attached.
func WithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// TxFromContext retrieves a sql.Tx from context if present.
//
// It returns the transaction and a boolean indicating if it was found.
func TxFromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	return tx, ok
}
