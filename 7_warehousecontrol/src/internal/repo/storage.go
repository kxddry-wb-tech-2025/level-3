package repo

import (
	"time"
	"warehousecontrol/src/internal/models"
)

// HistoryArgs is the arguments for the history query.
type HistoryArgs struct {
	FilterByUserID   string
	FilterByItemID   string
	FilterByAction   string
	FilterDateFrom   time.Time
	FilterDateTo     time.Time
	FilterByUserRole models.Role
	Limit            int64
	Offset           int64
}
