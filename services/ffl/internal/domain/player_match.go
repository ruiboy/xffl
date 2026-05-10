package domain

import (
	"context"
	"strings"
)

// Position represents a fantasy football position that determines scoring.
type Position string

const (
	PositionGoals     Position = "goals"
	PositionKicks     Position = "kicks"
	PositionHandballs Position = "handballs"
	PositionMarks     Position = "marks"
	PositionTackles   Position = "tackles"
	PositionHitouts   Position = "hitouts"
	PositionStar      Position = "star"
)

// PositionSlots defines the maximum number of starter slots per position.
var PositionSlots = map[Position]int{
	PositionGoals:     3,
	PositionKicks:     4,
	PositionHandballs: 4,
	PositionMarks:     2,
	PositionTackles:   2,
	PositionHitouts:   2,
	PositionStar:      1,
}

// Scoring multipliers per position.
const (
	GoalsMultiplier     = 5
	KicksMultiplier     = 1
	HandballsMultiplier = 1
	MarksMultiplier     = 2
	TacklesMultiplier   = 4
	HitoutsMultiplier   = 1
)

// PlayerMatchStatus reflects the player's AFL match status (denormalised from AFL data).
type PlayerMatchStatus string

const (
	PlayerMatchStatusNamed  PlayerMatchStatus = "named"  // selected in AFL team, match not yet played
	PlayerMatchStatusPlayed PlayerMatchStatus = "played" // played in the AFL match
	PlayerMatchStatusDNP    PlayerMatchStatus = "dnp"    // did not play
)

// AFLStats holds the AFL performance statistics used to calculate fantasy scores.
type AFLStats struct {
	Goals     int
	Kicks     int
	Handballs int
	Marks     int
	Tackles   int
	Hitouts   int
}

type PlayerMatch struct {
	ID                  int
	ClubMatchID         int
	PlayerSeasonID      int
	Position            *Position
	Status              *PlayerMatchStatus
	BackupPositions     *string
	InterchangePosition *string
	Score               int
	AFLPlayerMatchID    *int
}

// isBench returns true if this player is on the bench (has backup positions).
// InterchangePosition is always co-present with BackupPositions.
func (pm PlayerMatch) isBench() bool {
	return pm.BackupPositions != nil
}

// CalculateScore computes the fantasy score for this player based on their
// position and the given AFL match statistics. Returns 0 if position is nil.
func (pm PlayerMatch) CalculateScore(stats AFLStats) int {
	if pm.Position == nil {
		return 0
	}
	switch *pm.Position {
	case PositionGoals:
		return stats.Goals * GoalsMultiplier
	case PositionKicks:
		return stats.Kicks * KicksMultiplier
	case PositionHandballs:
		return stats.Handballs * HandballsMultiplier
	case PositionMarks:
		return stats.Marks * MarksMultiplier
	case PositionTackles:
		return stats.Tackles * TacklesMultiplier
	case PositionHitouts:
		return stats.Hitouts * HitoutsMultiplier
	case PositionStar:
		return stats.Goals*GoalsMultiplier +
			stats.Kicks*KicksMultiplier +
			stats.Handballs*HandballsMultiplier +
			stats.Marks*MarksMultiplier +
			stats.Tackles*TacklesMultiplier
	default:
		return 0
	}
}

// parsePositions splits a comma-separated position string into a slice of Position values.
func parsePositions(s string) []Position {
	parts := strings.Split(s, ",")
	out := make([]Position, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			out = append(out, Position(trimmed))
		}
	}
	return out
}

// Ptr helpers for use in struct literals.
func PositionPtr(p Position) *Position                            { return &p }
func PlayerMatchStatusPtr(s PlayerMatchStatus) *PlayerMatchStatus { return &s }

type PlayerMatchRepository interface {
	DeleteByClubMatchID(ctx context.Context, clubMatchID int) error
	DeleteByID(ctx context.Context, id int) error
	FindByClubMatchID(ctx context.Context, clubMatchID int) ([]PlayerMatch, error)
	FindByID(ctx context.Context, id int) (PlayerMatch, error)
	FindByPlayerSeasonAndRound(ctx context.Context, playerSeasonID int, roundID int) (PlayerMatch, error)
	UpdateAFLPlayerMatchID(ctx context.Context, id int, aflPlayerMatchID int) error
	UpdateStatus(ctx context.Context, id int, status PlayerMatchStatus) error
	Upsert(ctx context.Context, params UpsertPlayerMatchParams) (PlayerMatch, error)
}

// UpsertPlayerMatchParams holds fields for creating or updating a PlayerMatch.
type UpsertPlayerMatchParams struct {
	ClubMatchID         int
	PlayerSeasonID      int
	Position            *Position
	Status              *PlayerMatchStatus
	BackupPositions     *string
	InterchangePosition *string
	Score               *int
}
