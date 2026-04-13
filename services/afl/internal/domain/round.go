package domain

import (
	"context"
	"time"
)

type Round struct {
	ID       int
	Name     string
	SeasonID int
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
