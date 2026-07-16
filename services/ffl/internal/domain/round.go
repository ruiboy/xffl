package domain

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("not found")

type Round struct {
	ID         int
	Name       string
	SeasonID   int
	AFLRoundID int
	Type       RoundType
}

// RoundType classifies an FFL round as home-and-away ("minor") or the grand
// final. FFL models no finals lead-up rounds by type — everything but the grand
// final (the super-bye round included) is MINOR.
//
// The home-and-away ladder is built from MINOR rounds only; the grand final is
// excluded.
type RoundType string

const (
	RoundTypeMinor      RoundType = "MINOR"
	RoundTypeGrandFinal RoundType = "GRAND_FINAL"
)

// IsFinal reports whether the round is the grand final (the only finals round
// FFL distinguishes). Finals do not count toward the home-and-away ladder.
func (rt RoundType) IsFinal() bool {
	return rt == RoundTypeGrandFinal
}

type RoundRepository interface {
	FindBySeasonID(ctx context.Context, seasonID int) ([]Round, error)
	FindByID(ctx context.Context, id int) (Round, error)
	FindByAFLRoundID(ctx context.Context, aflRoundID int) (Round, error)
	Create(ctx context.Context, seasonID int, name string, aflRoundID int, roundType RoundType) (Round, error)
	Update(ctx context.Context, id int, name string, aflRoundID int, roundType RoundType) error
	SoftDelete(ctx context.Context, id int) error
}
