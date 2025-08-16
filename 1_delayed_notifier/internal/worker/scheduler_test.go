package worker

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"delayed-notifier/internal/models"
)

type fakeStore struct {
	popDueResp map[string][]string
	getByID    map[string]*models.Notification
	saved      map[string]*models.Notification
	enqueued   []string
	retries    map[string]time.Time
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		popDueResp: make(map[string][]string),
		getByID:    make(map[string]*models.Notification),
		saved:      make(map[string]*models.Notification),
		retries:    make(map[string]time.Time),
	}
}

func (f *fakeStore) PopDue(ctx context.Context, which string, now time.Time, limit int64) ([]string, error) {
	ids := f.popDueResp[which]
	f.popDueResp[which] = nil
	return ids, nil
}

func (f *fakeStore) GetNotification(ctx context.Context, id string) (*models.Notification, error) {
	return f.getByID[id], nil
}

func (f *fakeStore) SaveNotification(ctx context.Context, n *models.Notification) error {
	cpy := *n
	f.saved[n.ID] = &cpy
	return nil
}

func (f *fakeStore) EnqueueNow(ctx context.Context, id string) error {
	f.enqueued = append(f.enqueued, id)
	return nil
}

func (f *fakeStore) AddToRetry(ctx context.Context, id string, when time.Time) error {
	f.retries[id] = when
	return nil
}

type fakePublisher struct {
	bodies    [][]byte
	shouldErr bool
}

func (p *fakePublisher) Publish(ctx context.Context, body []byte) error {
	p.bodies = append(p.bodies, body)
	if p.shouldErr {
		return context.DeadlineExceeded
	}
	return nil
}

func TestSchedulerPublishDueSuccess(t *testing.T) {
	store := newFakeStore()
	p := &fakePublisher{}
	s := NewScheduler(store, p)

	n := &models.Notification{ID: "n1", Status: models.StatusScheduled}
	store.getByID["n1"] = n
	store.popDueResp["due"] = []string{"n1"}

	s.publishDue(context.Background(), time.Now())

	if len(p.bodies) != 1 {
		t.Fatalf("expected 1 publish, got %d", len(p.bodies))
	}
	var out models.Notification
	_ = json.Unmarshal(p.bodies[0], &out)
	if out.ID != "n1" {
		t.Fatalf("unexpected published id: %s", out.ID)
	}
	saved := store.saved["n1"]
	if saved == nil || saved.Status != models.StatusQueued {
		t.Fatalf("expected saved status queued, got %#v", saved)
	}
}

func TestSchedulerPublishDueErrorReenqueue(t *testing.T) {
	store := newFakeStore()
	p := &fakePublisher{shouldErr: true}
	s := NewScheduler(store, p)

	n := &models.Notification{ID: "n1", Status: models.StatusScheduled}
	store.getByID["n1"] = n
	store.popDueResp["due"] = []string{"n1"}

	s.publishDue(context.Background(), time.Now())

	if len(store.enqueued) != 1 || store.enqueued[0] != "n1" {
		t.Fatalf("expected re-enqueue of n1, got %#v", store.enqueued)
	}
}

func TestSchedulerPublishRetryPath(t *testing.T) {
	store := newFakeStore()
	p := &fakePublisher{}
	s := NewScheduler(store, p)

	n := &models.Notification{ID: "n1", Status: models.StatusRetrying}
	store.getByID["n1"] = n
	store.popDueResp["retry"] = []string{"n1"}

	s.publishRetry(context.Background(), time.Now())

	if len(p.bodies) != 1 {
		t.Fatalf("expected 1 publish on retry, got %d", len(p.bodies))
	}
}

func TestSchedulerSkipsCancelled(t *testing.T) {
	store := newFakeStore()
	p := &fakePublisher{}
	s := NewScheduler(store, p)

	n := &models.Notification{ID: "n1", Status: models.StatusCancelled}
	store.getByID["n1"] = n
	store.popDueResp["due"] = []string{"n1"}

	s.publishDue(context.Background(), time.Now())

	if len(p.bodies) != 0 {
		t.Fatalf("expected no publish for cancelled, got %d", len(p.bodies))
	}
	if _, ok := store.saved["n1"]; ok {
		t.Fatalf("expected not to save cancelled notification")
	}
}

func TestSchedulerGetNotificationError(t *testing.T) {
	store := newFakeStore()
	p := &fakePublisher{}
	s := NewScheduler(store, p)

	store.popDueResp["due"] = []string{"missing"}

	s.publishDue(context.Background(), time.Now())

	if len(p.bodies) != 0 {
		t.Fatalf("expected no publish when get fails")
	}
}
