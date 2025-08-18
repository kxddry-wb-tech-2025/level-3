package postgres

import (
	"comment-tree/internal/domain"
	"comment-tree/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/kxddry/wbf/dbpg"
)

type Storage struct {
	db *dbpg.DB
}

func New(db *dbpg.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) AddComment(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
	query := `
		INSERT INTO comments (content, parent_id, created_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	// if parentID is empty (comment is a root comment), use NULL to avoid UUID errors
	var parentParam sql.NullString
	if comment.ParentID != "" {
		parentParam = sql.NullString{String: comment.ParentID, Valid: true}
	}

	rows, err := s.db.QueryContext(ctx, query, comment.Content, parentParam, comment.CreatedAt)
	if err != nil {
		return domain.Comment{}, err
	}

	defer rows.Close()

	if !rows.Next() {
		return domain.Comment{}, storage.ErrNotFound
	}

	if err := rows.Scan(&comment.ID); err != nil {
		return domain.Comment{}, err
	}

	return comment, nil
}

func (s *Storage) GetComments(ctx context.Context, parentID string, asc bool, limit, offset int) (*domain.CommentTree, error) {
	order := "ASC"
	if !asc {
		order = "DESC"
	}

	var query string
	var args []any

	if parentID == "" {
		query = fmt.Sprintf(`
		WITH roots AS ( SELECT id FROM comments WHERE parent_id IS NULL ORDER BY created_at %s LIMIT $1 OFFSET $2),
		thread AS ( SELECT c.id, c.content, c.parent_id, c.created_at FROM comments c JOIN roots r ON c.id = r.id UNION ALL
		SELECT c.id, c.content, c.parent_id, c.created_at FROM comments c INNER JOIN thread t ON c.parent_id = t.id)
		SELECT id, content, parent_id, created_at FROM thread ORDER BY created_at %s;`, order, order )
		args = []any{limit, offset}
	} else {
		query = fmt.Sprintf(`
		WITH RECURSIVE thread AS (
			SELECT id, content, parent_id, created_at
			FROM comments
			WHERE id = $1 UNION ALL SELECT c.id, c.content, c.parent_id, c.created_at FROM comments c INNER JOIN thread t ON c.parent_id = t.id)
			SELECT id, content, parent_id, created_at FROM thread ORDER BY created_at %s;`, order)
		args = []any{parentID}
	}


	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	childrenMap := make(map[string][]*domain.CommentTree)
	var rootComment *domain.CommentTree
	foundAny := false

	for rows.Next() {
		var c domain.Comment
		var parentNullable sql.NullString
		if err := rows.Scan(&c.ID, &c.Content, &parentNullable, &c.CreatedAt); err != nil {
			return nil, err
		}
		foundAny = true
		if parentNullable.Valid {
			c.ParentID = parentNullable.String
		}

		node := &domain.CommentTree{
			ID:        c.ID,
			Content:   c.Content,
			CreatedAt: c.CreatedAt,
		}

		if c.ID == parentID {
			rootComment = node
		}

		childrenMap[c.ParentID] = append(childrenMap[c.ParentID], node)
	}

	if rootComment == nil {
		// if there are no comments whatsoever, return empty tree if parentID is empty
		if parentID == "" {
			if foundAny {
				rootComment = &domain.CommentTree{}
			} else {
				return nil, nil
			}
		} else {
			return nil, storage.ErrNotFound
		}
	}

	var attachChildren func(node *domain.CommentTree)
	attachChildren = func(node *domain.CommentTree) {
		node.Children = childrenMap[node.ID]
		for _, child := range node.Children {
			attachChildren(child)
		}
	}

	attachChildren(rootComment)
	return rootComment, nil
}

func (s *Storage) SearchComments(ctx context.Context, q string, limit, offset int) ([]domain.Comment, error) {
	query := `
		SELECT id, content, parent_id, created_at
		FROM comments
		WHERE to_tsvector('simple', content) @@ plainto_tsquery('simple', $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, query, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Comment
	for rows.Next() {
		var c domain.Comment
		var parentNullable sql.NullString
		if err := rows.Scan(&c.ID, &c.Content, &parentNullable, &c.CreatedAt); err != nil {
			return nil, err
		}
		if parentNullable.Valid {
			c.ParentID = parentNullable.String
		}
		out = append(out, c)
	}

	return out, nil
}

func (s *Storage) DeleteComments(ctx context.Context, id string) error {
	query := `
		DELETE FROM comments
		WHERE id = $1
	`

	if _, err := s.db.ExecContext(ctx, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrNotFound
		}
		return err
	}

	return nil
}

func (s *Storage) Close() error {
	for _, conn := range s.db.Slaves {
		_ = conn.Close()
	}

	return s.db.Master.Close()
}
