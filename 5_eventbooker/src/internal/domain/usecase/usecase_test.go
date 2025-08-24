package usecase

import (
	"context"
	"errors"
	"eventbooker/src/internal/domain"
	"testing"
	"time"
)

type stubTx struct {
	createEventFunc           func(ctx context.Context, event domain.CreateEventRequest) (string, error)
	updateEventFunc           func(ctx context.Context, event domain.Event) error
	getEventFunc              func(ctx context.Context, eventID string) (domain.Event, error)
	bookFunc                  func(ctx context.Context, eventID, userID string, paymentDeadline time.Time) (string, error)
	getBookingFunc            func(ctx context.Context, bookingID string) (domain.Booking, error)
	confirmFunc               func(ctx context.Context, bookingID string) (string, error)
	bookingSetDecrementedFunc func(ctx context.Context, bookingID string, decremented bool) error
	bookingSetStatusFunc      func(ctx context.Context, bookingID string, status string) error
	addNotificationFunc       func(ctx context.Context, notif domain.DelayedNotification) error
	getNotificationIDFunc     func(ctx context.Context, bookingID string) (string, error)
}

func (s *stubTx) CreateEvent(ctx context.Context, event domain.CreateEventRequest) (string, error) {
	return s.createEventFunc(ctx, event)
}
func (s *stubTx) UpdateEvent(ctx context.Context, event domain.Event) error { return s.updateEventFunc(ctx, event) }
func (s *stubTx) GetEvent(ctx context.Context, eventID string) (domain.Event, error) { return s.getEventFunc(ctx, eventID) }
func (s *stubTx) Book(ctx context.Context, eventID, userID string, paymentDeadline time.Time) (string, error) {
	return s.bookFunc(ctx, eventID, userID, paymentDeadline)
}
func (s *stubTx) GetBooking(ctx context.Context, bookingID string) (domain.Booking, error) {
	return s.getBookingFunc(ctx, bookingID)
}
func (s *stubTx) Confirm(ctx context.Context, bookingID string) (string, error) {
	return s.confirmFunc(ctx, bookingID)
}
func (s *stubTx) BookingSetDecremented(ctx context.Context, bookingID string, decremented bool) error {
	return s.bookingSetDecrementedFunc(ctx, bookingID, decremented)
}
func (s *stubTx) BookingSetStatus(ctx context.Context, bookingID string, status string) error {
	return s.bookingSetStatusFunc(ctx, bookingID, status)
}
func (s *stubTx) AddNotification(ctx context.Context, notif domain.DelayedNotification) error {
	return s.addNotificationFunc(ctx, notif)
}
func (s *stubTx) GetNotificationID(ctx context.Context, bookingID string) (string, error) {
	return s.getNotificationIDFunc(ctx, bookingID)
}
func (s *stubTx) Commit() error   { return nil }
func (s *stubTx) Rollback() error { return nil }

type stubTxManager struct{ do func(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error }

func (m *stubTxManager) Do(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error {
	return m.do(ctx, fn)
}

type stubNotif struct {
	send func(ctx context.Context, notif domain.DelayedNotification) (string, error)
	canc func(ctx context.Context, id string) error
}

func (n *stubNotif) SendNotification(ctx context.Context, notif domain.DelayedNotification) (string, error) {
	return n.send(ctx, notif)
}
func (n *stubNotif) CancelNotification(ctx context.Context, notificationID string) error { return n.canc(ctx, notificationID) }

type stubCancel struct{ ch chan domain.CancelBookingEvent }

func (c *stubCancel) Messages(ctx context.Context) <-chan domain.CancelBookingEvent { return c.ch }

func TestCreateEvent_SuccessAndValidation(t *testing.T) {
	ctx := context.Background()
	future := time.Now().Add(24 * time.Hour)
	uc := &Usecase{
		cs: &stubCancel{ch: make(chan domain.CancelBookingEvent)},
		nf: &stubNotif{send: func(ctx context.Context, notif domain.DelayedNotification) (string, error) { return "", nil }, canc: func(ctx context.Context, id string) error { return nil }},
		storage: &stubTxManager{do: func(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error {
			return fn(ctx, &stubTx{createEventFunc: func(ctx context.Context, event domain.CreateEventRequest) (string, error) {
				return "event-id", nil
			}})
		}},
	}

	// success
	resp := uc.CreateEvent(ctx, domain.CreateEventRequest{Name: "n", Capacity: 10, Date: future, PaymentTTL: time.Hour})
	if resp.Error != "" || resp.ID != "event-id" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	// validation: past date
	past := time.Now().Add(-time.Hour)
	resp = uc.CreateEvent(ctx, domain.CreateEventRequest{Name: "n", Capacity: 10, Date: past, PaymentTTL: time.Hour})
	if resp.Error == "" {
		t.Fatalf("expected validation error for past date")
	}
}

func TestGetEvent_SuccessAndError(t *testing.T) {
	ctx := context.Background()
	date := time.Now().Add(48 * time.Hour)
	uc := &Usecase{storage: &stubTxManager{do: func(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error {
		return fn(ctx, &stubTx{getEventFunc: func(ctx context.Context, eventID string) (domain.Event, error) {
			return domain.Event{ID: eventID, Name: "name", Capacity: 100, Available: 90, Date: &date, PaymentTTL: time.Hour}, nil
		}})
	}}}
	resp := uc.GetEvent(ctx, "event-1")
	if resp.Error != "" || resp.Name != "name" || resp.Available != 90 {
		t.Fatalf("unexpected response: %+v", resp)
	}

	uc.storage = &stubTxManager{do: func(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error {
		return errors.New("db err")
	}}
	resp = uc.GetEvent(ctx, "event-1")
	if resp.Error == "" {
		t.Fatalf("expected error in response")
	}
}

func TestBook_Flows(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	future := now.Add(2 * time.Hour)
	event := domain.Event{ID: "e1", Name: "n", Capacity: 100, Available: 5, Date: &future, PaymentTTL: time.Hour}

	// success path, notification succeeds and decrements
	uc := &Usecase{
		nf: &stubNotif{send: func(ctx context.Context, notif domain.DelayedNotification) (string, error) { return "notif-1", nil }, canc: func(ctx context.Context, id string) error { return nil }},
		storage: &stubTxManager{do: func(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error {
			return fn(ctx, &stubTx{
				getEventFunc: func(ctx context.Context, eventID string) (domain.Event, error) { return event, nil },
				bookFunc: func(ctx context.Context, eventID, userID string, paymentDeadline time.Time) (string, error) { return "b1", nil },
				updateEventFunc: func(ctx context.Context, event domain.Event) error { return nil },
				bookingSetDecrementedFunc: func(ctx context.Context, bookingID string, decremented bool) error { return nil },
				addNotificationFunc: func(ctx context.Context, notif domain.DelayedNotification) error { return nil },
			})
		}},
	}
	resp := uc.Book(ctx, "e1", "u1")
	if resp.Error != "" || resp.ID != "b1" || resp.PaymentDeadline == nil {
		t.Fatalf("unexpected response: %+v", resp)
	}

	// event full
	eventFull := event
	eventFull.Available = 0
	uc.storage = &stubTxManager{do: func(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error {
		return fn(ctx, &stubTx{getEventFunc: func(ctx context.Context, eventID string) (domain.Event, error) { return eventFull, nil }})
	}}
	resp = uc.Book(ctx, "e1", "u1")
	if resp.Error == "" {
		t.Fatalf("expected error for full event")
	}

	// event in the past
	past := now.Add(-time.Hour)
	eventPast := event
	eventPast.Date = &past
	uc.storage = &stubTxManager{do: func(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error {
		return fn(ctx, &stubTx{getEventFunc: func(ctx context.Context, eventID string) (domain.Event, error) { return eventPast, nil }})
	}}
	resp = uc.Book(ctx, "e1", "u1")
	if resp.Error == "" {
		t.Fatalf("expected error for past event")
	}
}

func TestConfirm_Flows(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	future := now.Add(time.Hour)
	booking := domain.Booking{ID: "b1", EventID: "e1", Status: domain.BookingStatusPending, PaymentDeadline: &future}

	cancelCalled := false
	uc := &Usecase{
		nf: &stubNotif{
			send: func(ctx context.Context, notif domain.DelayedNotification) (string, error) { return "", nil },
			canc: func(ctx context.Context, id string) error { cancelCalled = true; return nil },
		},
		storage: &stubTxManager{do: func(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error {
			return fn(ctx, &stubTx{
				getEventFunc:          func(ctx context.Context, eventID string) (domain.Event, error) { return domain.Event{ID: eventID}, nil },
				getBookingFunc:        func(ctx context.Context, bookingID string) (domain.Booking, error) { return booking, nil },
				confirmFunc:           func(ctx context.Context, bookingID string) (string, error) { return domain.BookingStatusConfirmed, nil },
				getNotificationIDFunc: func(ctx context.Context, bookingID string) (string, error) { return "notif-1", nil },
			})
		}},
	}
	resp := uc.Confirm(ctx, "e1", "b1")
	if resp.Error != "" || resp.Status != domain.BookingStatusConfirmed || !cancelCalled {
		t.Fatalf("unexpected response: %+v cancel=%v", resp, cancelCalled)
	}

	// invalid association
	invalid := booking
	invalid.EventID = "other"
	uc.storage = &stubTxManager{do: func(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error {
		return fn(ctx, &stubTx{
			getEventFunc:   func(ctx context.Context, eventID string) (domain.Event, error) { return domain.Event{ID: eventID}, nil },
			getBookingFunc: func(ctx context.Context, bookingID string) (domain.Booking, error) { return invalid, nil },
		})
	}}
	resp = uc.Confirm(ctx, "e1", "b1")
	if resp.Error == "" {
		t.Fatalf("expected error for mismatched event")
	}

	// expired payment
	exp := booking
	exp.PaymentDeadline = timePtr(now.Add(-time.Minute))
	uc.storage = &stubTxManager{do: func(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error {
		return fn(ctx, &stubTx{getEventFunc: func(ctx context.Context, eventID string) (domain.Event, error) { return domain.Event{ID: eventID}, nil }, getBookingFunc: func(ctx context.Context, bookingID string) (domain.Booking, error) { return exp, nil }})
	}}
	resp = uc.Confirm(ctx, "e1", "b1")
	if resp.Error == "" {
		t.Fatalf("expected error for expired booking")
	}
}

func timePtr(t time.Time) *time.Time { return &t }