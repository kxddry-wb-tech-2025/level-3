package items

import (
	"context"
	"salestracker/src/internal/models"

	"github.com/kxddry/wbf/dbpg"
)

type Repository struct {
	db *dbpg.DB
}

func New(db *dbpg.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateItem(ctx context.Context, item models.PostRequest) (models.PostResponse, error) {
	panic("not implemented")
}

func (r *Repository) ReadItems(ctx context.Context) ([]models.Item, error) {
	panic("not implemented")
}

func (r *Repository) UpdateItem(ctx context.Context, id string, req models.PutRequest) (models.PutResponse, error) {
	panic("not implemented")
}

func (r *Repository) DeleteItem(ctx context.Context, id string) error {
	panic("not implemented")
}

func (r *Repository) Analytics(ctx context.Context) (models.Analytics, error) {
	panic("not implemented")
}
