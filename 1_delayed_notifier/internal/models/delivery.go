package models

// Delivery is a transport-agnostic message wrapper used across queue and workers.
type Delivery interface {
	Body() []byte
	Ack() error
	Nack(requeue bool) error
}
