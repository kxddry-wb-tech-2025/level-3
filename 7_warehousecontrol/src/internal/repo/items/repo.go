package items

import (
	"context"
	"database/sql"
	"errors"
	"warehousecontrol/src/internal/models"

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

func (r *Repository) CreateItem(ctx context.Context, req models.PostItemRequest, userID string) (id string, err error) {
	query := `
	INSERT INTO items (name, description, quantity, price, created_by) VALUES ($1, $2, $3, $4, $5) RETURNING id
	`

	err = r.db.Master.QueryRowContext(ctx, query, req.Name, req.Description, req.Quantity, req.Price, userID).Scan(&id)
	return
}

func (r *Repository) GetItems(ctx context.Context) ([]models.Item, error) {
	query := `
		SELECT id, name, description, quantity, price FROM items WHERE deleted_at IS NULL ORDER BY created_at DESC
	`

	rows, err := r.db.Master.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []models.Item{}
	for rows.Next() {
		var item models.Item
		err = rows.Scan(&item.ID, &item.Name, &item.Description, &item.Quantity, &item.Price)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *Repository) GetItem(ctx context.Context, id string) (item models.Item, err error) {
	query := `
		SELECT id, name, description, quantity, price FROM items WHERE id = $1 AND deleted_at IS NULL
	`

	err = r.db.Master.QueryRowContext(ctx, query, id).Scan(&item.ID, &item.Name, &item.Description, &item.Quantity, &item.Price)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Item{}, models.ErrItemNotFound
		}
		return models.Item{}, err
	}
	return item, nil
}

func (r *Repository) UpdateItem(ctx context.Context, req models.PutItemRequest, userID string) error {
	query := `
		UPDATE items SET name = $1, description = $2, quantity = $3, price = $4, updated_by = $5 WHERE id = $6
	`

	_, err := r.db.Master.ExecContext(ctx, query, req.Name, req.Description, req.Quantity, req.Price, userID, req.ID)
	return err
}

func (r *Repository) DeleteItem(ctx context.Context, id string, userID string) error {
	query := `
		UPDATE items SET deleted_at = NOW(), deleted_by = $1 WHERE id = $2 AND deleted_at IS NULL
	`

	_, err := r.db.Master.ExecContext(ctx, query, userID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ErrItemNotFound
		}
		return err
	}

	return err
}

func (r *Repository) Close() error {
	for _, db := range r.db.Slaves {
		_ = db.Close()
	}

	return r.db.Master.Close()
}
