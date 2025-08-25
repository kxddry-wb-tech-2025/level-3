package service

import (
	"context"
	"errors"
	"fmt"
	"salestracker/src/internal/models"
	"salestracker/src/internal/storage"
	"time"
)

type Repository interface {
	CreateItem(ctx context.Context, req models.PostRequest) (models.PostResponse, error)
	ReadItems(ctx context.Context) ([]models.Item, error)
	UpdateItem(ctx context.Context, id string, req models.PutRequest) (models.PutResponse, error)
	DeleteItem(ctx context.Context, id string) error
	Analytics(ctx context.Context, from, to *time.Time) (models.Analytics, error)
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
	if req.Date.IsZero() {
		req.Date = time.Now()
	} else if req.Date.After(time.Now()) {
		return models.PostResponse{}, fmt.Errorf("%w: %s", models.ErrInvalidDate, "date is in the future")
	}
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
		if errors.Is(err, storage.ErrNotFound) {
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

func (u *Usecase) GetAnalytics(ctx context.Context, from, to *time.Time) (models.Analytics, error) {
	var analytics models.Analytics
	if from == nil {
		from = &time.Time{}
	}
	if to == nil {
		t := time.Now()
		to = &t
	}
	if err := u.mgr.Do(ctx, func(ctx context.Context, tx Tx) error {
		analytics2, err := tx.Analytics(ctx, from, to)
		if err != nil {
			return err
		}

		analytics = analytics2
		return err
	}); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return models.Analytics{}, models.ErrItemNotFound
		}

		return models.Analytics{}, err
	}

	return analytics, nil
}
