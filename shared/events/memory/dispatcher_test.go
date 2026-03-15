package memory

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func TestDispatcher_PublishCallsSubscriber(t *testing.T) {
	d := New()

	var got []byte
	d.Subscribe("test.event", func(ctx context.Context, payload []byte) error {
		got = payload
		return nil
	})

	err := d.Publish(context.Background(), "test.event", []byte(`{"id":1}`))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if string(got) != `{"id":1}` {
		t.Errorf("handler got %q, want %q", got, `{"id":1}`)
	}
}

func TestDispatcher_MultipleSubscribers(t *testing.T) {
	d := New()

	var count int
	var mu sync.Mutex
	handler := func(ctx context.Context, payload []byte) error {
		mu.Lock()
		count++
		mu.Unlock()
		return nil
	}

	d.Subscribe("test.event", handler)
	d.Subscribe("test.event", handler)

	err := d.Publish(context.Background(), "test.event", []byte(`{}`))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 handler calls, got %d", count)
	}
}

func TestDispatcher_NoSubscribers(t *testing.T) {
	d := New()

	err := d.Publish(context.Background(), "unsubscribed.event", []byte(`{}`))
	if err != nil {
		t.Fatalf("Publish() with no subscribers should not error, got %v", err)
	}
}

func TestDispatcher_HandlerError(t *testing.T) {
	d := New()

	wantErr := errors.New("handler failed")
	d.Subscribe("test.event", func(ctx context.Context, payload []byte) error {
		return wantErr
	})

	err := d.Publish(context.Background(), "test.event", []byte(`{}`))
	if !errors.Is(err, wantErr) {
		t.Fatalf("Publish() error = %v, want %v", err, wantErr)
	}
}

func TestDispatcher_DifferentEventTypes(t *testing.T) {
	d := New()

	var aCalled, bCalled bool
	d.Subscribe("event.a", func(ctx context.Context, payload []byte) error {
		aCalled = true
		return nil
	})
	d.Subscribe("event.b", func(ctx context.Context, payload []byte) error {
		bCalled = true
		return nil
	})

	if err := d.Publish(context.Background(), "event.a", []byte(`{}`)); err != nil {
		t.Fatal(err)
	}

	if !aCalled {
		t.Error("event.a handler should have been called")
	}
	if bCalled {
		t.Error("event.b handler should not have been called")
	}
}
