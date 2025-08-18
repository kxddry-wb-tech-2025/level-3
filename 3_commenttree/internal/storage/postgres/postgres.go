package postgres

import (
	"comment-tree/internal/domain"
	"comment-tree/internal/storage"
	"context"
	"database/sql"
	"errors"

	"github.com/kxddry/wbf/dbpg"
)

type Storage struct {
	db *dbpg.DB
}

func New(db *dbpg.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) AddComment(ctx context.Context, comment domain.Comment) error {
	query := `
		INSERT INTO comments (id, content, parent_id, created_at)
		VALUES ($1, $2, $3, $4)
	`

	if _, err := s.db.ExecContext(ctx, query, comment.ID, comment.Content, comment.ParentID, comment.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrNotFound
		}
		return err
	}

	return nil
}

func (s *Storage) GetCommentTree(ctx context.Context, rootID string) (*domain.CommentTree, error) {
	query := `
		WITH RECURSIVE thread AS (
			SELECT id, content, parent_id, created_at
			FROM comments
			WHERE id = $1
			UNION ALL
			SELECT c.id, c.content, c.parent_id, c.created_at
			FROM comments c
			INNER JOIN thread t ON c.parent_id = t.id
		)
		SELECT id, content, parent_id, created_at
		FROM thread
		ORDER BY created_at ASC;
	`

	rows, err := s.db.QueryContext(ctx, query, rootID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	childrenMap := make(map[string][]*domain.CommentTree)
	var rootComment *domain.CommentTree

	for rows.Next() {
		var c domain.Comment
		if err := rows.Scan(&c.ID, &c.Content, &c.ParentID, &c.CreatedAt); err != nil {
			return nil, err
		}

		node := &domain.CommentTree{
			ID:        c.ID,
			Content:   c.Content,
			CreatedAt: c.CreatedAt,
		}

		if c.ID == rootID {
			rootComment = node
		}

		childrenMap[c.ParentID] = append(childrenMap[c.ParentID], node)
	}

	if rootComment == nil {
		return nil, storage.ErrNotFound
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
