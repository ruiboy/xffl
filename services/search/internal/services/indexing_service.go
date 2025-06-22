package services

import (
	"context"
	"fmt"
	"log"
	"time"
	"xffl/pkg/events"
	"xffl/services/search/internal/domain"
	searchEvents "xffl/services/search/internal/domain/events"
)

// eventDispatcher defines the interface for publishing events
type eventDispatcher interface {
	Publish(ctx context.Context, event events.DomainEvent) error
}

// IndexingService handles indexing operations triggered by domain events
type IndexingService struct {
	searchService   *SearchService
	eventDispatcher eventDispatcher
	logger          *log.Logger
}

// NewIndexingService creates a new IndexingService
func NewIndexingService(searchService *SearchService, eventDispatcher eventDispatcher, logger *log.Logger) *IndexingService {
	if logger == nil {
		logger = log.Default()
	}
	
	return &IndexingService{
		searchService:   searchService,
		eventDispatcher: eventDispatcher,
		logger:          logger,
	}
}

// HandlePlayerMatchUpdated processes AFL player match updates for indexing
func (s *IndexingService) HandlePlayerMatchUpdated(ctx context.Context, event events.DomainEvent) error {
	s.logger.Printf("Processing PlayerMatchUpdated event for indexing: %s", event.AggregateID())
	
	// Get event data
	eventData := event.EventData()
	
	// Create player match document for indexing
	doc := s.createPlayerMatchDocument(eventData, "afl")
	
	// Index the document
	if err := s.searchService.IndexDocument(ctx, doc); err != nil {
		s.logger.Printf("Failed to index player match document: %v", err)
		return fmt.Errorf("failed to index player match: %w", err)
	}
	
	// Publish index update event
	indexEvent := searchEvents.NewIndexUpdated(doc.ID, string(doc.Type), doc.Source, "update")
	if err := s.eventDispatcher.Publish(ctx, indexEvent); err != nil {
		s.logger.Printf("Failed to publish index update event: %v", err)
		// Don't fail the operation for event publishing failures
	}
	
	s.logger.Printf("Successfully indexed player match document: %s", doc.ID)
	return nil
}

// HandleFantasyScoreCalculated processes FFL fantasy score updates for indexing
func (s *IndexingService) HandleFantasyScoreCalculated(ctx context.Context, event events.DomainEvent) error {
	s.logger.Printf("Processing FantasyScoreCalculated event for indexing: %s", event.AggregateID())
	
	// Get event data
	eventData := event.EventData()
	
	// Create player document for indexing (fantasy scores update player searchability)
	doc := s.createPlayerDocument(eventData, "ffl")
	
	// Index the document
	if err := s.searchService.IndexDocument(ctx, doc); err != nil {
		s.logger.Printf("Failed to index player document: %v", err)
		return fmt.Errorf("failed to index player: %w", err)
	}
	
	// Publish index update event
	indexEvent := searchEvents.NewIndexUpdated(doc.ID, string(doc.Type), doc.Source, "update")
	if err := s.eventDispatcher.Publish(ctx, indexEvent); err != nil {
		s.logger.Printf("Failed to publish index update event: %v", err)
		// Don't fail the operation for event publishing failures
	}
	
	s.logger.Printf("Successfully indexed player document: %s", doc.ID)
	return nil
}

// createPlayerMatchDocument creates a search document from player match event data
func (s *IndexingService) createPlayerMatchDocument(eventData map[string]interface{}, source string) domain.SearchDocument {
	// Extract data with safe type assertions
	playerMatchID := s.getUintFromInterface(eventData["player_match_id"])
	playerSeasonID := s.getUintFromInterface(eventData["player_season_id"])
	
	// Create a generic document for player match statistics
	return domain.SearchDocument{
		ID:      fmt.Sprintf("%s_player_match_%d", source, playerMatchID),
		Type:    domain.DocumentTypePlayerMatch,
		Source:  source,
		Title:   fmt.Sprintf("Player Match Stats #%d", playerMatchID),
		Content: s.generatePlayerMatchContent(eventData, source),
		Tags:    []string{source, "player_match", "statistics"},
		Metadata: map[string]interface{}{
			"player_match_id":  playerMatchID,
			"player_season_id": playerSeasonID,
			"kicks":            s.getIntFromInterface(eventData["kicks"]),
			"handballs":        s.getIntFromInterface(eventData["handballs"]),
			"goals":            s.getIntFromInterface(eventData["goals"]),
		},
		IndexedAt:    time.Now(),
		LastModified: time.Now(),
	}
}

// createPlayerDocument creates a search document from player event data
func (s *IndexingService) createPlayerDocument(eventData map[string]interface{}, source string) domain.SearchDocument {
	// Extract data with safe type assertions
	playerID := s.getUintFromInterface(eventData["player_id"])
	
	return domain.SearchDocument{
		ID:      fmt.Sprintf("%s_player_%d", source, playerID),
		Type:    domain.DocumentTypePlayer,
		Source:  source,
		Title:   s.getStringFromInterface(eventData["player_name"]),
		Content: s.generatePlayerContent(eventData, source),
		Tags:    []string{source, "player"},
		Metadata: map[string]interface{}{
			"player_id": playerID,
			"club_id":   s.getUintFromInterface(eventData["club_id"]),
			"score":     s.getFloatFromInterface(eventData["score"]),
		},
		IndexedAt:    time.Now(),
		LastModified: time.Now(),
	}
}

// Helper methods for safe type conversion
func (s *IndexingService) getUintFromInterface(v interface{}) uint {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case uint:
		return val
	case int:
		return uint(val)
	case float64:
		return uint(val)
	default:
		return 0
	}
}

func (s *IndexingService) getIntFromInterface(v interface{}) int {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int:
		return val
	case uint:
		return int(val)
	case float64:
		return int(val)
	default:
		return 0
	}
}

func (s *IndexingService) getFloatFromInterface(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case uint:
		return float64(val)
	default:
		return 0
	}
}

func (s *IndexingService) getStringFromInterface(v interface{}) string {
	if v == nil {
		return ""
	}
	if str, ok := v.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", v)
}

func (s *IndexingService) generatePlayerMatchContent(eventData map[string]interface{}, source string) string {
	kicks := s.getIntFromInterface(eventData["kicks"])
	handballs := s.getIntFromInterface(eventData["handballs"])
	goals := s.getIntFromInterface(eventData["goals"])
	
	return fmt.Sprintf("Player match statistics: %d kicks, %d handballs, %d goals in %s", 
		kicks, handballs, goals, source)
}

func (s *IndexingService) generatePlayerContent(eventData map[string]interface{}, source string) string {
	playerName := s.getStringFromInterface(eventData["player_name"])
	score := s.getFloatFromInterface(eventData["score"])
	
	return fmt.Sprintf("Player %s with score %.2f in %s", playerName, score, source)
}