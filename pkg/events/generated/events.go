// Package generated contains auto-generated Go structs from AsyncAPI specifications
// DO NOT EDIT - This file is generated automatically
package generated

// PlayerMatchUpdatedPayload represents the payload for AFL.PlayerMatchUpdated events
type PlayerMatchUpdatedPayload struct {
	PlayerSeasonId int         `json:"playerSeasonId" validate:"required,min=1"` // ID of the player season record
	ClubMatchId    int         `json:"clubMatchId" validate:"required,min=1"`    // ID of the club match
	OldStats       PlayerStats `json:"oldStats" validate:"required"`            // Player statistics before the update
	NewStats       PlayerStats `json:"newStats" validate:"required"`            // Player statistics after the update
}

// PlayerStats represents statistical data for a player in a match
type PlayerStats struct {
	Kicks     int `json:"kicks" validate:"min=0"`     // Number of kicks
	Handballs int `json:"handballs" validate:"min=0"` // Number of handballs
	Marks     int `json:"marks" validate:"min=0"`     // Number of marks (catches)
	Hitouts   int `json:"hitouts" validate:"min=0"`   // Number of hitouts (ruck contests won)
	Tackles   int `json:"tackles" validate:"min=0"`   // Number of tackles
	Goals     int `json:"goals" validate:"min=0"`     // Number of goals scored
	Behinds   int `json:"behinds" validate:"min=0"`   // Number of behinds scored
}

// FantasyScoreCalculatedPayload represents the payload for FFL.FantasyScoreCalculated events
type FantasyScoreCalculatedPayload struct {
	PlayerSeasonId int    `json:"playerSeasonId" validate:"required,min=1"`                           // ID of the player season record
	ClubMatchId    int    `json:"clubMatchId" validate:"required,min=1"`                              // ID of the club match
	AflScore       int    `json:"aflScore" validate:"required,min=0"`                                 // Original AFL statistical score
	FantasyScore   int    `json:"fantasyScore" validate:"required,min=0"`                             // Calculated fantasy score based on league rules
	Source         string `json:"source" validate:"required,oneof=AFL.PlayerMatchUpdated Manual"`    // Source event that triggered this calculation
}