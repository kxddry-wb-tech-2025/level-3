package items

import (
	"context"
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

func (r *Repository) CreateItem(ctx context.Context, req models.PostItemRequest, role models.Role, userID string) (id string, err error) {
	panic("not implemented")
}

func (r *Repository) GetItems(ctx context.Context) ([]models.Item, error) {
	panic("not implemented")
}

func (r *Repository) GetItem(ctx context.Context, id string) (models.Item, error) {
	panic("not implemented")
}

func (r *Repository) UpdateItem(ctx context.Context, req models.PutItemRequest, userID string) error {
	panic("not implemented")
}

func (r *Repository) DeleteItem(ctx context.Context, id string, userID string) error {
	panic("not implemented")
}

func (r *Repository) Close() error {
	for _, db := range r.db.Slaves {
		_ = db.Close()
	}

	return r.db.Master.Close()
}
