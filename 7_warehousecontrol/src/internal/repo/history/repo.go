package history

import (
	"context"
	"fmt"
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
		SELECT id, action, item_id, user_id, changed_at 
		FROM items_history
		WHERE 1=1
	`
	var queryArgs []any
	argIdx := 1

	if args != nil {
		if args.FilterByAction != "" {
			query += fmt.Sprintf(" AND action = $%d", argIdx)
			queryArgs = append(queryArgs, args.FilterByAction)
			argIdx++
		}

		if args.FilterByItemID != "" {
			query += fmt.Sprintf(" AND item_id = $%d", argIdx)
			queryArgs = append(queryArgs, args.FilterByItemID)
			argIdx++
		}

		if args.FilterByUserID != "" {
			query += fmt.Sprintf(" AND user_id = $%d", argIdx)
			queryArgs = append(queryArgs, args.FilterByUserID)
			argIdx++
		}

		if !args.FilterDateFrom.IsZero() {
			query += fmt.Sprintf(" AND changed_at >= $%d", argIdx)
			queryArgs = append(queryArgs, args.FilterDateFrom)
			argIdx++
		}

		if !args.FilterDateTo.IsZero() {
			query += fmt.Sprintf(" AND changed_at <= $%d", argIdx)
			queryArgs = append(queryArgs, args.FilterDateTo)
			argIdx++
		}

		if args.FilterByUserRole != 0 {
			query += fmt.Sprintf(" AND user_id IN (SELECT id FROM users WHERE role = $%d)", argIdx)
			queryArgs = append(queryArgs, args.FilterByUserRole)
			argIdx++
		}
	}

	query += " ORDER BY changed_at DESC"

	if args != nil {
		if args.Limit > 0 {
			query += fmt.Sprintf(" LIMIT $%d", argIdx)
			queryArgs = append(queryArgs, args.Limit)
			argIdx++
		}
		if args.Offset > 0 {
			query += fmt.Sprintf(" OFFSET $%d", argIdx)
			queryArgs = append(queryArgs, args.Offset)
			argIdx++
		}
	}

	rows, err := r.db.Master.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	output := []repo.HistoryEntry{}
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
