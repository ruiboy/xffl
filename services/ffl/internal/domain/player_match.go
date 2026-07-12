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

// PlayerMatchStatus reflects the TM's explicit declaration for this player's role.
type PlayerMatchStatus string

const (
	PlayerMatchStatusNamed           PlayerMatchStatus = "named"            // starter playing; unused bench
	PlayerMatchStatusSubbedOut       PlayerMatchStatus = "subbed_out"       // starter explicitly replaced by TM
	PlayerMatchStatusSubbedIn        PlayerMatchStatus = "subbed_in"        // bench player brought in by TM
	PlayerMatchStatusInterchangedOut PlayerMatchStatus = "interchanged_out" // starter displaced by TM interchange
	PlayerMatchStatusInterchangedIn  PlayerMatchStatus = "interchanged_in"  // interchange bench player activated
)

// AFLStatus is the AFL participation status for this player in this match.
// Computed from AFL data and propagated via events; dnp is inferred after match finalisation.
type AFLStatus string

const (
	AFLStatusPlaying AFLStatus = "playing" // AFL match in progress, player has stats
	AFLStatusPlayed  AFLStatus = "played"  // AFL match final, player participated
	AFLStatusDNP     AFLStatus = "dnp"     // did not play in AFL match
	AFLStatusBye     AFLStatus = "bye"     // player's AFL club has a bye this round
)

type PlayerMatch struct {
	ID                  int
	ClubMatchID         int
	PlayerSeasonID      int
	Position            *Position
	Status              *PlayerMatchStatus
	AFLStatus           *AFLStatus
	BackupPositions     *string
	InterchangePosition *string
	DisplayOrder        int
	Notes               *string
	Score               int
	AFLPlayerMatchID    *int
}

// isBench returns true if this player is on the bench (has backup positions).
// InterchangePosition is always co-present with BackupPositions.
func (pm PlayerMatch) isBench() bool {
	return pm.BackupPositions != nil
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
func PositionPtr(p Position) *Position { return &p }

type PlayerMatchRepository interface {
	DeleteByClubMatchID(ctx context.Context, clubMatchID int) error
	DeleteByID(ctx context.Context, id int) error
	FindByClubMatchID(ctx context.Context, clubMatchID int) ([]PlayerMatch, error)
	FindByID(ctx context.Context, id int) (PlayerMatch, error)
	FindByPlayerSeasonID(ctx context.Context, playerSeasonID int) ([]PlayerMatch, error)
	FindByPlayerSeasonAndRound(ctx context.Context, playerSeasonID int, roundID int) (PlayerMatch, error)
	UpdateAFLPlayerMatchID(ctx context.Context, id int, aflPlayerMatchID int) error
	UpdateStatus(ctx context.Context, id int, status PlayerMatchStatus) error
	UpdatePosition(ctx context.Context, id int, position *Position) error
	UpdateAFLStatus(ctx context.Context, id int, status AFLStatus) error
	UpdateDisplayOrder(ctx context.Context, id int, displayOrder int) error
	AllAFLStatusesFinal(ctx context.Context, clubMatchID int) (bool, error)
	Upsert(ctx context.Context, params UpsertPlayerMatchParams) (PlayerMatch, error)
}

// UpsertPlayerMatchParams holds fields for creating or updating a PlayerMatch.
type UpsertPlayerMatchParams struct {
	ClubMatchID         int
	PlayerSeasonID      int
	Position            *Position
	Status              *PlayerMatchStatus
	AFLStatus           *AFLStatus
	BackupPositions     *string
	InterchangePosition *string
	DisplayOrder        int
	Notes               *string
	Score               *int
}
