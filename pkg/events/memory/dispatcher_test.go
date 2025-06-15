package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"xffl/pkg/events"
)

// TestEvent is a simple event for testing
type TestEvent struct {
	events.BaseEvent
	PlayerID uint   `json:"playerId"`
	Message  string `json:"message"`
}

func NewTestEvent(playerID uint, message string) *TestEvent {
	return &TestEvent{
		BaseEvent: events.NewBaseEvent("TestEvent", "v1", "test-aggregate"),
		PlayerID:  playerID,
		Message:   message,
	}
}

func (e *TestEvent) EventData() map[string]interface{} {
	data, _ := json.Marshal(e)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}

func TestInMemoryDispatcher_Basic(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	dispatcher := NewInMemoryDispatcher(logger)
	
	ctx := context.Background()
	
	// Start dispatcher
	if err := dispatcher.Start(ctx); err != nil {
		t.Fatalf("Failed to start dispatcher: %v", err)
	}
	defer dispatcher.Stop()
	
	// Create test handlers
	var handler1Called, handler2Called bool
	var handler1Event events.DomainEvent
	
	handler1 := events.NewEventHandlerFunc("handler1", func(ctx context.Context, event events.DomainEvent) error {
		handler1Called = true
		handler1Event = event
		return nil
	})
	
	handler2 := events.NewEventHandlerFunc("handler2", func(ctx context.Context, event events.DomainEvent) error {
		handler2Called = true
		return nil
	})
	
	// Subscribe handlers
	if err := dispatcher.Subscribe("TestEvent", handler1); err != nil {
		t.Fatalf("Failed to subscribe handler1: %v", err)
	}
	
	if err := dispatcher.Subscribe("TestEvent", handler2); err != nil {
		t.Fatalf("Failed to subscribe handler2: %v", err)
	}
	
	// Publish event
	event := NewTestEvent(123, "test message")
	if err := dispatcher.Publish(ctx, event); err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}
	
	// Verify both handlers were called
	if !handler1Called {
		t.Error("Handler1 was not called")
	}
	if !handler2Called {
		t.Error("Handler2 was not called")
	}
	
	// Verify event data
	if handler1Event.EventType() != "TestEvent" {
		t.Errorf("Expected event type 'TestEvent', got '%s'", handler1Event.EventType())
	}
	
	if handler1Event.AggregateID() != "test-aggregate" {
		t.Errorf("Expected aggregate ID 'test-aggregate', got '%s'", handler1Event.AggregateID())
	}
}

func TestInMemoryDispatcher_NoHandlers(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	dispatcher := NewInMemoryDispatcher(logger)
	
	ctx := context.Background()
	
	// Start dispatcher
	if err := dispatcher.Start(ctx); err != nil {
		t.Fatalf("Failed to start dispatcher: %v", err)
	}
	defer dispatcher.Stop()
	
	// Publish event with no handlers
	event := NewTestEvent(123, "test message")
	if err := dispatcher.Publish(ctx, event); err != nil {
		t.Fatalf("Publishing to no handlers should not error: %v", err)
	}
}

func TestInMemoryDispatcher_HandlerError(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	dispatcher := NewInMemoryDispatcher(logger)
	
	ctx := context.Background()
	
	// Start dispatcher
	if err := dispatcher.Start(ctx); err != nil {
		t.Fatalf("Failed to start dispatcher: %v", err)
	}
	defer dispatcher.Stop()
	
	// Create handlers - one that succeeds, one that fails
	var successCalled bool
	
	successHandler := events.NewEventHandlerFunc("success", func(ctx context.Context, event events.DomainEvent) error {
		successCalled = true
		return nil
	})
	
	errorHandler := events.NewEventHandlerFunc("error", func(ctx context.Context, event events.DomainEvent) error {
		return fmt.Errorf("intentional error for testing")
	})
	
	// Subscribe handlers
	dispatcher.Subscribe("TestEvent", successHandler)
	dispatcher.Subscribe("TestEvent", errorHandler)
	
	// Publish event - should not fail even if handler errors
	event := NewTestEvent(123, "test message")
	if err := dispatcher.Publish(ctx, event); err != nil {
		t.Fatalf("Publish should not fail due to handler errors: %v", err)
	}
	
	// Success handler should still have been called
	if !successCalled {
		t.Error("Success handler should have been called despite other handler failing")
	}
}

func TestInMemoryDispatcher_DuplicateHandler(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	dispatcher := NewInMemoryDispatcher(logger)
	
	handler1 := events.NewEventHandlerFunc("duplicate", func(ctx context.Context, event events.DomainEvent) error {
		return nil
	})
	
	handler2 := events.NewEventHandlerFunc("duplicate", func(ctx context.Context, event events.DomainEvent) error {
		return nil
	})
	
	// First subscription should succeed
	if err := dispatcher.Subscribe("TestEvent", handler1); err != nil {
		t.Fatalf("First subscription should succeed: %v", err)
	}
	
	// Second subscription with same name should fail
	if err := dispatcher.Subscribe("TestEvent", handler2); err == nil {
		t.Error("Duplicate handler subscription should fail")
	}
}