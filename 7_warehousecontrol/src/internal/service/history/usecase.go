package history

import (
	"context"
	"fmt"
	"time"
	"warehousecontrol/src/internal/models"
	"warehousecontrol/src/internal/repo"

	"github.com/google/uuid"
)

// Repository is the interface for the repository.
type Repository interface {
	GetHistory(ctx context.Context, args *repo.HistoryArgs) ([]models.HistoryEntry, error)
}

// Usecase is the usecase for the history.
type Usecase struct {
	repo Repository
}

// NewUsecase creates a new usecase.
func NewUsecase(repo Repository) *Usecase {
	return &Usecase{repo: repo}
}

// GetHistory gets the history.
func (u *Usecase) GetHistory(
	ctx context.Context, role models.Role, filterByUserID string,
	filterByItemID string, filterByAction string, filterDateFrom time.Time,
	filterDateTo time.Time, filterByUserRole string, limit, offset int64,
) ([]models.HistoryEntry, error) {
	if role == models.RoleUser {
		return nil, fmt.Errorf("%w: user role cannot access history", models.ErrForbidden)
	}

	if filterByUserID != "" && uuid.Validate(filterByUserID) != nil {
		return nil, fmt.Errorf("invalid user id")
	}

	if filterByItemID != "" && uuid.Validate(filterByItemID) != nil {
		return nil, fmt.Errorf("invalid item id")
	}
	switch filterByAction {
	case "INSERT", "UPDATE", "DELETE":
	case models.ActionCreate:
		filterByAction = "INSERT"
	case models.ActionDelete:
		filterByAction = "DELETE"
	case models.ActionUpdate:
		filterByAction = "UPDATE"
	case "":
	default:
		return nil, fmt.Errorf("invalid action")
	}

	filterByUserRoleInt := models.Role(0)
	switch filterByUserRole {
	case "user":
		filterByUserRoleInt = models.RoleUser
	case "manager":
		filterByUserRoleInt = models.RoleManager
	case "admin":
		filterByUserRoleInt = models.RoleAdmin
	case "":
	default:
		return nil, fmt.Errorf("invalid user role")
	}

	args := &repo.HistoryArgs{
		FilterByUserID:   filterByUserID,
		FilterByItemID:   filterByItemID,
		FilterByAction:   filterByAction,
		FilterDateFrom:   filterDateFrom,
		FilterDateTo:     filterDateTo,
		FilterByUserRole: filterByUserRoleInt,
		Limit:            limit,
		Offset:           offset,
	}

	return u.repo.GetHistory(ctx, args)
}
