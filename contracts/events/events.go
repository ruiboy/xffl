// Package events defines the shared event types used for cross-service communication.
package events

// Event types for cross-service communication.
const (
	// PlayerMatchUpdated is published by the AFL service when a player's match stats change.
	PlayerMatchUpdated = "AFL.PlayerMatchUpdated"

	// AflMatchFinalized is published by the AFL service when afl.match.data_status transitions to final.
	AflMatchFinalized = "AFL.MatchFinalized"

	// FantasyScoreCalculated is published by the FFL service after recalculating a fantasy score.
	FantasyScoreCalculated = "FFL.FantasyScoreCalculated"

	// FflTeamSubmitted is published by the FFL service when ffl.club_match.data_status transitions to submitted.
	FflTeamSubmitted = "FFL.TeamSubmitted"

	// FflTeamFinalized is published by the FFL service when ffl.club_match.data_status transitions to final.
	FflTeamFinalized = "FFL.TeamFinalized"

	// FflClubMatchScoreFinalized is published by the FFL service when a single club's score is locked
	// (AFL match final + FFL team final). Fires independently per club.
	FflClubMatchScoreFinalized = "FFL.ClubMatchScoreFinalized"

	// FflMatchFinalized is published by the FFL service when both clubs in an FFL match are finalized.
	FflMatchFinalized = "FFL.MatchFinalized"
)

// PlayerMatchUpdatedPayload carries the full player match stats.
type PlayerMatchUpdatedPayload struct {
	PlayerMatchID  int    `json:"player_match_id"`
	PlayerSeasonID int    `json:"player_season_id"`
	ClubMatchID    int    `json:"club_match_id"`
	RoundID        int    `json:"round_id"`
	Status         string `json:"status"`
	Kicks          int    `json:"kicks"`
	Handballs      int    `json:"handballs"`
	Marks          int    `json:"marks"`
	Hitouts        int    `json:"hitouts"`
	Tackles        int    `json:"tackles"`
	Goals          int    `json:"goals"`
	Behinds        int    `json:"behinds"`
}

// AflMatchFinalizedPayload carries identifiers for the finalized AFL match.
type AflMatchFinalizedPayload struct {
	MatchID  int `json:"match_id"`
	SeasonID int `json:"season_id"`
	RoundID  int `json:"round_id"`
}

// FantasyScoreCalculatedPayload carries the calculated fantasy score.
type FantasyScoreCalculatedPayload struct {
	PlayerMatchID int `json:"player_match_id"`
	Score         int `json:"score"`
}

// FflTeamSubmittedPayload carries identifiers for the submitted FFL club match.
type FflTeamSubmittedPayload struct {
	ClubMatchID int `json:"club_match_id"`
	MatchID     int `json:"match_id"`
	RoundID     int `json:"round_id"`
}

// FflTeamFinalizedPayload carries identifiers for the finalized FFL club match.
type FflTeamFinalizedPayload struct {
	ClubMatchID int `json:"club_match_id"`
	MatchID     int `json:"match_id"`
	RoundID     int `json:"round_id"`
}

// FflClubMatchScoreFinalizedPayload carries identifiers for a single club's finalized score.
type FflClubMatchScoreFinalizedPayload struct {
	ClubMatchID int `json:"club_match_id"`
	MatchID     int `json:"match_id"`
}

// FflMatchFinalizedPayload carries identifiers for a fully finalized FFL match.
type FflMatchFinalizedPayload struct {
	MatchID int `json:"match_id"`
	RoundID int `json:"round_id"`
}
