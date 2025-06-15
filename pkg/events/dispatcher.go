package events

import "context"

// EventDispatcher defines the interface for publishing and subscribing to domain events
type EventDispatcher interface {
	// Publish publishes a domain event to all subscribed handlers
	Publish(ctx context.Context, event DomainEvent) error
	
	// Subscribe subscribes a handler to events of the specified type
	Subscribe(eventType string, handler EventHandler) error
	
	// Start initializes the dispatcher (for implementations that need setup)
	Start(ctx context.Context) error
	
	// Stop gracefully shuts down the dispatcher
	Stop() error
}

// PublishResult represents the result of publishing an event
type PublishResult struct {
	EventType     string
	HandlersCount int
	Errors        []HandlerError
}

// HandlerError represents an error that occurred in a specific handler
type HandlerError struct {
	HandlerName string
	Error       error
}

// HasErrors returns true if any handlers failed
func (r *PublishResult) HasErrors() bool {
	return len(r.Errors) > 0
}