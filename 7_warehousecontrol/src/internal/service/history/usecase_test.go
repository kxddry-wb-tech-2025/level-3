package history

import (
	"context"
	"errors"
	"testing"
	"time"
	"warehousecontrol/src/internal/models"
	"warehousecontrol/src/internal/repo"
)

type mockRepo struct {
	getFn func(ctx context.Context, args *repo.HistoryArgs) ([]models.HistoryEntry, error)
}

func (m *mockRepo) GetHistory(ctx context.Context, args *repo.HistoryArgs) ([]models.HistoryEntry, error) {
	return m.getFn(ctx, args)
}

func TestGetHistory_ForbiddenForUser(t *testing.T) {
	u := NewUsecase(&mockRepo{})
	_, err := u.GetHistory(context.Background(), models.RoleUser, "", "", "", time.Time{}, time.Time{}, "", 10, 0)
	if err == nil {
		t.Fatalf("expected forbidden error")
	}
}

func TestGetHistory_ActionMappingCreate(t *testing.T) {
	var captured *repo.HistoryArgs
	u := NewUsecase(&mockRepo{getFn: func(ctx context.Context, args *repo.HistoryArgs) ([]models.HistoryEntry, error) {
		captured = args
		return nil, nil
	}})

	_, err := u.GetHistory(context.Background(), models.RoleManager, "", "", "create", time.Time{}, time.Time{}, "", 5, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured == nil || captured.FilterByAction != "INSERT" {
		t.Fatalf("action not mapped to INSERT: %+v", captured)
	}
}

func TestGetHistory_InvalidUserRoleFilter(t *testing.T) {
	u := NewUsecase(&mockRepo{getFn: func(ctx context.Context, args *repo.HistoryArgs) ([]models.HistoryEntry, error) { return nil, nil }})
	_, err := u.GetHistory(context.Background(), models.RoleAdmin, "", "", "", time.Time{}, time.Time{}, "unknown", 5, 0)
	if err == nil {
		t.Fatalf("expected error for invalid user role filter")
	}
}

func TestGetHistory_RepoErrorBubbles(t *testing.T) {
	boom := errors.New("boom")
	u := NewUsecase(&mockRepo{getFn: func(ctx context.Context, args *repo.HistoryArgs) ([]models.HistoryEntry, error) { return nil, boom }})
	_, err := u.GetHistory(context.Background(), models.RoleAdmin, "", "", "", time.Time{}, time.Time{}, "", 5, 0)
	if !errors.Is(err, boom) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
