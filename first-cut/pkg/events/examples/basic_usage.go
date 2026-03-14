package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"xffl/pkg/events"
	"xffl/pkg/events/memory"
)

// PlayerMatchUpdatedEvent represents a domain event when a player match is updated
type PlayerMatchUpdatedEvent struct {
	events.BaseEvent
	PlayerSeasonID uint `json:"playerSeasonId"`
	ClubMatchID    uint `json:"clubMatchId"`
	Kicks          int  `json:"kicks"`
	Handballs      int  `json:"handballs"`
	Goals          int  `json:"goals"`
}

// NewPlayerMatchUpdatedEvent creates a new PlayerMatchUpdatedEvent
func NewPlayerMatchUpdatedEvent(playerSeasonID, clubMatchID uint, kicks, handballs, goals int) *PlayerMatchUpdatedEvent {
	return &PlayerMatchUpdatedEvent{
		BaseEvent:      events.NewBaseEvent("PlayerMatchUpdated", "v1", fmt.Sprintf("player-season-%d", playerSeasonID)),
		PlayerSeasonID: playerSeasonID,
		ClubMatchID:    clubMatchID,
		Kicks:          kicks,
		Handballs:      handballs,
		Goals:          goals,
	}
}

// EventData implements DomainEvent interface
func (e *PlayerMatchUpdatedEvent) EventData() map[string]interface{} {
	data, _ := json.Marshal(e)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}

// Example handlers
type FantasyScoreCalculatorHandler struct {
	name string
}

func NewFantasyScoreCalculatorHandler() *FantasyScoreCalculatorHandler {
	return &FantasyScoreCalculatorHandler{name: "FantasyScoreCalculator"}
}

func (h *FantasyScoreCalculatorHandler) HandlerName() string {
	return h.name
}

func (h *FantasyScoreCalculatorHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	if event.EventType() != "PlayerMatchUpdated" {
		return nil // Not interested in this event type
	}
	
	// Type assert to get the specific event data
	if playerEvent, ok := event.(*PlayerMatchUpdatedEvent); ok {
		fantasyScore := (playerEvent.Kicks * 3) + (playerEvent.Handballs * 2) + (playerEvent.Goals * 6)
		fmt.Printf("🏆 Fantasy Score Calculator: Player %d scored %d fantasy points\n", 
			playerEvent.PlayerSeasonID, fantasyScore)
	}
	
	return nil
}

type SearchIndexHandler struct {
	name string
}

func NewSearchIndexHandler() *SearchIndexHandler {
	return &SearchIndexHandler{name: "SearchIndexUpdater"}
}

func (h *SearchIndexHandler) HandlerName() string {
	return h.name
}

func (h *SearchIndexHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	if event.EventType() != "PlayerMatchUpdated" {
		return nil
	}
	
	fmt.Printf("🔍 Search Index: Updated search index for aggregate %s\n", event.AggregateID())
	return nil
}

func main() {
	fmt.Println("🚀 Event System Basic Usage Example")
	fmt.Println("=====================================")
	
	// Create logger
	logger := log.New(os.Stdout, "[EVENTS] ", log.LstdFlags)
	
	// Create dispatcher
	dispatcher := memory.NewInMemoryDispatcher(logger)
	
	ctx := context.Background()
	
	// Start dispatcher
	if err := dispatcher.Start(ctx); err != nil {
		log.Fatalf("Failed to start dispatcher: %v", err)
	}
	defer dispatcher.Stop()
	
	// Create and subscribe handlers
	fantasyHandler := NewFantasyScoreCalculatorHandler()
	searchHandler := NewSearchIndexHandler()
	
	if err := dispatcher.Subscribe("PlayerMatchUpdated", fantasyHandler); err != nil {
		log.Fatalf("Failed to subscribe fantasy handler: %v", err)
	}
	
	if err := dispatcher.Subscribe("PlayerMatchUpdated", searchHandler); err != nil {
		log.Fatalf("Failed to subscribe search handler: %v", err)
	}
	
	// Simulate publishing events
	fmt.Println("\n📢 Publishing PlayerMatchUpdated events...")
	
	// Event 1: Jordan Dawson has a good game
	event1 := NewPlayerMatchUpdatedEvent(1, 1, 25, 12, 3)
	if err := dispatcher.Publish(ctx, event1); err != nil {
		log.Printf("Failed to publish event1: %v", err)
	}
	
	// Event 2: Another player has a different game
	event2 := NewPlayerMatchUpdatedEvent(2, 1, 18, 8, 1)
	if err := dispatcher.Publish(ctx, event2); err != nil {
		log.Printf("Failed to publish event2: %v", err)
	}
	
	fmt.Println("\n✅ Event publishing complete!")
}