package events

import "time"

// DomainEvent represents a domain event that occurred in the system
type DomainEvent interface {
	// EventType returns the type of the event (e.g., "PlayerMatchUpdated")
	EventType() string
	
	// EventVersion returns the version of the event schema
	EventVersion() string
	
	// AggregateID returns the ID of the aggregate that generated the event
	AggregateID() string
	
	// OccurredAt returns when the event occurred
	OccurredAt() time.Time
	
	// EventData returns the event data as key-value pairs for serialization
	EventData() map[string]interface{}
}

// BaseEvent provides common fields for domain events
type BaseEvent struct {
	eventType   string
	version     string
	aggregateID string
	occurredAt  time.Time
}

// NewBaseEvent creates a new base event with common fields
func NewBaseEvent(eventType, version, aggregateID string) BaseEvent {
	return BaseEvent{
		eventType:   eventType,
		version:     version,
		aggregateID: aggregateID,
		occurredAt:  time.Now(),
	}
}

// EventType returns the type of the event
func (e BaseEvent) EventType() string {
	return e.eventType
}

// EventVersion returns the version of the event schema
func (e BaseEvent) EventVersion() string {
	return e.version
}

// AggregateID returns the ID of the aggregate that generated the event
func (e BaseEvent) AggregateID() string {
	return e.aggregateID
}

// OccurredAt returns when the event occurred
func (e BaseEvent) OccurredAt() time.Time {
	return e.occurredAt
}