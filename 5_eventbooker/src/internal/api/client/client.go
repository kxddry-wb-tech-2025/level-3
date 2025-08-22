package client

import (
	"bytes"
	"encoding/json"
	"eventbooker/src/internal/domain"
	"fmt"
	"net/http"
)

var message = func(bookingID, eventID string) string {
	return fmt.Sprintf("Your booking %s for event %s has been canceled due to not paying in time. \nPlease contact support if you believe this is an error.", bookingID, eventID)
}

type Client struct {
	client *http.Client
	addr   string
	mp     map[string]string
}

func NewClient(addr string) *Client {
	return &Client{
		client: http.DefaultClient,
		addr:   addr,
		mp:     make(map[string]string),
	}
}

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

	return nil
}
