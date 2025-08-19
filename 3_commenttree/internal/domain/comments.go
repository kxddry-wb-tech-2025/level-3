package domain

import "time"

// Comment is the main comment struct that contains the comment's id, content, parent id, and creation time.
type Comment struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	ParentID  string    `json:"parent_id"`
	CreatedAt time.Time `json:"created_at"`
}

// CommentTree is the struct that contains the comment's id, content, creation time, and children comments.
type CommentTree struct {
	ID        string         `json:"id"`
	Content   string         `json:"content"`
	CreatedAt time.Time      `json:"created_at"`
	Children  []*CommentTree `json:"children"`
}

// AddCommentRequest is the struct that contains the comment's content and parent id.
type AddCommentRequest struct {
	Content  string `json:"content"`
	ParentID string `json:"parent_id,omitempty"`
}
