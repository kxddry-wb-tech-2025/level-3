package postgres

import (
	"eventbooker/src/internal/storage/repositories/booking"
	"eventbooker/src/internal/storage/repositories/events"
	"eventbooker/src/internal/storage/repositories/notifications"

	"github.com/kxddry/wbf/dbpg"
)

type TxManager struct {
	db    *dbpg.DB
	repos *Repositories
}

type Repositories struct {
	EventRepository        *events.EventRepository
	BookingRepository      *booking.BookingRepository
	NotificationRepository *notifications.NotificationRepository
}

func NewTxManager(db *dbpg.DB, repos *Repositories) *TxManager {
	return &TxManager{db: db, repos: repos}
}
