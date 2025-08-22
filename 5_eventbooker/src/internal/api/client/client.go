package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"eventbooker/src/internal/domain"
	"eventbooker/src/internal/storage"
	"eventbooker/src/internal/storage/cache"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kxddry/wbf/zlog"
)

var message = func(bookingID, eventID string) string {
	return fmt.Sprintf("Your booking %s for event %s has been canceled due to not paying in time. \nPlease contact support if you believe this is an error.", bookingID, eventID)
}

// Client is the client for the event booking service.
type Client struct {
	client *http.Client
	addr   string
	mp     *cache.Cache
}

// NewClient creates a new client.
// It uses a cache with the given TTL.
// It is thread-safe.
//
// The cache is used as "bookingID" -> "notificationID" mapping.
//
// Do not forget to call Stop() on the client when you're done with it.
// Make sure that limit and TTL are enough to fit all your users.
// Sure, we could use Postgres to store Telegram IDs and Booking IDs, but it's a bit too much for this project.
func NewClient(addr string, cacheTTL time.Duration, cacheLimit int) *Client {
	return &Client{
		client: http.DefaultClient,
		addr:   addr,
		mp:     cache.NewCache(cacheTTL, cacheLimit),
	}
}

// SendDelayed sends a delayed notification at a specific time.
func (c *Client) SendDelayed(notif domain.DelayedNotification) error {
	req := SendNotificationRequest{
		SendAt:    notif.SendAt,
		Channel:   "telegram",
		Recipient: notif.TelegramID,
		Message:   message(notif.BookingID, notif.EventID),
	}

	js, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := c.client.Post(c.addr+"/notify", "application/json", bytes.NewBuffer(js))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != SendSuccess {
		return fmt.Errorf("failed to send notification: %s", resp.Status)
	}
	slice, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var respBody SendNotificationResponse
	err = json.Unmarshal(slice, &respBody)
	if err != nil {
		return err
	}

	// save notificationID to cache
	c.mp.Set(notif.BookingID, respBody.ID)

	return nil
}

// CancelDelayed cancels a delayed notification.
func (c *Client) CancelDelayed(bookingID string) error {
	recipient, err := c.mp.Get(bookingID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			zlog.Logger.Err(err).Msg("DEVELOPERS MISTAKE: recipient not found in cache")
		}
		return err
	}

	str, ok := recipient.(string)
	if !ok {
		zlog.Logger.Err(err).Msg("DEVELOPERS MISTAKE: recipient is not a string")
		return fmt.Errorf("recipient is not a string")
	}

	req := CancelNotificationRequest{
		ID: str,
	}

	js, err := json.Marshal(req)
	if err != nil {
		return err
	}

	reqHt, err := http.NewRequest(http.MethodDelete, c.addr+"/notify/"+bookingID, bytes.NewReader(js))
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

	c.mp.Remove(bookingID)

	return nil
}
