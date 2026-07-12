package domain

import (
	"context"
	"time"
)

type Round struct {
	ID       int
	Name     string
	SeasonID int
	Type     RoundType
}

// RoundType classifies a round as home-and-away ("minor") or a specific finals
// round.
//
// The home-and-away ladder is built from MINOR rounds only; finals are excluded.
type RoundType string

const (
	RoundTypeMinor            RoundType = "MINOR"
	RoundTypeWildcardFinal    RoundType = "WILDCARD_FINAL"
	RoundTypeQualifyingFinal  RoundType = "QUALIFYING_FINAL"
	RoundTypeEliminationFinal RoundType = "ELIMINATION_FINAL"
	RoundTypeSemiFinal        RoundType = "SEMI_FINAL"
	RoundTypePreliminaryFinal RoundType = "PRELIMINARY_FINAL"
	RoundTypeGrandFinal       RoundType = "GRAND_FINAL"
)

// IsFinal reports whether the round is part of the finals series (i.e. not a
// home-and-away round). Finals do not count toward the home-and-away ladder.
// An unknown/empty value is treated as non-final so that a mislabelled round
// counts toward the ladder (a visible wrong total) rather than silently
// vanishing from it.
func (rt RoundType) IsFinal() bool {
	switch rt {
	case RoundTypeWildcardFinal, RoundTypeQualifyingFinal, RoundTypeEliminationFinal,
		RoundTypeSemiFinal, RoundTypePreliminaryFinal, RoundTypeGrandFinal:
		return true
	default:
		return false
	}
}

// RoundWithStart pairs a round with the start time of its first match.
type RoundWithStart struct {
	Round          Round
	FirstMatchTime time.Time
}

type RoundRepository interface {
	FindBySeasonID(ctx context.Context, seasonID int) ([]Round, error)
	FindByID(ctx context.Context, id int) (Round, error)
	// FindNeighbours returns at most two rounds: the most recently started
	// (first_match_dt <= asOf) and the first upcoming (first_match_dt > asOf).
	FindNeighbours(ctx context.Context, asOf time.Time) ([]RoundWithStart, error)
}
