// Package memory provides an in-memory EventDispatcher for testing.
package memory

import (
	"context"
	"sync"

	"xffl/shared/events"
)

// Dispatcher is a synchronous in-memory event dispatcher.
// Handlers are called in the order they were registered.
type Dispatcher struct {
	mu       sync.RWMutex
	handlers map[string][]events.Handler
}

// New creates a new in-memory Dispatcher.
func New() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[string][]events.Handler),
	}
}

// Publish calls all handlers registered for the given event type.
// Returns the first handler error encountered.
func (d *Dispatcher) Publish(ctx context.Context, eventType string, payload []byte) error {
	d.mu.RLock()
	handlers := d.handlers[eventType]
	d.mu.RUnlock()

	for _, h := range handlers {
		if err := h(ctx, payload); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe registers a handler for the given event type.
func (d *Dispatcher) Subscribe(eventType string, handler events.Handler) {
	d.mu.Lock()
	d.handlers[eventType] = append(d.handlers[eventType], handler)
	d.mu.Unlock()
}
