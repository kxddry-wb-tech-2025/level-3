package domain

import "time"

type Comment struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	ParentID  string    `json:"parent_id"`
	CreatedAt time.Time `json:"created_at"`
}

type CommentTree struct {
	ID        string         `json:"id"`
	Content   string         `json:"content"`
	CreatedAt time.Time      `json:"created_at"`
	Children  []*CommentTree `json:"children"`
}

type AddCommentRequest struct {
	Content  string `json:"content"`
	ParentID string `json:"parent_id,omitempty"`
}
