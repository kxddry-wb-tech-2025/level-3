package telegram

import (
	"bytes"
	"context"
	"delayed-notifier/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const baseURL = "https://api.telegram.org/bot"

// Sender is a structure representing a Telegram bot that can send messages.
type Sender struct {
	token string
	http  *http.Client
}

// New initializes Sender
func New(token string, timeout time.Duration) (*Sender, error) {
	if token == "" {
		return nil, errors.New("telegram token is required")
	}

	req, err := http.NewRequest(http.MethodGet, baseURL+token+"/getMe", nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	return &Sender{
		token: token,
		http: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// Name returns the name of the Sender.
func (s *Sender) Name() string { return models.ChannelTelegram }

// Send sends the notification and returns an error.
func (s *Sender) Send(ctx context.Context, n models.Notification) error {
	url := fmt.Sprintf("%s%s/sendMessage", baseURL, s.token)
	payload := map[string]string{
		"chat_id": n.Recipient, "text": n.Message,
	}
	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := s.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("telegram status %d", resp.StatusCode)
	}
	return nil
}
