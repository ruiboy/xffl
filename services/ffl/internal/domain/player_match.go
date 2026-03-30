package domain

import "context"

// Position represents a fantasy football position that determines scoring.
type Position string

const (
	PositionGoals    Position = "goals"
	PositionKicks    Position = "kicks"
	PositionHandballs Position = "handballs"
	PositionMarks    Position = "marks"
	PositionTackles  Position = "tackles"
	PositionHitouts  Position = "hitouts"
	PositionStar     Position = "star"
)

// Scoring multipliers per position.
const (
	GoalsMultiplier    = 5
	KicksMultiplier    = 1
	HandballsMultiplier = 1
	MarksMultiplier    = 2
	TacklesMultiplier  = 4
	HitoutsMultiplier  = 1
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
	Goals    int
	Kicks    int
	Handballs int
	Marks    int
	Tackles  int
	Hitouts  int
}

type PlayerMatch struct {
	ID                  int
	ClubMatchID         int
	PlayerSeasonID      int
	Position            Position
	Status              PlayerMatchStatus
	BackupPositions     *string
	InterchangePosition *string
	Score               int
}

// isBench returns true if this player is on the bench (has backup or interchange positions).
func (pm PlayerMatch) isBench() bool {
	return pm.BackupPositions != nil || pm.InterchangePosition != nil
}

// CalculateScore computes the fantasy score for this player based on their
// position and the given AFL match statistics.
func (pm PlayerMatch) CalculateScore(stats AFLStats) int {
	switch pm.Position {
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

type PlayerMatchRepository interface {
	FindByClubMatchID(ctx context.Context, clubMatchID int) ([]PlayerMatch, error)
	FindByID(ctx context.Context, id int) (PlayerMatch, error)
	Upsert(ctx context.Context, params UpsertPlayerMatchParams) (PlayerMatch, error)
}

// UpsertPlayerMatchParams holds fields for creating or updating a PlayerMatch.
type UpsertPlayerMatchParams struct {
	ClubMatchID         int
	PlayerSeasonID      int
	Position            Position
	Status              PlayerMatchStatus
	BackupPositions     *string
	InterchangePosition *string
	Score               *int
}
