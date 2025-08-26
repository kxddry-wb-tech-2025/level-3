package auth

import (
	"context"
	"errors"
	"testing"
	"warehousecontrol/src/internal/models"
)

type mockRepo struct {
	createFn func(ctx context.Context, req models.PostItemRequest, userID string) (string, error)
	getAllFn func(ctx context.Context) ([]models.Item, error)
	getFn    func(ctx context.Context, id string) (models.Item, error)
	updateFn func(ctx context.Context, req models.PutItemRequest, userID string) error
	deleteFn func(ctx context.Context, id string, userID string) error
}

func (m *mockRepo) CreateItem(ctx context.Context, req models.PostItemRequest, userID string) (string, error) {
	return m.createFn(ctx, req, userID)
}
func (m *mockRepo) GetItems(ctx context.Context) ([]models.Item, error) { return m.getAllFn(ctx) }
func (m *mockRepo) GetItem(ctx context.Context, id string) (models.Item, error) {
	return m.getFn(ctx, id)
}
func (m *mockRepo) UpdateItem(ctx context.Context, req models.PutItemRequest, userID string) error {
	return m.updateFn(ctx, req, userID)
}
func (m *mockRepo) DeleteItem(ctx context.Context, id string, userID string) error {
	return m.deleteFn(ctx, id, userID)
}

func TestCreateItem_Authorization(t *testing.T) {
	u := NewUsecase(&mockRepo{})
	_, err := u.CreateItem(context.Background(), models.PostItemRequest{Name: "n", Quantity: 1, Price: 1}, models.RoleUser, "uid")
	if err == nil {
		t.Fatalf("expected unauthorized error")
	}
}

func TestCreateItem_Success(t *testing.T) {
	repo := &mockRepo{
		createFn: func(ctx context.Context, req models.PostItemRequest, userID string) (string, error) {
			return "id-1", nil
		},
	}
	u := NewUsecase(repo)
	item, err := u.CreateItem(context.Background(), models.PostItemRequest{Name: "n", Quantity: 1, Price: 1}, models.RoleManager, "uid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "id-1" || item.Name != "n" {
		t.Fatalf("unexpected item: %+v", item)
	}
}

func TestGetItems_Error(t *testing.T) {
	repo := &mockRepo{getAllFn: func(ctx context.Context) ([]models.Item, error) { return nil, errors.New("boom") }}
	u := NewUsecase(repo)
	_, err := u.GetItems(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestUpdateItem_Authorization(t *testing.T) {
	u := NewUsecase(&mockRepo{})
	_, err := u.UpdateItem(context.Background(), models.PutItemRequest{ID: "id"}, models.RoleUser, "uid")
	if err == nil {
		t.Fatalf("expected unauthorized error")
	}
}

func TestUpdateItem_Success(t *testing.T) {
	repo := &mockRepo{updateFn: func(ctx context.Context, req models.PutItemRequest, userID string) error { return nil }}
	u := NewUsecase(repo)
	out, err := u.UpdateItem(context.Background(), models.PutItemRequest{ID: "id", Name: "n"}, models.RoleAdmin, "uid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ID != "id" || out.Name != "n" {
		t.Fatalf("unexpected item: %+v", out)
	}
}

func TestDeleteItem_RoleCheck(t *testing.T) {
	u := NewUsecase(&mockRepo{})
	if err := u.DeleteItem(context.Background(), "id", "uid", models.RoleManager); err == nil {
		t.Fatalf("expected unauthorized error")
	}
}

func TestDeleteItem_Success(t *testing.T) {
	repo := &mockRepo{deleteFn: func(ctx context.Context, id string, userID string) error { return nil }}
	u := NewUsecase(repo)
	if err := u.DeleteItem(context.Background(), "id", "uid", models.RoleAdmin); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
