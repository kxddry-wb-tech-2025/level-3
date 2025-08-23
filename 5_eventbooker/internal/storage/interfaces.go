package storage

import (
	"context"
	"eventbooker/internal/domain"
)

// EventRepository defines the interface for event storage operations
type EventRepository interface {
	Create(ctx context.Context, event *domain.Event) error
	GetByID(ctx context.Context, id string) (*domain.Event, error)
	GetAll(ctx context.Context) ([]*domain.Event, error)
	Update(ctx context.Context, event *domain.Event) error
	Delete(ctx context.Context, id string) error
	GetAvailableCapacity(ctx context.Context, eventID string) (int, error)
	DecreaseCapacity(ctx context.Context, eventID string) error
	IncreaseCapacity(ctx context.Context, eventID string) error
}

// UserRepository defines the interface for user storage operations
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
}

// BookingRepository defines the interface for booking storage operations
type BookingRepository interface {
	Create(ctx context.Context, booking *domain.Booking) error
	GetByID(ctx context.Context, id string) (*domain.Booking, error)
	GetByUserID(ctx context.Context, userID string) ([]*domain.Booking, error)
	GetByEventID(ctx context.Context, eventID string) ([]*domain.Booking, error)
	Update(ctx context.Context, booking *domain.Booking) error
	Delete(ctx context.Context, id string) error
	GetUserBookingsForEvent(ctx context.Context, userID, eventID string) ([]*domain.Booking, error)
}

// TransactionalEventRepository extends EventRepository with transaction support
type TransactionalEventRepository interface {
	EventRepository
	WithTransaction(ctx context.Context, fn func(EventRepository) error) error
}

// TransactionalUserRepository extends UserRepository with transaction support
type TransactionalUserRepository interface {
	UserRepository
	WithTransaction(ctx context.Context, fn func(UserRepository) error) error
}

// TransactionalBookingRepository extends BookingRepository with transaction support
type TransactionalBookingRepository interface {
	BookingRepository
	WithTransaction(ctx context.Context, fn func(BookingRepository) error) error
}