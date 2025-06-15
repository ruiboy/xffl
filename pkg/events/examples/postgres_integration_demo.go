package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
	
	"xffl/pkg/events"
	"xffl/pkg/events/postgres"
)

// SimplePlayerMatchEvent simulates AFL domain event
type SimplePlayerMatchEvent struct {
	events.BaseEvent
	PlayerSeasonID uint `json:"playerSeasonId"`
	ClubMatchID    uint `json:"clubMatchId"`
	OldStats       map[string]int `json:"oldStats"`
	NewStats       map[string]int `json:"newStats"`
}

func NewSimplePlayerMatchEvent(playerSeasonID, clubMatchID uint, oldStats, newStats map[string]int) *SimplePlayerMatchEvent {
	return &SimplePlayerMatchEvent{
		BaseEvent:      events.NewBaseEvent("AFL.PlayerMatchUpdated", "v1", fmt.Sprintf("afl-player-season-%d", playerSeasonID)),
		PlayerSeasonID: playerSeasonID,
		ClubMatchID:    clubMatchID,
		OldStats:       oldStats,
		NewStats:       newStats,
	}
}

func (e *SimplePlayerMatchEvent) EventData() map[string]interface{} {
	data, _ := json.Marshal(e)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}

// MockFFLHandler simulates FFL service
type MockFFLHandler struct {
	name string
}

func (h *MockFFLHandler) HandlerName() string {
	return "FFL.FantasyScoreCalculator"
}

func (h *MockFFLHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	if event.EventType() != "AFL.PlayerMatchUpdated" {
		return nil
	}

	eventData := event.EventData()
	playerSeasonID := eventData["playerSeasonId"].(float64)
	
	newStatsData := eventData["newStats"].(map[string]interface{})
	kicks := int(newStatsData["kicks"].(float64))
	handballs := int(newStatsData["handballs"].(float64))
	goals := int(newStatsData["goals"].(float64))
	
	fantasyScore := (kicks * 3) + (handballs * 2) + (goals * 6)
	
	fmt.Printf("🏆 FFL Service: Calculated fantasy score for player %d: %d points\n", 
		int(playerSeasonID), fantasyScore)
	
	return nil
}

func main() {
	fmt.Println("🐘 PostgreSQL Cross-Service Integration Demo")
	fmt.Println("=============================================")
	fmt.Println("This demonstrates AFL → FFL events via PostgreSQL LISTEN/NOTIFY")
	
	eventLogger := log.New(os.Stdout, "[DEMO-EVENTS] ", log.LstdFlags)
	connStr := "user=postgres dbname=xffl sslmode=disable"
	
	// Create publisher dispatcher (simulates AFL service)
	publisher, err := postgres.NewPostgresDispatcher(connStr, eventLogger)
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}
	
	// Create subscriber dispatcher (simulates FFL service)
	subscriber, err := postgres.NewPostgresDispatcher(connStr, eventLogger)
	if err != nil {
		log.Fatalf("Failed to create subscriber: %v", err)
	}
	
	ctx := context.Background()
	
	// Start both dispatchers
	if err := publisher.Start(ctx); err != nil {
		log.Fatalf("Failed to start publisher: %v", err)
	}
	defer publisher.Stop()
	
	if err := subscriber.Start(ctx); err != nil {
		log.Fatalf("Failed to start subscriber: %v", err)
	}
	defer subscriber.Stop()
	
	// Subscribe FFL handler to AFL events
	fflHandler := &MockFFLHandler{}
	if err := subscriber.Subscribe("AFL.PlayerMatchUpdated", fflHandler); err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	
	// Give listener time to connect
	time.Sleep(1 * time.Second)
	
	fmt.Println("\n🏈 Publishing AFL PlayerMatchUpdated event...")
	
	// Create and publish AFL event
	oldStats := map[string]int{
		"kicks": 20, "handballs": 10, "goals": 1,
	}
	newStats := map[string]int{
		"kicks": 25, "handballs": 12, "goals": 3,
	}
	
	aflEvent := NewSimplePlayerMatchEvent(1, 1, oldStats, newStats)
	
	fmt.Printf("📊 Jordan Dawson stats: Kicks %d→%d, Goals %d→%d\n",
		oldStats["kicks"], newStats["kicks"], oldStats["goals"], newStats["goals"])
	
	// Publish event
	if err := publisher.Publish(ctx, aflEvent); err != nil {
		log.Fatalf("Failed to publish event: %v", err)
	}
	
	// Wait for event processing
	time.Sleep(2 * time.Second)
	
	fmt.Println("\n✅ Cross-service PostgreSQL event flow complete!")
	fmt.Println("\nEvent Flow:")
	fmt.Println("1. 🏈 AFL service publishes PlayerMatchUpdated via PostgreSQL NOTIFY")
	fmt.Println("2. 🏆 FFL service receives event via PostgreSQL LISTEN")
	fmt.Println("3. 📱 FFL service calculates fantasy score")
	fmt.Println("\nThis same pattern works between your actual AFL and FFL services!")
}