package service

import (
	"context"
	"errors"
	"salestracker/src/internal/models"
	"salestracker/src/internal/repo"
)

type Repository interface {
	CreateItem(ctx context.Context, req models.PostRequest) (models.PostResponse, error)
	ReadItems(ctx context.Context) ([]models.Item, error)
	UpdateItem(ctx context.Context, id string, req models.PutRequest) (models.PutResponse, error)
	DeleteItem(ctx context.Context, id string) error
	Analytics(ctx context.Context) (models.Analytics, error)
}

type Tx interface {
	Repository
	Commit() error
	Rollback() error
}

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error
}

type Usecase struct {
	mgr TxManager
}

func NewUsecase(mgr TxManager) *Usecase {
	return &Usecase{mgr: mgr}
}

func (u *Usecase) PostItem(ctx context.Context, req models.PostRequest) (models.PostResponse, error) {
	var resp models.PostResponse
	if err := u.mgr.Do(ctx, func(ctx context.Context, tx Tx) error {
		resp2, err := tx.CreateItem(ctx, req)
		if err != nil {
			return err
		}

		resp = resp2
		return err
	}); err != nil {
		return models.PostResponse{}, err
	}

	return resp, nil
}

func (u *Usecase) GetItems(ctx context.Context) ([]models.Item, error) {
	var items []models.Item
	if err := u.mgr.Do(ctx, func(ctx context.Context, tx Tx) error {
		items2, err := tx.ReadItems(ctx)
		if err != nil {
			return err
		}

		items = items2
		return err
	}); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, models.ErrItemNotFound
		}

		return nil, err
	}

	return items, nil
}

func (u *Usecase) PutItem(ctx context.Context, id string, req models.PutRequest) (models.PutResponse, error) {
	var resp models.PutResponse
	if err := u.mgr.Do(ctx, func(ctx context.Context, tx Tx) error {
		resp2, err := tx.UpdateItem(ctx, id, req)
		if err != nil {
			return err
		}

		resp = resp2
		return err
	}); err != nil {
		return models.PutResponse{}, err
	}

	return resp, nil
}

func (u *Usecase) DeleteItem(ctx context.Context, id string) error {
	if err := u.mgr.Do(ctx, func(ctx context.Context, tx Tx) error {
		return tx.DeleteItem(ctx, id)
	}); err != nil {
		return err
	}

	return nil
}

func (u *Usecase) GetAnalytics(ctx context.Context) (models.Analytics, error) {
	var analytics models.Analytics
	if err := u.mgr.Do(ctx, func(ctx context.Context, tx Tx) error {
		analytics2, err := tx.Analytics(ctx)
		if err != nil {
			return err
		}

		analytics = analytics2
		return err
	}); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return models.Analytics{}, models.ErrItemNotFound
		}

		return models.Analytics{}, err
	}

	return analytics, nil
}
