// Package events defines the shared event types used for cross-service communication.
package events

// Event types for cross-service communication.
const (
	// PlayerMatchUpdated is published by the AFL service when a player's match stats change.
	PlayerMatchUpdated = "AFL.PlayerMatchUpdated"

	// FantasyScoreCalculated is published by the FFL service after recalculating a fantasy score.
	FantasyScoreCalculated = "FFL.FantasyScoreCalculated"
)

// PlayerMatchUpdatedPayload carries the full player match stats.
type PlayerMatchUpdatedPayload struct {
	PlayerMatchID  int `json:"player_match_id"`
	PlayerSeasonID int `json:"player_season_id"`
	ClubMatchID    int `json:"club_match_id"`
	RoundID        int `json:"round_id"`
	Kicks          int `json:"kicks"`
	Handballs      int `json:"handballs"`
	Marks          int `json:"marks"`
	Hitouts        int `json:"hitouts"`
	Tackles        int `json:"tackles"`
	Goals          int `json:"goals"`
	Behinds        int `json:"behinds"`
}

// FantasyScoreCalculatedPayload carries the calculated fantasy score.
type FantasyScoreCalculatedPayload struct {
	PlayerMatchID int `json:"player_match_id"`
	Score         int `json:"score"`
}
