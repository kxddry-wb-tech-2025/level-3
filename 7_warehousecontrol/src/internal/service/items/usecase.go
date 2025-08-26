package auth

import (
	"context"
	"fmt"
	"warehousecontrol/src/internal/models"
)

type Repository interface {
	CreateItem(ctx context.Context, req models.PostItemRequest, userID string) (id string, err error)
	GetItems(ctx context.Context) ([]models.Item, error)
	GetItem(ctx context.Context, id string) (models.Item, error)
	UpdateItem(ctx context.Context, req models.PutItemRequest, userID string) error
	DeleteItem(ctx context.Context, id string, userID string) error
}

type Usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) CreateItem(ctx context.Context, req models.PostItemRequest, role models.Role, userID string) (*models.Item, error) {
	if role != models.RoleAdmin && role != models.RoleManager {
		return nil, fmt.Errorf("unauthorized")
	}

	id, err := u.repo.CreateItem(ctx, req, userID)
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
	var items []models.Item
	items, err := u.repo.GetItems(ctx)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (u *Usecase) UpdateItem(ctx context.Context, req models.PutItemRequest, role models.Role, userID string) (*models.Item, error) {
	if role != models.RoleAdmin && role != models.RoleManager {
		return nil, fmt.Errorf("unauthorized")
	}

	err := u.repo.UpdateItem(ctx, req, userID)
	if err != nil {
		return nil, err
	}

	return &models.Item{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Quantity:    req.Quantity,
		Price:       req.Price,
	}, nil
}

func (u *Usecase) DeleteItem(ctx context.Context, id string, userID string, role models.Role) error {
	if role != models.RoleAdmin {
		return fmt.Errorf("unauthorized")
	}

	err := u.repo.DeleteItem(ctx, id, userID)
	if err != nil {
		return err
	}
	return nil
}
