package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"delayed-notifier/internal/models"
)

// MockSender is a mock implementation of the Sender interface for testing
type MockSender struct {
	sendFunc func(ctx context.Context, n *models.Notification) error
}

func (m *MockSender) Send(ctx context.Context, n *models.Notification) error {
	if m.sendFunc != nil {
		return m.sendFunc(ctx, n)
	}
	return nil
}

// MockConsumerQueue is a mock implementation of ConsumerQueue interface
type MockConsumerQueue struct {
	consumeFunc func(ctx context.Context) (<-chan models.Delivery, error)
}

func (m *MockConsumerQueue) Consume(ctx context.Context) (<-chan models.Delivery, error) {
	if m.consumeFunc != nil {
		return m.consumeFunc(ctx)
	}
	return nil, nil
}

// MockDelivery is a mock implementation of models.Delivery interface
type MockDelivery struct {
	bodyFunc  func() []byte
	ackFunc   func() error
	nackFunc  func(requeue bool) error
}

func (m *MockDelivery) Body() []byte {
	if m.bodyFunc != nil {
		return m.bodyFunc()
	}
	return []byte{}
}

func (m *MockDelivery) Ack() error {
	if m.ackFunc != nil {
		return m.ackFunc()
	}
	return nil
}

func (m *MockDelivery) Nack(requeue bool) error {
	if m.nackFunc != nil {
		return m.nackFunc(requeue)
	}
	return nil
}

// MockStorageAccess is a mock implementation of storageAccess interface
type MockStorageAccess struct {
	getNotificationFunc func(ctx context.Context, id string) (*models.Notification, error)
	saveNotificationFunc func(ctx context.Context, n *models.Notification) error
	addToRetryFunc func(ctx context.Context, id string, when time.Time) error
}

func (m *MockStorageAccess) GetNotification(ctx context.Context, id string) (*models.Notification, error) {
	if m.getNotificationFunc != nil {
		return m.getNotificationFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockStorageAccess) SaveNotification(ctx context.Context, n *models.Notification) error {
	if m.saveNotificationFunc != nil {
		return m.saveNotificationFunc(ctx, n)
	}
	return nil
}

func (m *MockStorageAccess) AddToRetry(ctx context.Context, id string, when time.Time) error {
	if m.addToRetryFunc != nil {
		return m.addToRetryFunc(ctx, id, when)
	}
	return nil
}

func TestConsumerQueueInterface(t *testing.T) {
	tests := []struct {
		name        string
		consumeFunc func(ctx context.Context) (<-chan models.Delivery, error)
		expectError bool
	}{
		{
			name: "successful consume",
			consumeFunc: func(ctx context.Context) (<-chan models.Delivery, error) {
				ch := make(chan models.Delivery, 1)
				ch <- &MockDelivery{
					bodyFunc: func() []byte {
						return []byte(`{"id":"test-123","channel":"email","recipient":"test@example.com","message":"test","status":"scheduled"}`)
					},
				}
				close(ch)
				return ch, nil
			},
			expectError: false,
		},
		{
			name: "consume error",
			consumeFunc: func(ctx context.Context) (<-chan models.Delivery, error) {
				return nil, errors.New("consume failed")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			queue := &MockConsumerQueue{consumeFunc: tt.consumeFunc}

			_, err := queue.Consume(ctx)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestDeliveryInterface(t *testing.T) {
	tests := []struct {
		name       string
		delivery   *MockDelivery
		expectBody []byte
		expectAck  error
		expectNack error
	}{
		{
			name: "successful delivery",
			delivery: &MockDelivery{
				bodyFunc: func() []byte {
					return []byte("test body")
				},
				ackFunc: func() error {
					return nil
				},
				nackFunc: func(requeue bool) error {
					return nil
				},
			},
			expectBody: []byte("test body"),
			expectAck:  nil,
			expectNack: nil,
		},
		{
			name: "delivery with errors",
			delivery: &MockDelivery{
				bodyFunc: func() []byte {
					return []byte("error body")
				},
				ackFunc: func() error {
					return errors.New("ack failed")
				},
				nackFunc: func(requeue bool) error {
					return errors.New("nack failed")
				},
			},
			expectBody: []byte("error body"),
			expectAck:  errors.New("ack failed"),
			expectNack: errors.New("nack failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := tt.delivery.Body()
			if string(body) != string(tt.expectBody) {
				t.Errorf("Expected body %s, got %s", tt.expectBody, body)
			}

			err := tt.delivery.Ack()
			if (err == nil) != (tt.expectAck == nil) {
				t.Errorf("Expected ack error %v, got %v", tt.expectAck, err)
			}

			err = tt.delivery.Nack(false)
			if (err == nil) != (tt.expectNack == nil) {
				t.Errorf("Expected nack error %v, got %v", tt.expectNack, err)
			}
		})
	}
}

func TestNewConsumer(t *testing.T) {
	store := &MockStorageAccess{}
	queue := &MockConsumerQueue{}
	sender := &MockSender{}

	consumer := NewConsumer(store, queue, sender)

	if consumer == nil {
		t.Error("Expected consumer to be created, got nil")
	}
	if consumer.store != store {
		t.Error("Expected store to be set correctly")
	}
	if consumer.q != queue {
		t.Error("Expected queue to be set correctly")
	}
	if consumer.sender != sender {
		t.Error("Expected sender to be set correctly")
	}
}
