package history

import (
	"context"
	"warehousecontrol/src/internal/repo"

	"github.com/kxddry/wbf/dbpg"
)

type Repository struct {
	db *dbpg.DB
}

func New(masterDSN string, slaveDSNs ...string) (*Repository, error) {
	db, err := dbpg.New(masterDSN, slaveDSNs, nil)
	if err != nil {
		return nil, err
	}

	return &Repository{db: db}, db.Master.Ping()
}

func (r *Repository) Close() error {
	for _, db := range r.db.Slaves {
		_ = db.Close()
	}

	return r.db.Master.Close()
}

func (r *Repository) GetHistory(ctx context.Context, args *repo.HistoryArgs) ([]repo.HistoryEntry, error) {
	query := `
		SELECT id, action, item_id, user_id, changed_at FROM items_history
	`
	var queryArgs []any

	if args != nil {
		if args.FilterByAction != "" {
			query += " AND action = $1"
			queryArgs = append(queryArgs, args.FilterByAction)
		}

		if args.FilterByItemID != "" {
			query += " AND item_id = $1"
			queryArgs = append(queryArgs, args.FilterByItemID)
		}

		if args.FilterByUserID != "" {
			query += " AND user_id = $1"
			queryArgs = append(queryArgs, args.FilterByUserID)
		}

		if !args.FilterDateFrom.IsZero() {
			query += " AND changed_at >= $1"
			queryArgs = append(queryArgs, args.FilterDateFrom)
		}

		if !args.FilterDateTo.IsZero() {
			query += " AND changed_at <= $1"
			queryArgs = append(queryArgs, args.FilterDateTo)
		}

		if args.FilterByUserRole != 0 {
			query += " AND user_id IN (SELECT id FROM users WHERE role = $1)"
			queryArgs = append(queryArgs, args.FilterByUserRole)
		}
	}

	query += " ORDER BY changed_at DESC"

	rows, err := r.db.Master.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var output []repo.HistoryEntry
	for rows.Next() {
		var entry repo.HistoryEntry
		err = rows.Scan(&entry.ID, &entry.Action, &entry.ItemID, &entry.UserID, &entry.ChangedAt)
		if err != nil {
			return nil, err
		}
		output = append(output, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return output, nil
}
