// Package events defines the shared event types used for cross-service communication.
package events

// Event types for cross-service communication.
const (
	// AflPlayerMatchUpdated is published by the AFL service when a player's match stats change.
	AflPlayerMatchUpdated = "AFL.PlayerMatchUpdated"

	// AflMatchUpdated is published by the AFL service on match status transitions
	// (no_data→partial, partial→final). Carries full participation status snapshot.
	AflMatchUpdated = "AFL.MatchUpdated"

	// FflClubMatchUpdated is published by the FFL service on any change to a club's
	// team for a round: initial submission, correction, subs declared, or finalization.
	FflClubMatchUpdated = "FFL.ClubMatchUpdated"

	// FflPlayerMatchUpdated is published by the FFL service after recalculating a fantasy score.
	FflPlayerMatchUpdated = "FFL.PlayerMatchUpdated"

	// FflClubMatchScoreFinalized is published by the FFL service when a single club's score is locked
	// (AFL match final + FFL team final). Fires independently per club.
	FflClubMatchScoreFinalized = "FFL.ClubMatchScoreFinalized"

	// FflMatchScoreFinalized is published by the FFL service when both clubs in an FFL match are finalized.
	FflMatchScoreFinalized = "FFL.MatchScoreFinalized"
)

// AflPlayerMatchUpdatedPayload carries the full player match stats. Note there is no status field —
// participation status is carried exclusively by AflMatchUpdatedPayload, which tracks the match state.
type AflPlayerMatchUpdatedPayload struct {
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

// AflMatchUpdatedPayload is published on AFL match status transitions.
// PlayerSeasonIDStatusMap maps afl_player_season_id → status ("playing"/"played"/"dnp").
// On partial: players with stats → "playing". On final: players with stats → "played";
// all squad members without stats → "dnp".
type AflMatchUpdatedPayload struct {
	MatchID                 int            `json:"match_id"`
	RoundID                 int            `json:"round_id"`
	SeasonID                int            `json:"season_id"`
	MatchStatus             string         `json:"match_status"`
	PlayerSeasonIDStatusMap map[int]string `json:"player_season_id_status_map"`
}

// FflPlayerMatchInfo describes a single player's position and role in an FFL team snapshot.
type FflPlayerMatchInfo struct {
	Position            string `json:"position"`
	Status              string `json:"status"`
	BackupPositions     string `json:"backup_positions"`
	InterchangePosition string `json:"interchange_position"`
}

// FflClubMatchUpdatedPayload is published on any team change: submission, correction,
// subs declared, or finalization. PlayerMatches maps player_match_id → info snapshot.
type FflClubMatchUpdatedPayload struct {
	ClubMatchID   int                        `json:"club_match_id"`
	MatchID       int                        `json:"match_id"`
	RoundID       int                        `json:"round_id"`
	DataStatus    string                     `json:"data_status"`
	PlayerMatches map[int]FflPlayerMatchInfo `json:"player_matches"`
}

// FflPlayerMatchUpdatedPayload carries the calculated fantasy score for a single player.
type FflPlayerMatchUpdatedPayload struct {
	PlayerMatchID int `json:"player_match_id"`
	ClubMatchID   int `json:"club_match_id"`
	Score         int `json:"score"`
}

// FflClubMatchScoreFinalizedPayload carries identifiers for a single club's finalized score.
type FflClubMatchScoreFinalizedPayload struct {
	ClubMatchID int `json:"club_match_id"`
	MatchID     int `json:"match_id"`
}

// FflMatchScoreFinalizedPayload carries identifiers for a fully finalized FFL match.
type FflMatchScoreFinalizedPayload struct {
	MatchID int `json:"match_id"`
	RoundID int `json:"round_id"`
}
