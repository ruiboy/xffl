package events

import (
	"encoding/json"
	"fmt"
	"xffl/pkg/events"
)

// PlayerMatchUpdatedEvent represents when a player's match statistics are updated
type PlayerMatchUpdatedEvent struct {
	events.BaseEvent
	PlayerSeasonID uint         `json:"playerSeasonId"`
	ClubMatchID    uint         `json:"clubMatchId"`
	OldStats       PlayerStats  `json:"oldStats"`
	NewStats       PlayerStats  `json:"newStats"`
}

// PlayerStats represents the statistical data for a player match
type PlayerStats struct {
	Kicks     int `json:"kicks"`
	Handballs int `json:"handballs"`
	Marks     int `json:"marks"`
	Hitouts   int `json:"hitouts"`
	Tackles   int `json:"tackles"`
	Goals     int `json:"goals"`
	Behinds   int `json:"behinds"`
}

// NewPlayerMatchUpdatedEvent creates a new PlayerMatchUpdatedEvent
func NewPlayerMatchUpdatedEvent(playerSeasonID, clubMatchID uint, oldStats, newStats PlayerStats) *PlayerMatchUpdatedEvent {
	return &PlayerMatchUpdatedEvent{
		BaseEvent: events.NewBaseEvent(
			"AFL.PlayerMatchUpdated",
			"v1", 
			fmt.Sprintf("afl-player-season-%d", playerSeasonID),
		),
		PlayerSeasonID: playerSeasonID,
		ClubMatchID:    clubMatchID,
		OldStats:       oldStats,
		NewStats:       newStats,
	}
}

// EventData implements DomainEvent interface for serialization
func (e *PlayerMatchUpdatedEvent) EventData() map[string]interface{} {
	data, _ := json.Marshal(e)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}