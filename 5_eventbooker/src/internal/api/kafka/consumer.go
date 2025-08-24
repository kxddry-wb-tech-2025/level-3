package kafka

import (
	"context"
	"encoding/json"
	"eventbooker/src/internal/domain"
	"fmt"

	"github.com/kxddry/wbf/zlog"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	r *kafka.Reader
}

func NewConsumer(ctx context.Context, brokers []string, topic string, groupID string) (*Consumer, error) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	conn, err := kafka.DialLeader(ctx, "tcp", brokers[0], topic, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to leader for topic %q: %w", topic, err)
	}
	_ = conn.Close()

	return &Consumer{r: r}, nil
}

func (c *Consumer) Close() error {
	return c.r.Close()
}

func (c *Consumer) Messages(ctx context.Context) <-chan domain.CancelBookingEvent {
	out := make(chan domain.CancelBookingEvent)
	go func() {
		log := zlog.Logger.With().Str("component", "kafka_consumer").Logger()
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			msg, err := c.r.ReadMessage(ctx)
			if err != nil {
				log.Error().Err(err).Msg("failed to read message")
				continue
			}
			var nk NotificationEvent
			if err := json.Unmarshal(msg.Value, &nk); err != nil {
				log.Error().Err(err).Msg("failed to unmarshal message")
				continue
			}
			log.Info().Msgf("notification sent: %+v", nk)
			event, err := c.msgToEvent(nk)
			if err != nil {
				log.Error().Err(err).Msg("failed to convert message to event")
				continue
			}
			out <- event
		}
	}()

	return out
}

func (c *Consumer) msgToEvent(msg NotificationEvent) (domain.CancelBookingEvent, error) {
	var event domain.CancelBookingEvent

	ids, err := messageToID(msg)
	if err != nil {
		return domain.CancelBookingEvent{}, err
	}

	event.NotificationID = msg.NotificationID
	event.BookingID = ids.BookingID
	event.EventID = ids.EventID

	return event, nil
}

func messageToID(msg NotificationEvent) (ids, error) {
	var bookingID, eventID string

	if _, err := fmt.Sscanf(msg.Message, domain.MessageCancelBookingTemplate, &bookingID, &eventID); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to scan message")
		return ids{}, err
	}

	return ids{
		EventID:   eventID,
		BookingID: bookingID,
	}, nil
}
