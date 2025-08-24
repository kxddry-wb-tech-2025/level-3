package kafka

// NotificationEvent is read from Kafka sent by the delayed notifier microservice
// It is used to cancel a booking. (because the user didn't pay for the booking in time)
type NotificationEvent struct {
	NotificationID string `json:"notification_id"`
	Message        string `json:"message"`
}

type ids struct {
	EventID   string `json:"event_id"`
	BookingID string `json:"booking_id"`
}
