package application

import (
	"context"
	"fmt"
	"log"
	"xffl/pkg/events"
	domainEvents "xffl/services/ffl/internal/domain/events"
)

// FantasyScoreService handles fantasy score calculations based on AFL events
type FantasyScoreService struct {
	eventDispatcher events.EventDispatcher
	logger          *log.Logger
}

// NewFantasyScoreService creates a new fantasy score service
func NewFantasyScoreService(eventDispatcher events.EventDispatcher, logger *log.Logger) *FantasyScoreService {
	return &FantasyScoreService{
		eventDispatcher: eventDispatcher,
		logger:          logger,
	}
}

// HandlerName implements EventHandler interface
func (s *FantasyScoreService) HandlerName() string {
	return "FFL.FantasyScoreCalculator"
}

// Handle processes AFL PlayerMatchUpdated events and calculates fantasy scores
func (s *FantasyScoreService) Handle(ctx context.Context, event events.DomainEvent) error {
	if event.EventType() != "AFL.PlayerMatchUpdated" {
		return nil // Not interested in other event types
	}

	// Extract event data
	eventData := event.EventData()
	
	// Parse player and match IDs
	playerSeasonID, ok := eventData["playerSeasonId"].(float64)
	if !ok {
		return fmt.Errorf("invalid playerSeasonId in event data")
	}
	
	clubMatchID, ok := eventData["clubMatchId"].(float64)
	if !ok {
		return fmt.Errorf("invalid clubMatchId in event data")
	}
	
	// Parse new stats
	newStatsData, ok := eventData["newStats"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid newStats in event data")
	}
	
	// Extract individual stats with safe type conversion
	kicks := s.safeFloatToInt(newStatsData["kicks"])
	handballs := s.safeFloatToInt(newStatsData["handballs"])
	marks := s.safeFloatToInt(newStatsData["marks"])
	hitouts := s.safeFloatToInt(newStatsData["hitouts"])
	tackles := s.safeFloatToInt(newStatsData["tackles"])
	goals := s.safeFloatToInt(newStatsData["goals"])
	behinds := s.safeFloatToInt(newStatsData["behinds"])
	
	// Calculate AFL score (simple sum for now)
	aflScore := kicks + handballs + marks + hitouts + tackles + goals + behinds
	
	// Calculate fantasy score using typical fantasy football scoring
	fantasyScore := (kicks * 3) + (handballs * 2) + (marks * 3) + 
		(hitouts * 1) + (tackles * 4) + (goals * 6) + (behinds * 1)
	
	s.logger.Printf("Calculated fantasy score for player %d: AFL=%d, Fantasy=%d", 
		int(playerSeasonID), aflScore, fantasyScore)
	
	// Create and publish FantasyScoreCalculated event
	fantasyEvent := domainEvents.NewFantasyScoreCalculatedEvent(
		uint(playerSeasonID), 
		uint(clubMatchID), 
		aflScore, 
		fantasyScore, 
		"AFL.PlayerMatchUpdated",
	)
	
	if err := s.eventDispatcher.Publish(ctx, fantasyEvent); err != nil {
		s.logger.Printf("Failed to publish FantasyScoreCalculated event: %v", err)
		return err
	}
	
	return nil
}

// safeFloatToInt safely converts interface{} to int, defaulting to 0
func (s *FantasyScoreService) safeFloatToInt(value interface{}) int {
	if f, ok := value.(float64); ok {
		return int(f)
	}
	return 0
}