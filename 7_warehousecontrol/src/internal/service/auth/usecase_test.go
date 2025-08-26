package auth

import (
	"context"
	"testing"
	"warehousecontrol/src/internal/models"
)

type mockUserRepo struct {
	createFn func(ctx context.Context, role models.Role) (string, error)
	roleFn   func(ctx context.Context, id string) (models.Role, error)
}

func (m *mockUserRepo) CreateUser(ctx context.Context, role models.Role) (string, error) {
	return m.createFn(ctx, role)
}
func (m *mockUserRepo) GetRole(ctx context.Context, id string) (models.Role, error) {
	return m.roleFn(ctx, id)
}

func TestCreateJWT_Success(t *testing.T) {
	u := NewUsecase(&mockUserRepo{createFn: func(ctx context.Context, role models.Role) (string, error) {
		return "00000000-0000-0000-0000-000000000001", nil
	}}, "secret")
	token, err := u.CreateJWT(context.Background(), models.RoleManager)
	if err != nil || token == "" {
		t.Fatalf("expected token, got %q err=%v", token, err)
	}
}

func TestVerifyJWT_RoleMismatch(t *testing.T) {
	// create token with role admin but repo returns manager
	ur := &mockUserRepo{createFn: func(ctx context.Context, role models.Role) (string, error) {
		return "00000000-0000-0000-0000-000000000001", nil
	}}
	u := NewUsecase(ur, "secret")
	token, _ := u.CreateJWT(context.Background(), models.RoleAdmin)
	ur.roleFn = func(ctx context.Context, id string) (models.Role, error) { return models.RoleManager, nil }
	_, _, err := u.VerifyJWT(context.Background(), token)
	if err == nil {
		t.Fatalf("expected role mismatch error")
	}
}

func TestVerifyJWT_InvalidToken(t *testing.T) {
	u := NewUsecase(&mockUserRepo{roleFn: func(ctx context.Context, id string) (models.Role, error) { return models.RoleAdmin, nil }}, "secret")
	if _, _, err := u.VerifyJWT(context.Background(), "bad.token"); err == nil {
		t.Fatalf("expected error for invalid token")
	}
}
