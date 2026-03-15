// Package events defines the EventDispatcher interface for cross-service communication.
package events

import "context"

// Handler is a function that processes an event payload.
type Handler func(ctx context.Context, payload []byte) error

// Dispatcher publishes events and manages subscriptions.
type Dispatcher interface {
	// Publish sends an event with the given type and JSON payload.
	Publish(ctx context.Context, eventType string, payload []byte) error

	// Subscribe registers a handler for a given event type.
	Subscribe(eventType string, handler Handler)
}
