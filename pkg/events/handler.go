package events

import "context"

// EventHandler defines the interface for handling domain events
type EventHandler interface {
	// Handle processes the given domain event
	Handle(ctx context.Context, event DomainEvent) error
	
	// HandlerName returns a unique name for this handler
	HandlerName() string
}

// EventHandlerFunc is a function adapter for EventHandler
type EventHandlerFunc struct {
	name string
	fn   func(ctx context.Context, event DomainEvent) error
}

// NewEventHandlerFunc creates a new EventHandlerFunc
func NewEventHandlerFunc(name string, fn func(ctx context.Context, event DomainEvent) error) *EventHandlerFunc {
	return &EventHandlerFunc{
		name: name,
		fn:   fn,
	}
}

// Handle implements EventHandler
func (h *EventHandlerFunc) Handle(ctx context.Context, event DomainEvent) error {
	return h.fn(ctx, event)
}

// HandlerName implements EventHandler
func (h *EventHandlerFunc) HandlerName() string {
	return h.name
}