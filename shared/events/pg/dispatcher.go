// Package pg provides a PG LISTEN/NOTIFY EventDispatcher implementation.
package pg

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"xffl/shared/events"
)

// message is the JSON envelope sent over NOTIFY.
type message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Dispatcher publishes and subscribes to events via PG LISTEN/NOTIFY.
// All events use a single PG channel to keep things simple.
type Dispatcher struct {
	pool    *pgxpool.Pool
	channel string

	mu       sync.RWMutex
	handlers map[string][]events.Handler
}

// New creates a PG dispatcher that uses the given pool and channel name.
func New(pool *pgxpool.Pool, channel string) *Dispatcher {
	return &Dispatcher{
		pool:     pool,
		channel:  channel,
		handlers: make(map[string][]events.Handler),
	}
}

// Publish sends an event via PG NOTIFY.
func (d *Dispatcher) Publish(ctx context.Context, eventType string, payload []byte) error {
	msg, err := json.Marshal(message{
		Type:    eventType,
		Payload: payload,
	})
	if err != nil {
		return fmt.Errorf("pg dispatch marshal: %w", err)
	}

	_, err = d.pool.Exec(ctx, "SELECT pg_notify($1, $2)", d.channel, string(msg))
	if err != nil {
		return fmt.Errorf("pg dispatch notify: %w", err)
	}
	return nil
}

// Subscribe registers a handler for a given event type.
func (d *Dispatcher) Subscribe(eventType string, handler events.Handler) {
	d.mu.Lock()
	d.handlers[eventType] = append(d.handlers[eventType], handler)
	d.mu.Unlock()
}

// Listen starts listening for notifications on the channel and dispatches
// them to registered handlers. It blocks until the context is cancelled.
// Call this in a goroutine.
func (d *Dispatcher) Listen(ctx context.Context) error {
	conn, err := d.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("pg listen acquire: %w", err)
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, "LISTEN "+d.channel)
	if err != nil {
		return fmt.Errorf("pg listen: %w", err)
	}

	for {
		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil // clean shutdown
			}
			return fmt.Errorf("pg wait notification: %w", err)
		}

		var msg message
		if err := json.Unmarshal([]byte(notification.Payload), &msg); err != nil {
			log.Printf("pg dispatch: invalid message: %v", err)
			continue
		}

		d.mu.RLock()
		handlers := d.handlers[msg.Type]
		d.mu.RUnlock()

		for _, h := range handlers {
			if err := h(ctx, msg.Payload); err != nil {
				log.Printf("pg dispatch: handler error for %s: %v", msg.Type, err)
			}
		}
	}
}
