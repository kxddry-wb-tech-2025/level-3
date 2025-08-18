package sender

import (
	"context"
	"delayed-notifier/internal/models"
	"errors"
	"testing"
	"time"
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

func TestSenderInterface(t *testing.T) {
	tests := []struct {
		name        string
		notification *models.Notification
		sendFunc    func(ctx context.Context, n *models.Notification) error
		expectError bool
	}{
		{
			name: "successful send",
			notification: &models.Notification{
				ID:        "test-123",
				Channel:   "email",
				Recipient: "test@example.com",
				Message:   "Test message",
				SendAt:    time.Now(),
				Status:    models.StatusScheduled,
			},
			sendFunc: func(ctx context.Context, n *models.Notification) error {
				return nil
			},
			expectError: false,
		},
		{
			name: "send error",
			notification: &models.Notification{
				ID:        "test-456",
				Channel:   "sms",
				Recipient: "+1234567890",
				Message:   "Test SMS",
				SendAt:    time.Now(),
				Status:    models.StatusScheduled,
			},
			sendFunc: func(ctx context.Context, n *models.Notification) error {
				return errors.New("send failed")
			},
			expectError: true,
		},
		{
			name: "unsupported channel",
			notification: &models.Notification{
				ID:        "test-789",
				Channel:   "unsupported",
				Recipient: "test",
				Message:   "Test message",
				SendAt:    time.Now(),
				Status:    models.StatusScheduled,
			},
			sendFunc: func(ctx context.Context, n *models.Notification) error {
				return ErrUnsupportedChannel
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			sender := &MockSender{sendFunc: tt.sendFunc}

			err := sender.Send(ctx, tt.notification)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestErrUnsupportedChannel(t *testing.T) {
	if ErrUnsupportedChannel.Error() != "unsupported channel" {
		t.Errorf("Expected error message 'unsupported channel', got: %s", ErrUnsupportedChannel.Error())
	}
}