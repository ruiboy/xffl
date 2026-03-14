package events

import (
	"encoding/json"
	"fmt"
	"xffl/pkg/events"
)

// FantasyScoreCalculatedEvent represents when a fantasy score is calculated for a player
type FantasyScoreCalculatedEvent struct {
	events.BaseEvent
	PlayerSeasonID uint `json:"playerSeasonId"`
	ClubMatchID    uint `json:"clubMatchId"`
	AFLScore       int  `json:"aflScore"`
	FantasyScore   int  `json:"fantasyScore"`
	Source         string `json:"source"` // e.g., "AFL.PlayerMatchUpdated"
}

// NewFantasyScoreCalculatedEvent creates a new FantasyScoreCalculatedEvent
func NewFantasyScoreCalculatedEvent(playerSeasonID, clubMatchID uint, aflScore, fantasyScore int, source string) *FantasyScoreCalculatedEvent {
	return &FantasyScoreCalculatedEvent{
		BaseEvent: events.NewBaseEvent(
			"FFL.FantasyScoreCalculated",
			"v1",
			fmt.Sprintf("ffl-player-season-%d", playerSeasonID),
		),
		PlayerSeasonID: playerSeasonID,
		ClubMatchID:    clubMatchID,
		AFLScore:       aflScore,
		FantasyScore:   fantasyScore,
		Source:         source,
	}
}

// EventData implements DomainEvent interface for serialization
func (e *FantasyScoreCalculatedEvent) EventData() map[string]interface{} {
	data, _ := json.Marshal(e)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}