package delivery

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"warehousecontrol/src/internal/config"
	"warehousecontrol/src/internal/models"
)

type mockSvc struct {
	createFn func(ctx context.Context, req models.PostItemRequest, role models.Role, userID string) (*models.Item, error)
	getFn    func(ctx context.Context) ([]models.Item, error)
	updateFn func(ctx context.Context, req models.PutItemRequest, role models.Role, userID string) (*models.Item, error)
	deleteFn func(ctx context.Context, id string, userID string, role models.Role) error
}

func (m *mockSvc) CreateItem(ctx context.Context, req models.PostItemRequest, role models.Role, userID string) (*models.Item, error) {
	return m.createFn(ctx, req, role, userID)
}
func (m *mockSvc) GetItems(ctx context.Context) ([]models.Item, error) { return m.getFn(ctx) }
func (m *mockSvc) UpdateItem(ctx context.Context, req models.PutItemRequest, role models.Role, userID string) (*models.Item, error) {
	return m.updateFn(ctx, req, role, userID)
}
func (m *mockSvc) DeleteItem(ctx context.Context, id string, userID string, role models.Role) error {
	return m.deleteFn(ctx, id, userID, role)
}

type mockAuth struct {
	verifyFn func(ctx context.Context, token string) (models.Role, string, error)
	createFn func(ctx context.Context, role models.Role) (string, error)
}

func (m *mockAuth) VerifyJWT(ctx context.Context, token string) (models.Role, string, error) {
	return m.verifyFn(ctx, token)
}
func (m *mockAuth) CreateJWT(ctx context.Context, role models.Role) (string, error) {
	return m.createFn(ctx, role)
}

// not used in this test, keep placeholder type signature out
// type mockHistory struct{getFn func(ctx context.Context, role models.Role, uID, iID, action string, from, to time.Time, roleStr string, limit, offset int64) ([]models.HistoryEntry, error)}

func TestCreateItem_Handler(t *testing.T) {
	cfg := &config.Config{}
	s := NewServer(cfg, &mockSvc{createFn: func(ctx context.Context, req models.PostItemRequest, role models.Role, userID string) (*models.Item, error) {
		return &models.Item{ID: "x", Name: req.Name, Quantity: req.Quantity, Price: req.Price}, nil
	}}, &mockAuth{verifyFn: func(ctx context.Context, token string) (models.Role, string, error) {
		return models.RoleAdmin, "00000000-0000-0000-0000-000000000001", nil
	}}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/items", strings.NewReader(`{"name":"n","quantity":1,"price":2}`))
	req.Header.Set("Authorization", "Bearer test")
	req.Header.Set("Content-Type", "application/json")
	s.r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}
