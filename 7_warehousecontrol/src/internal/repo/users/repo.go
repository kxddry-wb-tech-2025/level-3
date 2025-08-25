package users

import (
	"context"
	"database/sql"
	"errors"
	"warehousecontrol/src/internal/models"

	"github.com/kxddry/wbf/dbpg"
)

type Repository struct {
	db *dbpg.DB
}

func New(masterDSN string, slaveDSNs ...string) (*Repository, error) {
	db, err := dbpg.New(masterDSN, slaveDSNs, nil)
	if err != nil {
		return nil, err
	}

	return &Repository{db: db}, db.Master.Ping()
}

func (r *Repository) CreateUser(ctx context.Context, role string) (id string, err error) {
	query := `
		INSERT INTO users (role) VALUES ($1) RETURNING id
	`

	err = r.db.Master.QueryRowContext(ctx, query, role).Scan(&id)
	return
}

func (r *Repository) GetRole(ctx context.Context, id string) (role string, err error) {
	query := `
		SELECT role FROM users WHERE id = $1
	`

	err = r.db.Master.QueryRowContext(ctx, query, id).Scan(&role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", models.ErrUserNotFound
		}
		return "", err
	}

	return
}
