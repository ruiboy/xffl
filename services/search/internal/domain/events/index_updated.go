package events

import (
	"xffl/pkg/events"
)

// IndexUpdated represents when a document is updated in the search index
type IndexUpdated struct {
	events.BaseEvent
	DocumentID   string `json:"document_id"`
	DocumentType string `json:"document_type"`
	Source       string `json:"source"`
	Operation    string `json:"operation"` // "create", "update", "delete"
}

// EventData returns the event data for serialization
func (e IndexUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"document_id":   e.DocumentID,
		"document_type": e.DocumentType,
		"source":        e.Source,
		"operation":     e.Operation,
	}
}

// Ensure IndexUpdated implements DomainEvent
var _ events.DomainEvent = (*IndexUpdated)(nil)

// NewIndexUpdated creates a new IndexUpdated event
func NewIndexUpdated(documentID, documentType, source, operation string) IndexUpdated {
	baseEvent := events.NewBaseEvent("search.index_updated", "v1", documentID)
	
	return IndexUpdated{
		BaseEvent:    baseEvent,
		DocumentID:   documentID,
		DocumentType: documentType,
		Source:       source,
		Operation:    operation,
	}
}