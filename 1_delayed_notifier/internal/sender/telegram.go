package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"delayed-notifier/internal/models"
)

type TelegramSender struct {
	client *http.Client
	apiURL string
}

func NewTelegramSender(botToken string, timeout time.Duration) *TelegramSender {
	return &TelegramSender{
		client: &http.Client{Timeout: timeout},
		apiURL: fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken),
	}
}

type tgReq struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

func (t *TelegramSender) Send(ctx context.Context, n *models.Notification) error {
	if n.Channel != "telegram" {
		return ErrUnsupportedChannel
	}
	if n.Recipient == "" {
		return errors.New("empty recipient")
	}
	body, _ := json.Marshal(tgReq{ChatID: n.Recipient, Text: n.Message})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.apiURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telegram http status %d", resp.StatusCode)
	}
	return nil
}
