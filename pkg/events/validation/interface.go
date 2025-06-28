package validation

import "xffl/pkg/events"

// EventValidator defines the interface for event validation
type EventValidator interface {
	// ValidateEvent validates a domain event
	ValidateEvent(event events.DomainEvent) error
	
	// GetSupportedEventTypes returns list of event types that have validation
	GetSupportedEventTypes() []string
}