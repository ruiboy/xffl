package pg

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://postgres:postgres@localhost:5432/xffl?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Skipf("skipping: cannot connect to postgres: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		t.Skipf("skipping: cannot ping postgres: %v", err)
	}

	t.Cleanup(func() { pool.Close() })
	return pool
}

func TestDispatcher_PublishAndListen(t *testing.T) {
	pool := testPool(t)
	d := New(pool, "test_events")

	received := make(chan []byte, 1)
	d.Subscribe("test.event", func(ctx context.Context, payload []byte) error {
		received <- payload
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	listenErr := make(chan error, 1)
	go func() {
		listenErr <- d.Listen(ctx)
	}()

	// Give the listener time to start
	time.Sleep(100 * time.Millisecond)

	err := d.Publish(context.Background(), "test.event", []byte(`{"id":42}`))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	select {
	case got := <-received:
		if string(got) != `{"id":42}` {
			t.Errorf("got payload %q, want %q", got, `{"id":42}`)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for event")
	}

	cancel()
	if err := <-listenErr; err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
}

func TestDispatcher_OnlyMatchingSubscribers(t *testing.T) {
	pool := testPool(t)
	d := New(pool, "test_events_filter")

	matched := make(chan bool, 1)
	d.Subscribe("wanted.event", func(ctx context.Context, payload []byte) error {
		matched <- true
		return nil
	})
	d.Subscribe("other.event", func(ctx context.Context, payload []byte) error {
		t.Error("other.event handler should not be called")
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go d.Listen(ctx)
	time.Sleep(100 * time.Millisecond)

	err := d.Publish(context.Background(), "wanted.event", []byte(`{}`))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	select {
	case <-matched:
		// success
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for event")
	}
}
