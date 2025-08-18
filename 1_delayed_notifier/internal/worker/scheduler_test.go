package worker

import (
	"context"
	"delayed-notifier/internal/models"
	"errors"
	"testing"
	"time"
)

// MockNotificationStore is a mock implementation of NotificationStore interface
type MockNotificationStore struct {
	popDueFunc        func(ctx context.Context, which string, now time.Time, limit int64) ([]string, error)
	getNotificationFunc func(ctx context.Context, id string) (*models.Notification, error)
	saveNotificationFunc func(ctx context.Context, n *models.Notification) error
	enqueueNowFunc    func(ctx context.Context, id string) error
	addToRetryFunc    func(ctx context.Context, id string, when time.Time) error
}

func (m *MockNotificationStore) PopDue(ctx context.Context, which string, now time.Time, limit int64) ([]string, error) {
	if m.popDueFunc != nil {
		return m.popDueFunc(ctx, which, now, limit)
	}
	return nil, nil
}

func (m *MockNotificationStore) GetNotification(ctx context.Context, id string) (*models.Notification, error) {
	if m.getNotificationFunc != nil {
		return m.getNotificationFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockNotificationStore) SaveNotification(ctx context.Context, n *models.Notification) error {
	if m.saveNotificationFunc != nil {
		return m.saveNotificationFunc(ctx, n)
	}
	return nil
}

func (m *MockNotificationStore) EnqueueNow(ctx context.Context, id string) error {
	if m.enqueueNowFunc != nil {
		return m.enqueueNowFunc(ctx, id)
	}
	return nil
}

func (m *MockNotificationStore) AddToRetry(ctx context.Context, id string, when time.Time) error {
	if m.addToRetryFunc != nil {
		return m.addToRetryFunc(ctx, id, when)
	}
	return nil
}

// MockPublisher is a mock implementation of Publisher interface
type MockPublisher struct {
	publishFunc func(ctx context.Context, body []byte) error
}

func (m *MockPublisher) Publish(ctx context.Context, body []byte) error {
	if m.publishFunc != nil {
		return m.publishFunc(ctx, body)
	}
	return nil
}

func TestNotificationStoreInterface(t *testing.T) {
	tests := []struct {
		name           string
		store          *MockNotificationStore
		expectPopDue   []string
		expectPopError bool
		expectGet      *models.Notification
		expectGetError bool
		expectSave     error
		expectEnqueue  error
		expectRetry    error
	}{
		{
			name: "successful operations",
			store: &MockNotificationStore{
				popDueFunc: func(ctx context.Context, which string, now time.Time, limit int64) ([]string, error) {
					return []string{"test-1", "test-2"}, nil
				},
				getNotificationFunc: func(ctx context.Context, id string) (*models.Notification, error) {
					return &models.Notification{
						ID:        id,
						Channel:   "email",
						Recipient: "test@example.com",
						Message:   "test message",
						Status:    models.StatusScheduled,
					}, nil
				},
				saveNotificationFunc: func(ctx context.Context, n *models.Notification) error {
					return nil
				},
				enqueueNowFunc: func(ctx context.Context, id string) error {
					return nil
				},
				addToRetryFunc: func(ctx context.Context, id string, when time.Time) error {
					return nil
				},
			},
			expectPopDue:   []string{"test-1", "test-2"},
			expectPopError: false,
			expectGet: &models.Notification{
				ID:        "test-123",
				Channel:   "email",
				Recipient: "test@example.com",
				Message:   "test message",
				Status:    models.StatusScheduled,
			},
			expectGetError: false,
			expectSave:     nil,
			expectEnqueue:  nil,
			expectRetry:    nil,
		},
		{
			name: "operations with errors",
			store: &MockNotificationStore{
				popDueFunc: func(ctx context.Context, which string, now time.Time, limit int64) ([]string, error) {
					return nil, errors.New("pop due failed")
				},
				getNotificationFunc: func(ctx context.Context, id string) (*models.Notification, error) {
					return nil, errors.New("get notification failed")
				},
				saveNotificationFunc: func(ctx context.Context, n *models.Notification) error {
					return errors.New("save failed")
				},
				enqueueNowFunc: func(ctx context.Context, id string) error {
					return errors.New("enqueue failed")
				},
				addToRetryFunc: func(ctx context.Context, id string, when time.Time) error {
					return errors.New("retry failed")
				},
			},
			expectPopDue:   nil,
			expectPopError: true,
			expectGet:      nil,
			expectGetError: true,
			expectSave:     errors.New("save failed"),
			expectEnqueue:  errors.New("enqueue failed"),
			expectRetry:    errors.New("retry failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			now := time.Now()

			// Test PopDue
			ids, err := tt.store.PopDue(ctx, "due", now, 100)
			if tt.expectPopError && err == nil {
				t.Errorf("Expected PopDue error but got none")
			}
			if !tt.expectPopError && err != nil {
				t.Errorf("Expected no PopDue error but got: %v", err)
			}
			if !tt.expectPopError && len(ids) != len(tt.expectPopDue) {
				t.Errorf("Expected %d IDs, got %d", len(tt.expectPopDue), len(ids))
			}

			// Test GetNotification
			notification, err := tt.store.GetNotification(ctx, "test-123")
			if tt.expectGetError && err == nil {
				t.Errorf("Expected GetNotification error but got none")
			}
			if !tt.expectGetError && err != nil {
				t.Errorf("Expected no GetNotification error but got: %v", err)
			}
			if !tt.expectGetError && notification == nil {
				t.Error("Expected notification but got nil")
			}

			// Test SaveNotification
			if tt.expectGet != nil {
				err = tt.store.SaveNotification(ctx, tt.expectGet)
				if (err == nil) != (tt.expectSave == nil) {
					t.Errorf("Expected save error %v, got %v", tt.expectSave, err)
				}
			}

			// Test EnqueueNow
			err = tt.store.EnqueueNow(ctx, "test-123")
			if (err == nil) != (tt.expectEnqueue == nil) {
				t.Errorf("Expected enqueue error %v, got %v", tt.expectEnqueue, err)
			}

			// Test AddToRetry
			err = tt.store.AddToRetry(ctx, "test-123", now)
			if (err == nil) != (tt.expectRetry == nil) {
				t.Errorf("Expected retry error %v, got %v", tt.expectRetry, err)
			}
		})
	}
}

func TestPublisherInterface(t *testing.T) {
	tests := []struct {
		name         string
		publisher    *MockPublisher
		body         []byte
		expectError  bool
	}{
		{
			name: "successful publish",
			publisher: &MockPublisher{
				publishFunc: func(ctx context.Context, body []byte) error {
					return nil
				},
			},
			body:        []byte("test message"),
			expectError: false,
		},
		{
			name: "publish error",
			publisher: &MockPublisher{
				publishFunc: func(ctx context.Context, body []byte) error {
					return errors.New("publish failed")
				},
			},
			body:        []byte("test message"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			err := tt.publisher.Publish(ctx, tt.body)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestNewScheduler(t *testing.T) {
	store := &MockNotificationStore{}
	publisher := &MockPublisher{}

	scheduler := NewScheduler(store, publisher)

	if scheduler == nil {
		t.Error("Expected scheduler to be created, got nil")
	}
	if scheduler.store != store {
		t.Error("Expected store to be set correctly")
	}
	if scheduler.q != publisher {
		t.Error("Expected publisher to be set correctly")
	}
}
