package events

import (
	"context"
	"log"
	"xffl/pkg/events"
	"xffl/services/search/internal/services"
)

// PlayerMatchHandler handles AFL player match update events for search indexing
type PlayerMatchHandler struct {
	indexingService *services.IndexingService
	logger          *log.Logger
}

// NewPlayerMatchHandler creates a new player match event handler
func NewPlayerMatchHandler(indexingService *services.IndexingService, logger *log.Logger) *PlayerMatchHandler {
	if logger == nil {
		logger = log.Default()
	}
	
	return &PlayerMatchHandler{
		indexingService: indexingService,
		logger:          logger,
	}
}

// Handle processes a player match updated event
func (h *PlayerMatchHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	h.logger.Printf("PlayerMatchHandler: Processing event %s for aggregate %s", 
		event.EventType(), event.AggregateID())
	
	return h.indexingService.HandlePlayerMatchUpdated(ctx, event)
}

// HandlerName returns the name of this handler
func (h *PlayerMatchHandler) HandlerName() string {
	return "search.player_match_indexer"
}

// FantasyScoreHandler handles FFL fantasy score calculation events for search indexing
type FantasyScoreHandler struct {
	indexingService *services.IndexingService
	logger          *log.Logger
}

// NewFantasyScoreHandler creates a new fantasy score event handler
func NewFantasyScoreHandler(indexingService *services.IndexingService, logger *log.Logger) *FantasyScoreHandler {
	if logger == nil {
		logger = log.Default()
	}
	
	return &FantasyScoreHandler{
		indexingService: indexingService,
		logger:          logger,
	}
}

// Handle processes a fantasy score calculated event
func (h *FantasyScoreHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	h.logger.Printf("FantasyScoreHandler: Processing event %s for aggregate %s", 
		event.EventType(), event.AggregateID())
	
	return h.indexingService.HandleFantasyScoreCalculated(ctx, event)
}

// HandlerName returns the name of this handler
func (h *FantasyScoreHandler) HandlerName() string {
	return "search.fantasy_score_indexer"
}

// IndexUpdateHandler handles search index update events (for monitoring/logging)
type IndexUpdateHandler struct {
	logger *log.Logger
}

// NewIndexUpdateHandler creates a new index update event handler
func NewIndexUpdateHandler(logger *log.Logger) *IndexUpdateHandler {
	if logger == nil {
		logger = log.Default()
	}
	
	return &IndexUpdateHandler{
		logger: logger,
	}
}

// Handle processes an index update event
func (h *IndexUpdateHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	h.logger.Printf("IndexUpdateHandler: Search index updated - %s for aggregate %s", 
		event.EventType(), event.AggregateID())
	
	// This handler is primarily for logging/monitoring
	// In a production system, you might want to:
	// - Update metrics
	// - Send notifications
	// - Update search analytics
	
	return nil
}

// HandlerName returns the name of this handler
func (h *IndexUpdateHandler) HandlerName() string {
	return "search.index_update_monitor"
}