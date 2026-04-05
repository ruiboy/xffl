package domain

import (
	"context"
	"fmt"
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

// isBench returns true if this player is on the bench (has backup or interchange positions).
func (pm PlayerMatch) isBench() bool {
	return pm.BackupPositions != nil || pm.InterchangePosition != nil
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

// ValidateTeam enforces team composition rules against a set of team entries.
// It returns a descriptive error if any rule is violated, or nil if the team is valid.
// Teams need not be full — all constraints are upper bounds, not minimums.
func ValidateTeam(entries []UpsertPlayerMatchParams) error {
	starterCounts := make(map[Position]int)
	var benchPlayers []UpsertPlayerMatchParams
	interchangeCount := 0

	for _, e := range entries {
		isBench := e.BackupPositions != nil || e.InterchangePosition != nil
		if isBench {
			benchPlayers = append(benchPlayers, e)
			if e.InterchangePosition != nil {
				interchangeCount++
			}
		} else {
			if e.Position == nil {
				return fmt.Errorf("team: starter must have a position")
			}
			starterCounts[*e.Position]++
		}
	}

	// Rule 1: starter count per position ≤ PositionSlots[pos].
	for pos, count := range starterCounts {
		max, ok := PositionSlots[pos]
		if !ok {
			return fmt.Errorf("team: unknown position %q", pos)
		}
		if count > max {
			return fmt.Errorf("team: position %q has %d players, maximum is %d", pos, count, max)
		}
	}

	// Rule 2: total bench ≤ 4.
	if len(benchPlayers) > 4 {
		return fmt.Errorf("team: bench has %d players, maximum is 4", len(benchPlayers))
	}

	benchStarCount := 0
	coveredPositions := make(map[Position]bool)

	for _, bp := range benchPlayers {
		if bp.BackupPositions == nil {
			continue
		}
		positions := parsePositions(*bp.BackupPositions)
		// A backup star has backup positions consisting solely of "star".
		isBenchStar := len(positions) == 1 && positions[0] == PositionStar

		if isBenchStar {
			// Rule 3: at most 1 backup star.
			benchStarCount++
			if benchStarCount > 1 {
				return fmt.Errorf("team: at most 1 backup star allowed on the bench")
			}
		} else {
			// Rule 4: non-star bench players have exactly 2 backup positions, none "star".
			if len(positions) != 2 {
				return fmt.Errorf("team: non-star bench player must have exactly 2 backup positions, got %d", len(positions))
			}
			for _, pos := range positions {
				if pos == PositionStar {
					return fmt.Errorf("team: non-star bench player cannot list star as a backup position")
				}
				if _, ok := PositionSlots[pos]; !ok {
					return fmt.Errorf("team: unknown backup position %q", pos)
				}
				// Rule 5: each non-star position covered by at most one bench player.
				if coveredPositions[pos] {
					return fmt.Errorf("team: position %q is already covered by another bench player", pos)
				}
				coveredPositions[pos] = true
			}
		}
	}

	// Rule 6: at most 1 interchange position across all bench players.
	if interchangeCount > 1 {
		return fmt.Errorf("team: at most 1 interchange position allowed, got %d", interchangeCount)
	}

	// Rule 7: interchange position must be a recognised Position.
	for _, bp := range benchPlayers {
		if bp.InterchangePosition != nil {
			pos := Position(*bp.InterchangePosition)
			if _, ok := PositionSlots[pos]; !ok {
				return fmt.Errorf("team: interchange position %q is not a valid position", pos)
			}
		}
	}

	return nil
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
	FindByClubMatchID(ctx context.Context, clubMatchID int) ([]PlayerMatch, error)
	FindByID(ctx context.Context, id int) (PlayerMatch, error)
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
