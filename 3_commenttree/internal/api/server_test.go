package api

import (
	"bytes"
	"comment-tree/internal/domain"
	"comment-tree/internal/storage"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
)

// MockStorage is a mock implementation of Storage interface
type MockStorage struct {
	addCommentFunc     func(ctx context.Context, comment domain.Comment) (domain.Comment, error)
	getCommentsFunc    func(ctx context.Context, parentID string, asc bool, limit, offset int) (*domain.CommentTree, error)
	deleteCommentsFunc func(ctx context.Context, id string) error
	searchCommentsFunc func(ctx context.Context, query string, limit, offset int) ([]domain.Comment, error)
}

func (m *MockStorage) AddComment(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
	if m.addCommentFunc != nil {
		return m.addCommentFunc(ctx, comment)
	}
	return domain.Comment{}, nil
}

func (m *MockStorage) GetComments(ctx context.Context, parentID string, asc bool, limit, offset int) (*domain.CommentTree, error) {
	if m.getCommentsFunc != nil {
		return m.getCommentsFunc(ctx, parentID, asc, limit, offset)
	}
	return nil, nil
}

func (m *MockStorage) DeleteComments(ctx context.Context, id string) error {
	if m.deleteCommentsFunc != nil {
		return m.deleteCommentsFunc(ctx, id)
	}
	return nil
}

func (m *MockStorage) SearchComments(ctx context.Context, query string, limit, offset int) ([]domain.Comment, error) {
	if m.searchCommentsFunc != nil {
		return m.searchCommentsFunc(ctx, query, limit, offset)
	}
	return nil, nil
}

func TestStorageInterface(t *testing.T) {
	tests := []struct {
		name         string
		storage      *MockStorage
		comment      domain.Comment
		parentID     string
		asc          bool
		limit        int
		offset       int
		query        string
		id           string
		expectAdd    domain.Comment
		expectGet    *domain.CommentTree
		expectSearch []domain.Comment
		expectError  bool
	}{
		{
			name: "successful comment operations",
			storage: &MockStorage{
				addCommentFunc: func(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
					comment.ID = "comment-123"
					comment.CreatedAt = time.Now()
					return comment, nil
				},
				getCommentsFunc: func(ctx context.Context, parentID string, asc bool, limit, offset int) (*domain.CommentTree, error) {
					return &domain.CommentTree{
						ID:        "comment-123",
						Content:   "Test comment",
						CreatedAt: time.Now(),
						Children:  []*domain.CommentTree{},
					}, nil
				},
				deleteCommentsFunc: func(ctx context.Context, id string) error {
					return nil
				},
				searchCommentsFunc: func(ctx context.Context, query string, limit, offset int) ([]domain.Comment, error) {
					return []domain.Comment{
						{
							ID:        "comment-123",
							Content:   "Test comment",
							ParentID:  "",
							CreatedAt: time.Now(),
						},
					}, nil
				},
			},
			comment: domain.Comment{
				Content:  "Test comment",
				ParentID: "",
			},
			parentID: "",
			asc:      true,
			limit:    10,
			offset:   0,
			query:    "test",
			id:       "comment-123",
			expectAdd: domain.Comment{
				ID:        "comment-123",
				Content:   "Test comment",
				ParentID:  "",
				CreatedAt: time.Now(),
			},
			expectGet: &domain.CommentTree{
				ID:        "comment-123",
				Content:   "Test comment",
				CreatedAt: time.Now(),
				Children:  []*domain.CommentTree{},
			},
			expectSearch: []domain.Comment{
				{
					ID:        "comment-123",
					Content:   "Test comment",
					ParentID:  "",
					CreatedAt: time.Now(),
				},
			},
			expectError: false,
		},
		{
			name: "operations with errors",
			storage: &MockStorage{
				addCommentFunc: func(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
					return domain.Comment{}, errors.New("add comment failed")
				},
				getCommentsFunc: func(ctx context.Context, parentID string, asc bool, limit, offset int) (*domain.CommentTree, error) {
					return nil, errors.New("get comments failed")
				},
				deleteCommentsFunc: func(ctx context.Context, id string) error {
					return errors.New("delete comment failed")
				},
				searchCommentsFunc: func(ctx context.Context, query string, limit, offset int) ([]domain.Comment, error) {
					return nil, errors.New("search comments failed")
				},
			},
			comment: domain.Comment{
				Content:  "Test comment",
				ParentID: "",
			},
			parentID:     "",
			asc:          true,
			limit:        10,
			offset:       0,
			query:        "test",
			id:           "comment-123",
			expectAdd:    domain.Comment{},
			expectGet:    nil,
			expectSearch: nil,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Test AddComment
			comment, err := tt.storage.AddComment(ctx, tt.comment)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && comment.ID == "" {
				t.Error("Expected comment ID to be set")
			}

			// Test GetComments
			commentTree, err := tt.storage.GetComments(ctx, tt.parentID, tt.asc, tt.limit, tt.offset)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && commentTree == nil {
				t.Error("Expected comment tree but got nil")
			}

			// Test DeleteComments
			err = tt.storage.DeleteComments(ctx, tt.id)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Test SearchComments
			comments, err := tt.storage.SearchComments(ctx, tt.query, tt.limit, tt.offset)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && len(comments) == 0 {
				t.Error("Expected comments but got empty slice")
			}
		})
	}
}

func TestNewServer(t *testing.T) {
	storage := &MockStorage{}

	server := New(storage)

	if server == nil {
		t.Error("Expected server to be created, got nil")
		return
	}
	if server.st != storage {
		t.Error("Expected storage to be set correctly")
	}
	if server.r == nil {
		t.Error("Expected router to be initialized")
	}
}

func TestStorageInterfaceWithNestedComments(t *testing.T) {
	storage := &MockStorage{
		getCommentsFunc: func(ctx context.Context, parentID string, asc bool, limit, offset int) (*domain.CommentTree, error) {
			// Create a nested comment tree
			childComment := &domain.CommentTree{
				ID:        "child-123",
				Content:   "Child comment",
				CreatedAt: time.Now(),
				Children:  []*domain.CommentTree{},
			}

			parentComment := &domain.CommentTree{
				ID:        "parent-123",
				Content:   "Parent comment",
				CreatedAt: time.Now(),
				Children:  []*domain.CommentTree{childComment},
			}

			return parentComment, nil
		},
	}

	ctx := context.Background()
	commentTree, err := storage.GetComments(ctx, "", true, 10, 0)

	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}
	if commentTree == nil {
		t.Error("Expected comment tree but got nil")
		return
	}
	if len(commentTree.Children) == 0 {
		t.Error("Expected children but got empty slice")
	}
	if commentTree.Children[0].ID != "child-123" {
		t.Errorf("Expected child ID 'child-123', got '%s'", commentTree.Children[0].ID)
	}
}

func TestStorageInterfaceWithSearch(t *testing.T) {
	storage := &MockStorage{
		searchCommentsFunc: func(ctx context.Context, query string, limit, offset int) ([]domain.Comment, error) {
			// Return multiple comments that match the search query
			return []domain.Comment{
				{
					ID:        "comment-1",
					Content:   "First test comment",
					ParentID:  "",
					CreatedAt: time.Now(),
				},
				{
					ID:        "comment-2",
					Content:   "Second test comment",
					ParentID:  "",
					CreatedAt: time.Now(),
				},
			}, nil
		},
	}

	ctx := context.Background()
	comments, err := storage.SearchComments(ctx, "test", 10, 0)

	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}
	if len(comments) != 2 {
		t.Errorf("Expected 2 comments, got %d", len(comments))
	}
	if comments[0].Content != "First test comment" {
		t.Errorf("Expected first comment content 'First test comment', got '%s'", comments[0].Content)
	}
	if comments[1].Content != "Second test comment" {
		t.Errorf("Expected second comment content 'Second test comment', got '%s'", comments[1].Content)
	}
}

// --- HTTP handler tests ---

func TestGetComments_Handler(t *testing.T) {
    // happy path: returns empty JSON when no comments overall (nil tree)
    st := &MockStorage{
        getCommentsFunc: func(ctx context.Context, parentID string, asc bool, limit, offset int) (*domain.CommentTree, error) {
            return nil, nil
        },
    }
    s := New(st)
    ts := httptest.NewServer(s.r)
    defer ts.Close()

    resp, err := http.Get(ts.URL + "/comments")
    if err != nil {
        t.Fatalf("http get error: %v", err)
    }
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("expected 200, got %d", resp.StatusCode)
    }

    // invalid parent id should return 400
    resp, err = http.Get(ts.URL + "/comments?parent=not-a-uuid")
    if err != nil {
        t.Fatalf("http get error: %v", err)
    }
    if resp.StatusCode != http.StatusBadRequest {
        t.Fatalf("expected 400, got %d", resp.StatusCode)
    }
}

func TestGetComments_NotFound(t *testing.T) {
    st := &MockStorage{
        getCommentsFunc: func(ctx context.Context, parentID string, asc bool, limit, offset int) (*domain.CommentTree, error) {
            return nil, storage.ErrNotFound
        },
    }
    s := New(st)
    ts := httptest.NewServer(s.r)
    defer ts.Close()

    // when storage says not found for specific parent, expect 404
    id := uuid.NewString()
    resp, err := http.Get(ts.URL + "/comments?parent=" + url.QueryEscape(id))
    if err != nil {
        t.Fatalf("http get error: %v", err)
    }
    if resp.StatusCode != http.StatusNotFound {
        t.Fatalf("expected 404, got %d", resp.StatusCode)
    }
}

func TestPostComment_Handler(t *testing.T) {
    st := &MockStorage{
        addCommentFunc: func(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
            comment.ID = uuid.NewString()
            return comment, nil
        },
    }
    s := New(st)
    ts := httptest.NewServer(s.r)
    defer ts.Close()

    // invalid body
    resp, err := http.Post(ts.URL+"/comments", "application/json", bytes.NewBufferString("{"))
    if err != nil {
        t.Fatalf("post error: %v", err)
    }
    if resp.StatusCode != http.StatusBadRequest {
        t.Fatalf("expected 400, got %d", resp.StatusCode)
    }

    // invalid parent id
    body, _ := json.Marshal(domain.AddCommentRequest{Content: "hi", ParentID: "not-a-uuid"})
    resp, err = http.Post(ts.URL+"/comments", "application/json", bytes.NewReader(body))
    if err != nil {
        t.Fatalf("post error: %v", err)
    }
    if resp.StatusCode != http.StatusBadRequest {
        t.Fatalf("expected 400, got %d", resp.StatusCode)
    }

    // success
    body, _ = json.Marshal(domain.AddCommentRequest{Content: "hello"})
    resp, err = http.Post(ts.URL+"/comments", "application/json", bytes.NewReader(body))
    if err != nil {
        t.Fatalf("post error: %v", err)
    }
    if resp.StatusCode != http.StatusCreated {
        t.Fatalf("expected 201, got %d", resp.StatusCode)
    }
}

func TestDeleteComment_Handler(t *testing.T) {
    called := false
    st := &MockStorage{
        deleteCommentsFunc: func(ctx context.Context, id string) error {
            called = true
            return nil
        },
    }
    s := New(st)

    ts := httptest.NewServer(s.r)
    defer ts.Close()

    // invalid id
    req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/comments/not-a-uuid", nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        t.Fatalf("delete error: %v", err)
    }
    if resp.StatusCode != http.StatusBadRequest {
        t.Fatalf("expected 400, got %d", resp.StatusCode)
    }

    // valid id
    id := uuid.NewString()
    req, _ = http.NewRequest(http.MethodDelete, ts.URL+"/comments/"+id, nil)
    resp, err = http.DefaultClient.Do(req)
    if err != nil {
        t.Fatalf("delete error: %v", err)
    }
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("expected 200, got %d", resp.StatusCode)
    }
    if !called {
        t.Fatalf("expected delete to be called")
    }
}

func TestSearchComments_Handler(t *testing.T) {
    st := &MockStorage{
        searchCommentsFunc: func(ctx context.Context, query string, limit, offset int) ([]domain.Comment, error) {
            return []domain.Comment{{ID: "1", Content: "ok"}}, nil
        },
    }
    s := New(st)
    ts := httptest.NewServer(s.r)
    defer ts.Close()

    // missing q
    resp, err := http.Get(ts.URL + "/comments/search")
    if err != nil {
        t.Fatalf("get error: %v", err)
    }
    if resp.StatusCode != http.StatusBadRequest {
        t.Fatalf("expected 400, got %d", resp.StatusCode)
    }

    // ok
    resp, err = http.Get(ts.URL + "/comments/search?q=hello&page=1&limit=10")
    if err != nil {
        t.Fatalf("get error: %v", err)
    }
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("expected 200, got %d", resp.StatusCode)
    }
}
