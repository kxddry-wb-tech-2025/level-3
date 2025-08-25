package txmanager

import (
	"context"
	"database/sql"
	"salestracker/src/internal/models"
	"salestracker/src/internal/service"
	"salestracker/src/internal/storage"
	"salestracker/src/internal/storage/repositories/items"
	"time"

	"github.com/kxddry/wbf/dbpg"
)

// TxManager is the transaction manager for the salestracker service.
type TxManager struct {
	db    *dbpg.DB
	repos *Repositories
}

// Repositories is the repositories for the salestracker service.
type Repositories struct {
	ItemRepository *items.Repository
}

// New creates a new TxManager instance.
func New(masterDSN string, slaveDSNs ...string) (*TxManager, error) {
	db, err := dbpg.New(masterDSN, slaveDSNs, nil)
	if err != nil {
		return nil, err
	}

	repos := &Repositories{
		ItemRepository: items.New(db),
	}
	return &TxManager{db: db, repos: repos}, db.Master.Ping()
}

// Close closes the TxManager instance.
func (m *TxManager) Close() error {
	for _, slave := range m.db.Slaves {
		_ = slave.Close()
	}

	return m.db.Master.Close()
}

// tx is the transaction for the salestracker service.
type tx struct {
	ctx   context.Context
	repos *Repositories
	tx    *sql.Tx
}

// CreateItem creates a new item.
func (t *tx) CreateItem(ctx context.Context, item models.PostRequest) (models.PostResponse, error) {
	return t.repos.ItemRepository.CreateItem(storage.WithTx(ctx, t.tx), item)
}

// ReadItems reads all items.
func (t *tx) ReadItems(ctx context.Context) ([]models.Item, error) {
	return t.repos.ItemRepository.ReadItems(storage.WithTx(ctx, t.tx))
}

// UpdateItem updates an item.
func (t *tx) UpdateItem(ctx context.Context, id string, req models.PutRequest) (models.PutResponse, error) {
	return t.repos.ItemRepository.UpdateItem(storage.WithTx(ctx, t.tx), id, req)
}

// DeleteItem deletes an item.
func (t *tx) DeleteItem(ctx context.Context, id string) error {
	return t.repos.ItemRepository.DeleteItem(storage.WithTx(ctx, t.tx), id)
}

// Analytics returns analytics for the items.
func (t *tx) Analytics(ctx context.Context, from, to *time.Time) (models.Analytics, error) {
	return t.repos.ItemRepository.Analytics(storage.WithTx(ctx, t.tx), from, to)
}

// Commit commits the transaction.
func (t *tx) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction.
func (t *tx) Rollback() error {
	return t.tx.Rollback()
}

// Do executes a function in a transaction.
func (m *TxManager) Do(ctx context.Context, fn func(ctx context.Context, tx service.Tx) error) error {
	sqlTx, err := m.db.Master.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	t := &tx{ctx: ctx, repos: m.repos, tx: sqlTx}
	ctxWithTx := storage.WithTx(ctx, sqlTx)
	if err := fn(ctxWithTx, t); err != nil {
		_ = t.Rollback()
		return err
	}
	return t.Commit()
}
