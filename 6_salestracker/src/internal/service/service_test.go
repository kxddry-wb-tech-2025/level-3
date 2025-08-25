package service_test

import (
	"context"
	"salestracker/src/internal/mocks"
	"salestracker/src/internal/models"
	"salestracker/src/internal/service"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_PostItem_Success(t *testing.T) {
	ctx := context.Background()

	mockTx := new(mocks.Tx)
	mockTxManager := new(mocks.TxManager)

	for _, req := range []models.PostRequest{
		{
			Title:       "Item 1",
			Price:       100,
			Description: "Description 1",
			Date:        time.Now(),
			Category:    "Category 1",
		},
		{
			Title:       "Item 2",
			Price:       200,
			Description: "Description 2",
			Date:        time.Now().Add(-time.Hour * 24),
			Category:    "Category 2",
		},
		{
			Title:       "Item 3",
			Price:       300,
			Description: "",
			Date:        time.Now().Add(-time.Hour * 24 * 2),
			Category:    "Category 3",
		},
	} {
		expectedResp := models.PostResponse{
			ID:          "1",
			Title:       req.Title,
			Price:       req.Price,
			Description: req.Description,
			Date:        req.Date,
			Category:    req.Category,
		}
		mockTx.On("CreateItem", ctx, req).Return(expectedResp, nil)
		mockTxManager.On("Do", ctx, mock.AnythingOfType("func(context.Context, service.Tx) error")).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(ctx context.Context, tx service.Tx) error)
			_ = fn(ctx, mockTx)
		}).Return(nil)

		uc := service.NewUsecase(mockTxManager)

		resp, err := uc.PostItem(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)

		mockTx.AssertExpectations(t)
		mockTxManager.AssertExpectations(t)
	}
}

func TestUsecase_PostItem_InvalidDate(t *testing.T) {
	ctx := context.Background()
	mockTxManager := new(mocks.TxManager)
	uc := service.NewUsecase(mockTxManager)

	// future date should return error
	req := models.PostRequest{
		Title: "Future Item",
		Date:  time.Now().Add(24 * time.Hour),
	}

	_, err := uc.PostItem(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "date is in the future")
}
