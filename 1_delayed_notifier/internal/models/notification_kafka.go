package models

// NotificationKafka is the model for an output message to Kafka.
// It is used to notify other services that the attempt to send a notification has been made.
type NotificationKafka struct {
	NotificationID string `json:"notification_id"`
	Message        string `json:"message"`
}
