package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
	"xffl/pkg/events"
	"xffl/pkg/events/memory"
)

// SimpleAFLEvent simulates an AFL player match updated event
type SimpleAFLEvent struct {
	events.BaseEvent
	PlayerSeasonID uint `json:"playerSeasonId"`
	ClubMatchID    uint `json:"clubMatchId"`
	OldStats       map[string]int `json:"oldStats"`
	NewStats       map[string]int `json:"newStats"`
}

func NewSimpleAFLEvent(playerSeasonID, clubMatchID uint, oldStats, newStats map[string]int) *SimpleAFLEvent {
	return &SimpleAFLEvent{
		BaseEvent:      events.NewBaseEvent("AFL.PlayerMatchUpdated", "v1", fmt.Sprintf("afl-player-season-%d", playerSeasonID)),
		PlayerSeasonID: playerSeasonID,
		ClubMatchID:    clubMatchID,
		OldStats:       oldStats,
		NewStats:       newStats,
	}
}

func (e *SimpleAFLEvent) EventData() map[string]interface{} {
	data, _ := json.Marshal(e)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}

// SimpleFFLEvent simulates an FFL fantasy score calculated event
type SimpleFFLEvent struct {
	events.BaseEvent
	PlayerSeasonID uint   `json:"playerSeasonId"`
	ClubMatchID    uint   `json:"clubMatchId"`
	AFLScore       int    `json:"aflScore"`
	FantasyScore   int    `json:"fantasyScore"`
	Source         string `json:"source"`
}

func NewSimpleFFLEvent(playerSeasonID, clubMatchID uint, aflScore, fantasyScore int, source string) *SimpleFFLEvent {
	return &SimpleFFLEvent{
		BaseEvent:      events.NewBaseEvent("FFL.FantasyScoreCalculated", "v1", fmt.Sprintf("ffl-player-season-%d", playerSeasonID)),
		PlayerSeasonID: playerSeasonID,
		ClubMatchID:    clubMatchID,
		AFLScore:       aflScore,
		FantasyScore:   fantasyScore,
		Source:         source,
	}
}

func (e *SimpleFFLEvent) EventData() map[string]interface{} {
	data, _ := json.Marshal(e)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}

// FFLFantasyCalculatorHandler simulates the FFL service
type FFLFantasyCalculatorHandler struct {
	name            string
	eventDispatcher events.EventDispatcher
	logger          *log.Logger
}

func NewFFLFantasyCalculatorHandler(dispatcher events.EventDispatcher, logger *log.Logger) *FFLFantasyCalculatorHandler {
	return &FFLFantasyCalculatorHandler{
		name:            "FFL.FantasyScoreCalculator",
		eventDispatcher: dispatcher,
		logger:          logger,
	}
}

func (h *FFLFantasyCalculatorHandler) HandlerName() string {
	return h.name
}

func (h *FFLFantasyCalculatorHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	if event.EventType() != "AFL.PlayerMatchUpdated" {
		return nil
	}

	eventData := event.EventData()
	
	playerSeasonID, _ := eventData["playerSeasonId"].(float64)
	clubMatchID, _ := eventData["clubMatchId"].(float64)
	
	newStatsData, _ := eventData["newStats"].(map[string]interface{})
	
	kicks := safeFloatToInt(newStatsData["kicks"])
	handballs := safeFloatToInt(newStatsData["handballs"])
	marks := safeFloatToInt(newStatsData["marks"])
	tackles := safeFloatToInt(newStatsData["tackles"])
	goals := safeFloatToInt(newStatsData["goals"])
	behinds := safeFloatToInt(newStatsData["behinds"])
	
	aflScore := kicks + handballs + marks + tackles + goals + behinds
	fantasyScore := (kicks * 3) + (handballs * 2) + (marks * 3) + (tackles * 4) + (goals * 6) + (behinds * 1)
	
	h.logger.Printf("🏆 FFL Service: Calculated fantasy score for player %d: AFL=%d, Fantasy=%d", 
		int(playerSeasonID), aflScore, fantasyScore)
	
	fantasyEvent := NewSimpleFFLEvent(
		uint(playerSeasonID), 
		uint(clubMatchID), 
		aflScore, 
		fantasyScore, 
		"AFL.PlayerMatchUpdated",
	)
	
	return h.eventDispatcher.Publish(ctx, fantasyEvent)
}

// NotificationHandler simulates a notification service
type NotificationHandler struct {
	name string
}

func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{name: "NotificationService"}
}

func (h *NotificationHandler) HandlerName() string {
	return h.name
}

func (h *NotificationHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	if event.EventType() == "FFL.FantasyScoreCalculated" {
		eventData := event.EventData()
		playerSeasonID, _ := eventData["playerSeasonId"].(float64)
		fantasyScore, _ := eventData["fantasyScore"].(float64)
		
		fmt.Printf("📱 Notification: Player %d scored %d fantasy points!\n", 
			int(playerSeasonID), int(fantasyScore))
	}
	return nil
}

func safeFloatToInt(value interface{}) int {
	if f, ok := value.(float64); ok {
		return int(f)
	}
	return 0
}

func main() {
	fmt.Println("🌐 Cross-Service Event Flow Simulation")
	fmt.Println("=======================================")
	fmt.Println("Simulating AFL → FFL → Notifications event chain")
	
	logger := log.New(os.Stdout, "[SHARED-EVENTS] ", log.LstdFlags)
	dispatcher := memory.NewInMemoryDispatcher(logger)
	
	ctx := context.Background()
	
	if err := dispatcher.Start(ctx); err != nil {
		log.Fatalf("Failed to start dispatcher: %v", err)
	}
	defer dispatcher.Stop()
	
	// Create handlers representing different services
	fflHandler := NewFFLFantasyCalculatorHandler(dispatcher, logger)
	notificationHandler := NewNotificationHandler()
	
	// Subscribe handlers
	if err := dispatcher.Subscribe("AFL.PlayerMatchUpdated", fflHandler); err != nil {
		log.Fatalf("Failed to subscribe FFL handler: %v", err)
	}
	
	if err := dispatcher.Subscribe("FFL.FantasyScoreCalculated", notificationHandler); err != nil {
		log.Fatalf("Failed to subscribe notification handler: %v", err)
	}
	
	fmt.Println("\n🏈 Simulating AFL Player Match Update...")
	
	// Create event data
	oldStats := map[string]int{
		"kicks": 20, "handballs": 10, "marks": 5, "tackles": 3, "goals": 1, "behinds": 1,
	}
	newStats := map[string]int{
		"kicks": 25, "handballs": 12, "marks": 8, "tackles": 5, "goals": 3, "behinds": 1,
	}
	
	aflEvent := NewSimpleAFLEvent(1, 1, oldStats, newStats)
	
	fmt.Printf("📊 AFL Service: Jordan Dawson updated stats - Kicks: %d→%d, Goals: %d→%d, Tackles: %d→%d\n",
		oldStats["kicks"], newStats["kicks"], oldStats["goals"], newStats["goals"], oldStats["tackles"], newStats["tackles"])
	
	// Publish AFL event
	if err := dispatcher.Publish(ctx, aflEvent); err != nil {
		log.Printf("Failed to publish AFL event: %v", err)
	}
	
	// Give events time to process
	time.Sleep(100 * time.Millisecond)
	
	fmt.Println("\n✅ Cross-service event flow complete!")
	fmt.Println("\nEvent Flow Summary:")
	fmt.Println("1. 🏈 AFL Service publishes PlayerMatchUpdated event")
	fmt.Println("2. 🏆 FFL Service calculates fantasy score and publishes FantasyScoreCalculated event")
	fmt.Println("3. 📱 Notification Service sends notification to user")
	fmt.Println("\nThis demonstrates how events can flow between microservices using a shared event dispatcher!")
}