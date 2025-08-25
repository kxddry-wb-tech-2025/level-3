package items

import (
	"context"
	"database/sql"
	"errors"
	"salestracker/src/internal/models"
	"salestracker/src/internal/storage"
	"time"

	"github.com/kxddry/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

type Repository struct {
	db *dbpg.DB
}

func New(db *dbpg.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateItem(ctx context.Context, item models.PostRequest) (models.PostResponse, error) {
	log := zlog.Logger.With().Str("component", "items").Logger().With().Str("operation", "CreateItem").Logger()
	var id string
	query := `
	INSERT INTO items (title, price, description, item_date, category) VALUES ($1, $2, $3, $4, $5) RETURNING id
	`
	if tx, ok := storage.TxFromContext(ctx); ok {
		err := tx.QueryRowContext(ctx, query, item.Title, item.Price, item.Description, item.Date, item.Category).Scan(&id)
		if err != nil {
			log.Error().Err(err).Msg("failed to create item")
			return models.PostResponse{}, err
		}
		resp := models.PostResponse{
			ID:          id,
			Title:       item.Title,
			Price:       item.Price,
			Description: item.Description,
			Date:        item.Date,
			Category:    item.Category,
		}
		return resp, nil
	}
	log.Warn().Msg("no transaction found, using master connection")
	err := r.db.Master.QueryRowContext(ctx, query, item.Title, item.Price, item.Description, item.Date, item.Category).Scan(&id)
	if err != nil {
		log.Error().Err(err).Msg("failed to create item")
		return models.PostResponse{}, err
	}
	resp := models.PostResponse{
		ID:          id,
		Title:       item.Title,
		Price:       item.Price,
		Description: item.Description,
		Date:        item.Date,
		Category:    item.Category,
	}
	return resp, nil
}

func (r *Repository) ReadItems(ctx context.Context) ([]models.Item, error) {
	log := zlog.Logger.With().Str("component", "items").Logger().With().Str("operation", "ReadItems").Logger()
	query := `
	SELECT id, title, price, description, item_date, category FROM items ORDER BY item_date DESC
	`
	var rows *sql.Rows
	if tx, ok := storage.TxFromContext(ctx); ok {
		var err error
		rows, err = tx.QueryContext(ctx, query)
		if err != nil {
			log.Error().Err(err).Msg("failed to read items")
			return nil, err
		}
	} else {
		log.Warn().Msg("no transaction found, using master connection")
		var err error
		rows, err = r.db.Master.QueryContext(ctx, query)
		if err != nil {
			log.Error().Err(err).Msg("failed to read items")
			return nil, err
		}
	}
	defer rows.Close()
	var items []models.Item
	for rows.Next() {
		var item models.Item
		err := rows.Scan(&item.ID, &item.Title, &item.Price, &item.Description, &item.Date, &item.Category)
		if err != nil {
			log.Error().Err(err).Msg("failed to scan item")
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) UpdateItem(ctx context.Context, id string, req models.PutRequest) (models.PutResponse, error) {
	log := zlog.Logger.With().Str("component", "items").Logger().With().Str("operation", "UpdateItem").Logger()
	query := `
	UPDATE items SET title = $1, price = $2, description = $3, item_date = $4, category = $5 WHERE id = $6
	`
	if tx, ok := storage.TxFromContext(ctx); ok {
		_, err := tx.ExecContext(ctx, query, req.Title, req.Price, req.Description, req.Date, req.Category, id)
		if err != nil {
			log.Error().Err(err).Msg("failed to update item")
			return models.PutResponse{}, err
		}
		return models.PutResponse{
			ID:          id,
			Title:       req.Title,
			Price:       req.Price,
			Description: req.Description,
			Date:        req.Date,
			Category:    req.Category,
		}, nil
	}
	log.Warn().Msg("no transaction found, using master connection")
	_, err := r.db.Master.ExecContext(ctx, query, req.Title, req.Price, req.Description, req.Date, req.Category, id)
	if err != nil {
		log.Error().Err(err).Msg("failed to update item")
		return models.PutResponse{}, err
	}
	return models.PutResponse{
		ID:          id,
		Title:       req.Title,
		Price:       req.Price,
		Description: req.Description,
		Date:        req.Date,
		Category:    req.Category,
	}, nil
}

func (r *Repository) DeleteItem(ctx context.Context, id string) error {
	log := zlog.Logger.With().Str("component", "items").Logger().With().Str("operation", "DeleteItem").Logger()
	query := `
	DELETE FROM items WHERE id = $1
	`
	var err error
	if tx, ok := storage.TxFromContext(ctx); ok {
		_, err = tx.ExecContext(ctx, query, id)
	} else {
		log.Warn().Msg("no transaction found, using master connection")
		_, err = r.db.Master.ExecContext(ctx, query, id)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Error().Err(err).Msg("item not found")
			return models.ErrItemNotFound
		}
		log.Error().Err(err).Msg("failed to delete item")
		return err
	}
	return nil
}

func (r *Repository) Analytics(ctx context.Context, from, to *time.Time) (models.Analytics, error) {
	log := zlog.Logger.With().Str("component", "items").Logger().With().Str("operation", "Analytics").Logger()
	query := `
	SELECT 
		COALESCE(SUM(price), 0) AS sum, 
		COUNT(*) AS count, 
		COALESCE(AVG(price), 0) AS average, 
		COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY price), 0) AS median, 
		COALESCE(PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY price), 0) AS percentile_90 FROM items
		WHERE item_date BETWEEN $1 AND $2
	`
	var err error
	var anal models.Analytics
	if tx, ok := storage.TxFromContext(ctx); ok {
		err = tx.QueryRowContext(ctx, query, from, to).Scan(&anal.Sum, &anal.Count, &anal.Average, &anal.Median, &anal.Percentile90)
	} else {
		log.Warn().Msg("no transaction found, using master connection")
		err = r.db.Master.QueryRowContext(ctx, query, from, to).Scan(&anal.Sum, &anal.Count, &anal.Average, &anal.Median, &anal.Percentile90)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Error().Err(err).Msg("no items found")
			return models.Analytics{}, models.ErrItemNotFound
		}
		log.Error().Err(err).Msg("failed to get analytics")
		return models.Analytics{}, err
	}
	return anal, nil
}
