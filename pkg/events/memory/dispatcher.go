package memory

import (
	"context"
	"fmt"
	"log"
	"sync"
	"xffl/pkg/events"
)

// InMemoryDispatcher is an in-memory implementation of EventDispatcher
type InMemoryDispatcher struct {
	mu       sync.RWMutex
	handlers map[string][]events.EventHandler
	running  bool
	logger   *log.Logger
}

// NewInMemoryDispatcher creates a new in-memory event dispatcher
func NewInMemoryDispatcher(logger *log.Logger) *InMemoryDispatcher {
	if logger == nil {
		logger = log.Default()
	}
	
	return &InMemoryDispatcher{
		handlers: make(map[string][]events.EventHandler),
		logger:   logger,
	}
}

// Start initializes the dispatcher
func (d *InMemoryDispatcher) Start(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	if d.running {
		return fmt.Errorf("dispatcher is already running")
	}
	
	d.running = true
	d.logger.Println("InMemoryDispatcher started")
	return nil
}

// Stop gracefully shuts down the dispatcher
func (d *InMemoryDispatcher) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	if !d.running {
		return fmt.Errorf("dispatcher is not running")
	}
	
	d.running = false
	d.logger.Println("InMemoryDispatcher stopped")
	return nil
}

// Subscribe adds a handler for the specified event type
func (d *InMemoryDispatcher) Subscribe(eventType string, handler events.EventHandler) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	if eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}
	
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}
	
	// Check for duplicate handler names
	for _, h := range d.handlers[eventType] {
		if h.HandlerName() == handler.HandlerName() {
			return fmt.Errorf("handler with name '%s' already subscribed to event type '%s'", handler.HandlerName(), eventType)
		}
	}
	
	d.handlers[eventType] = append(d.handlers[eventType], handler)
	d.logger.Printf("Subscribed handler '%s' to event type '%s'", handler.HandlerName(), eventType)
	
	return nil
}

// Publish publishes an event to all subscribed handlers
func (d *InMemoryDispatcher) Publish(ctx context.Context, event events.DomainEvent) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	if !d.running {
		return fmt.Errorf("dispatcher is not running")
	}
	
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}
	
	eventType := event.EventType()
	handlers, exists := d.handlers[eventType]
	
	if !exists || len(handlers) == 0 {
		d.logger.Printf("No handlers found for event type '%s'", eventType)
		return nil
	}
	
	var handlerErrors []events.HandlerError
	
	// Execute handlers synchronously in memory implementation
	for _, handler := range handlers {
		if err := d.executeHandler(ctx, handler, event); err != nil {
			handlerError := events.HandlerError{
				HandlerName: handler.HandlerName(),
				Error:       err,
			}
			handlerErrors = append(handlerErrors, handlerError)
			d.logger.Printf("Handler '%s' failed for event type '%s': %v", handler.HandlerName(), eventType, err)
		}
	}
	
	d.logger.Printf("Published event '%s' to %d handlers (%d errors)", eventType, len(handlers), len(handlerErrors))
	
	// For in-memory implementation, we log errors but don't fail the publish
	// In production implementations, you might want different error handling strategies
	return nil
}

// executeHandler executes a single handler with error recovery
func (d *InMemoryDispatcher) executeHandler(ctx context.Context, handler events.EventHandler, event events.DomainEvent) (err error) {
	// Recover from panics in handlers
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("handler panicked: %v", r)
		}
	}()
	
	return handler.Handle(ctx, event)
}

// GetSubscribedHandlers returns the handlers for a given event type (useful for testing)
func (d *InMemoryDispatcher) GetSubscribedHandlers(eventType string) []events.EventHandler {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	handlers := d.handlers[eventType]
	result := make([]events.EventHandler, len(handlers))
	copy(result, handlers)
	return result
}