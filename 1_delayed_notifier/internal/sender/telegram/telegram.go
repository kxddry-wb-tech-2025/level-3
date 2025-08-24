package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"delayed-notifier/internal/models"
	"delayed-notifier/internal/sender"

	"github.com/kxddry/wbf/zlog"
)

// Sender sends messages via Telegram Bot API.
type Sender struct {
	client *http.Client
	apiURL string
}

// NewSender creates a new Telegram sender with the provided bot token and timeout.
func NewSender(botToken string, timeout time.Duration) *Sender {
	return &Sender{
		client: &http.Client{Timeout: timeout},
		apiURL: fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken),
	}
}

type tgReq struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// Send delivers a notification text to the specified chat via Telegram.
func (t *Sender) Send(ctx context.Context, n models.Notification) error {
	log := zlog.Logger.With().Str("component", "telegram").Logger()
	if n.Channel != "telegram" {
		return sender.ErrUnsupportedChannel
	}
	if n.Recipient == "" {
		return errors.New("empty recipient")
	}
	body, _ := json.Marshal(tgReq{ChatID: n.Recipient, Text: n.Message})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.apiURL, bytes.NewReader(body))
	if err != nil {
		log.Error().Err(err).Msg("failed to create request")
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("failed to do request")
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Error().Int("status_code", resp.StatusCode).Msg("failed to send message")
		return fmt.Errorf("telegram http status %d", resp.StatusCode)
	}
	return nil
}
