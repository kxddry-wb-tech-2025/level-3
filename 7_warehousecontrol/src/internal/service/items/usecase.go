package auth

import (
	"context"
	"warehousecontrol/src/internal/models"
)

type Repository interface {
	CreateItem(ctx context.Context, req models.PostItemRequest) (id string, err error)
	GetItems(ctx context.Context) ([]models.Item, error)
	UpdateItem(ctx context.Context, req models.Item, userID string) error
	DeleteItem(ctx context.Context, id string) error
}

type Usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) CreateItem(ctx context.Context, req models.PostItemRequest) (*models.Item, error) {
	id, err := u.repo.CreateItem(ctx, req)
	if err != nil {
		return nil, err
	}

	return &models.Item{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Quantity:    req.Quantity,
		Price:       req.Price,
	}, nil
}

func (u *Usecase) GetItems(ctx context.Context) ([]models.Item, error) {
	return u.repo.GetItems(ctx)
}
