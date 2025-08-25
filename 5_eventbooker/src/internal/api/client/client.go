package client

import (
	"bytes"
	"context"
	"encoding/json"
	"eventbooker/src/internal/domain"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/kxddry/wbf/zlog"
)

// Client is the client for the event booking service.
type Client struct {
	client *http.Client
	addr   string
}

// New creates a new client.
// TODO: change this entire logic.
// It uses a cache with the given TTL.
// It is thread-safe.
//
// The cache is used as "bookingID" -> "notificationID" mapping.
//
// Do not forget to call Stop() on the client when you're done with it.
// Make sure that limit and TTL are enough to fit all your users.
func New(addr string) *Client {
	return &Client{
		client: http.DefaultClient,
		addr:   addr,
	}
}

// SendNotification sends a notification at a specific time.
// Returns the notification ID.
func (c *Client) SendNotification(ctx context.Context, notif domain.DelayedNotification) (string, error) {
	req := SendNotificationRequest{
		SendAt:    notif.SendAt,
		Channel:   "telegram",
		Recipient: strconv.FormatInt(notif.TelegramID, 10),
		Message:   fmt.Sprintf(domain.MessageCancelBookingTemplate, notif.BookingID, notif.EventID),
	}

	js, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	reqHt, err := http.NewRequestWithContext(ctx, http.MethodPost, c.addr+"/notify", bytes.NewBuffer(js))
	if err != nil {
		return "", err
	}
	reqHt.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(reqHt)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != SendSuccess {
		return "", fmt.Errorf("failed to send notification: %s", resp.Status)
	}
	slice, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var respBody SendNotificationResponse
	err = json.Unmarshal(slice, &respBody)
	if err != nil {
		return "", err
	}

	return respBody.ID, nil
}

// CancelNotification cancels a notification.
func (c *Client) CancelNotification(ctx context.Context, notificationID string) error {
	req := CancelNotificationRequest{
		ID: notificationID,
	}

	js, err := json.Marshal(req)
	if err != nil {
		return err
	}

	reqHt, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.addr+"/notify/"+notificationID, bytes.NewReader(js))
	if err != nil {
		return err
	}
	reqHt.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(reqHt)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != CancelSuccess {
		if resp.StatusCode == CancelNotFound {
			zlog.Logger.Err(err).Msg("notification was not found by the server")
		}
		return fmt.Errorf("failed to cancel notification: %s", resp.Status)
	}

	return nil
}
