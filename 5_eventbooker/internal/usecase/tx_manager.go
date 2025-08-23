package usecase

import (
	"context"
	"fmt"
	"eventbooker/internal/domain"
)

// TxManager provides transaction management capabilities for the usecase layer
type TxManager struct {
	eventRepo   domain.EventRepository
	userRepo    domain.UserRepository
	bookingRepo domain.BookingRepository
}

// NewTxManager creates a new transaction manager
func NewTxManager(
	eventRepo domain.EventRepository,
	userRepo domain.UserRepository,
	bookingRepo domain.BookingRepository,
) *TxManager {
	return &TxManager{
		eventRepo:   eventRepo,
		userRepo:    userRepo,
		bookingRepo: bookingRepo,
	}
}

// WithTransaction executes a function within a transaction context
func (tm *TxManager) WithTransaction(ctx context.Context, fn func(domain.Transaction) error) error {
	// Check if repositories support transactions
	eventTx, ok := tm.eventRepo.(interface {
		WithTransaction(ctx context.Context, fn func(domain.EventRepository) error) error
	})
	if !ok {
		return fmt.Errorf("event repository does not support transactions")
	}

	userTx, ok := tm.userRepo.(interface {
		WithTransaction(ctx context.Context, fn func(domain.UserRepository) error) error
	})
	if !ok {
		return fmt.Errorf("user repository does not support transactions")
	}

	bookingTx, ok := tm.bookingRepo.(interface {
		WithTransaction(ctx context.Context, fn func(domain.BookingRepository) error) error
	})
	if !ok {
		return fmt.Errorf("booking repository does not support transactions")
	}

	// Execute the transaction function
	return eventTx.WithTransaction(ctx, func(eventRepo domain.EventRepository) error {
		return userTx.WithTransaction(ctx, func(userRepo domain.UserRepository) error {
			return bookingTx.WithTransaction(ctx, func(bookingRepo domain.BookingRepository) error {
				// Create a transaction context that can be used by the function
				tx := &transactionContext{
					eventRepo:   eventRepo,
					userRepo:    userRepo,
					bookingRepo: bookingRepo,
				}
				return fn(tx)
			})
		})
	})
}

// transactionContext implements the Transaction interface and provides access to repositories
type transactionContext struct {
	eventRepo   domain.EventRepository
	userRepo    domain.UserRepository
	bookingRepo domain.BookingRepository
}

// Commit is a no-op for the transaction context as the actual commit is handled by the repository
func (tc *transactionContext) Commit() error {
	// The actual commit is handled by the repository layer
	return nil
}

// Rollback is a no-op for the transaction context as the actual rollback is handled by the repository
func (tc *transactionContext) Rollback() error {
	// The actual rollback is handled by the repository layer
	return nil
}

// EventRepository returns the event repository within the transaction context
func (tc *transactionContext) EventRepository() domain.EventRepository {
	return tc.eventRepo
}

// UserRepository returns the user repository within the transaction context
func (tc *transactionContext) UserRepository() domain.UserRepository {
	return tc.userRepo
}

// BookingRepository returns the booking repository within the transaction context
func (tc *transactionContext) BookingRepository() domain.BookingRepository {
	return tc.bookingRepo
}