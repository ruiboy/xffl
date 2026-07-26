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

// RoundType classifies an FFL round as home-and-away ("minor") or a final (semi
// or grand). The super-bye round is a home-and-away round and stays MINOR.
//
// The home-and-away ladder is built from MINOR rounds only; finals are excluded.
type RoundType string

const (
	RoundTypeMinor      RoundType = "MINOR"
	RoundTypeSemiFinal  RoundType = "SEMI_FINAL"
	RoundTypeGrandFinal RoundType = "GRAND_FINAL"
)

// IsFinal reports whether the round is a final (semi or grand). Finals do not
// count toward the home-and-away ladder.
func (rt RoundType) IsFinal() bool {
	return rt == RoundTypeSemiFinal || rt == RoundTypeGrandFinal
}

type RoundRepository interface {
	FindBySeasonID(ctx context.Context, seasonID int) ([]Round, error)
	FindByID(ctx context.Context, id int) (Round, error)
	FindByAFLRoundID(ctx context.Context, aflRoundID int) (Round, error)
	Create(ctx context.Context, seasonID int, name string, aflRoundID int, roundType RoundType) (Round, error)
	Update(ctx context.Context, id int, name string, aflRoundID int, roundType RoundType) error
	SoftDelete(ctx context.Context, id int) error
}
