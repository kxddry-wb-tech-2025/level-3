package repo

import (
	"time"
	"warehousecontrol/src/internal/models"
)

type HistoryEntry struct {
	ID        string    `json:"id"`
	Action    string    `json:"action"`
	ItemID    string    `json:"item_id"`
	UserID    string    `json:"user_id"`
	ChangedAt time.Time `json:"changed_at"`
}

type HistoryArgs struct {
	FilterByUserID   string
	FilterByItemID   string
	FilterByAction   string
	FilterDateFrom   time.Time
	FilterDateTo     time.Time
	FilterByUserRole models.Role
}
