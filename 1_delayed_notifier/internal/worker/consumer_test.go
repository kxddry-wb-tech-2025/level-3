package worker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"delayed-notifier/internal/models"
)

type fakeDelivery struct {
	body    []byte
	acked   bool
	nacked  bool
	requeue bool
}

func (d *fakeDelivery) Body() []byte            { return d.body }
func (d *fakeDelivery) Ack() error              { d.acked = true; return nil }
func (d *fakeDelivery) Nack(requeue bool) error { d.nacked = true; d.requeue = requeue; return nil }

type chanQueue struct{ ch chan models.Delivery }

func (q *chanQueue) Consume(ctx context.Context) (<-chan models.Delivery, error) { return q.ch, nil }

type fakeSender struct{ err error }

func (s *fakeSender) Send(ctx context.Context, n models.Notification) error { return s.err }

type fakeStoreC struct {
	saved   map[string]*models.Notification
	retried map[string]time.Time
}

func newFakeStoreC() *fakeStoreC {
	return &fakeStoreC{saved: map[string]*models.Notification{}, retried: map[string]time.Time{}}
}
func (s *fakeStoreC) GetNotification(ctx context.Context, id string) (*models.Notification, error) {
	return nil, nil
}
func (s *fakeStoreC) SaveNotification(ctx context.Context, n *models.Notification) error {
	c := *n
	s.saved[n.ID] = &c
	return nil
}
func (s *fakeStoreC) AddToRetry(ctx context.Context, id string, when time.Time) error {
	s.retried[id] = when
	return nil
}

func TestConsumerProcessSuccess(t *testing.T) {
	store := newFakeStoreC()
	q := &chanQueue{ch: make(chan models.Delivery, 1)}
	sender := &fakeSender{}
	c := NewConsumer(store, q, sender)

	n := models.Notification{ID: "a1", Channel: "telegram", Recipient: "1", Message: "hi"}
	bytes, _ := json.Marshal(n)
	q.ch <- &fakeDelivery{body: bytes}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { c.Run(ctx) }()

	// allow processing
	time.Sleep(50 * time.Millisecond)

	saved := store.saved["a1"]
	if saved == nil || saved.Status != models.StatusSent {
		t.Fatalf("expected sent status, got %#v", saved)
	}
}

func TestConsumerProcessFailureRetry(t *testing.T) {
	store := newFakeStoreC()
	q := &chanQueue{ch: make(chan models.Delivery, 1)}
	sender := &fakeSender{err: errors.New("boom")}
	c := NewConsumer(store, q, sender)

	n := models.Notification{ID: "a1", Channel: "telegram", Recipient: "1", Message: "hi"}
	bytes, _ := json.Marshal(n)
	q.ch <- &fakeDelivery{body: bytes}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { c.Run(ctx) }()

	// allow processing (include short in-process retries)
	time.Sleep(300 * time.Millisecond)

	saved := store.saved["a1"]
	if saved == nil || saved.Status != models.StatusRetrying {
		t.Fatalf("expected retrying status, got %#v", saved)
	}
	if _, ok := store.retried["a1"]; !ok {
		t.Fatalf("expected AddToRetry to be called")
	}
}

type flakySender struct{ fails int }

func (s *flakySender) Send(ctx context.Context, n models.Notification) error {
	if s.fails > 0 {
		s.fails--
		return errors.New("temp fail")
	}
	return nil
}

func TestConsumerTransientFailureThenSuccess(t *testing.T) {
	store := newFakeStoreC()
	q := &chanQueue{ch: make(chan models.Delivery, 1)}
	sender := &flakySender{fails: 2}
	c := NewConsumer(store, q, sender)

	n := models.Notification{ID: "t1", Channel: "telegram", Recipient: "1", Message: "hi"}
	bytes, _ := json.Marshal(n)
	q.ch <- &fakeDelivery{body: bytes}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { c.Run(ctx) }()

	time.Sleep(200 * time.Millisecond)
	saved := store.saved["t1"]
	if saved == nil || saved.Status != models.StatusSent {
		t.Fatalf("expected sent after transient failures, got %#v", saved)
	}
	if saved.NextAttemptAt != nil {
		t.Fatalf("expected no next attempt after success")
	}
}

type errQueue struct{}

func (e *errQueue) Consume(ctx context.Context) (<-chan models.Delivery, error) {
	return nil, errors.New("boom")
}

func TestConsumerQueueConsumeError(t *testing.T) {
	store := newFakeStoreC()
	sender := &fakeSender{}
	c := NewConsumer(store, &errQueue{}, sender)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Should return quickly without panic
	c.Run(ctx)
}

func TestConsumerBadPayloadNack(t *testing.T) {
	store := newFakeStoreC()
	q := &chanQueue{ch: make(chan models.Delivery, 1)}
	sender := &fakeSender{}
	c := NewConsumer(store, q, sender)

	// push invalid JSON
	fd := &fakeDelivery{body: []byte("{not-json}")}
	q.ch <- fd

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { c.Run(ctx) }()

	time.Sleep(50 * time.Millisecond)
	if !fd.nacked || fd.requeue != false {
		t.Fatalf("expected Nack(false) on bad payload")
	}
	if len(store.saved) != 0 {
		t.Fatalf("did not expect any save on bad payload")
	}
}

func TestConsumerCancelledAck(t *testing.T) {
	store := newFakeStoreC()
	q := &chanQueue{ch: make(chan models.Delivery, 1)}
	sender := &fakeSender{}
	c := NewConsumer(store, q, sender)

	n := models.Notification{ID: "c1", Status: models.StatusCancelled}
	bytes, _ := json.Marshal(n)
	fd := &fakeDelivery{body: bytes}
	q.ch <- fd

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { c.Run(ctx) }()

	time.Sleep(50 * time.Millisecond)
	if !fd.acked {
		t.Fatalf("expected Ack for cancelled notification")
	}
	if len(store.saved) != 0 {
		t.Fatalf("did not expect save for cancelled notification")
	}
}
