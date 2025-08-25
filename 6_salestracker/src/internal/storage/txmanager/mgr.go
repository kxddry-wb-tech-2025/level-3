package txmanager

import (
	"context"
	"database/sql"
	"salestracker/src/internal/models"
	"salestracker/src/internal/service"
	"salestracker/src/internal/storage"
	"salestracker/src/internal/storage/repositories/items"

	"github.com/kxddry/wbf/dbpg"
)

type TxManager struct {
	db    *dbpg.DB
	repos *Repositories
}

type Repositories struct {
	ItemRepository *items.Repository
}

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

func (m *TxManager) Close() error {
	for _, slave := range m.db.Slaves {
		_ = slave.Close()
	}

	return m.db.Master.Close()
}

type tx struct {
	ctx   context.Context
	repos *Repositories
	tx    *sql.Tx
}

func (t *tx) CreateItem(ctx context.Context, item models.PostRequest) (models.PostResponse, error) {
	return t.repos.ItemRepository.CreateItem(storage.WithTx(ctx, t.tx), item)
}

func (t *tx) ReadItems(ctx context.Context) ([]models.Item, error) {
	return t.repos.ItemRepository.ReadItems(storage.WithTx(ctx, t.tx))
}

func (t *tx) UpdateItem(ctx context.Context, id string, req models.PutRequest) (models.PutResponse, error) {
	return t.repos.ItemRepository.UpdateItem(storage.WithTx(ctx, t.tx), id, req)
}

func (t *tx) DeleteItem(ctx context.Context, id string) error {
	return t.repos.ItemRepository.DeleteItem(storage.WithTx(ctx, t.tx), id)
}

func (t *tx) Analytics(ctx context.Context) (models.Analytics, error) {
	return t.repos.ItemRepository.Analytics(storage.WithTx(ctx, t.tx))
}

func (t *tx) Commit() error {
	return t.tx.Commit()
}

func (t *tx) Rollback() error {
	return t.tx.Rollback()
}

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
