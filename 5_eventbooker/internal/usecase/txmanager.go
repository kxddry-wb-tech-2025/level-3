package usecase

import "context"

// TxManager abstracts transactional execution for usecases.
// Implementations should start a transaction, execute the provided function,
// and commit or rollback depending on whether the function returns an error.
// The provided context may be replaced by an implementation-specific context that
// carries the transactional handle.
type TxManager interface {
	// WithinTransaction executes fn inside a transaction boundary.
	// If fn returns an error, the transaction must be rolled back.
	// If fn returns nil, the transaction must be committed.
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}