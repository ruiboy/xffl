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
	AFLRoundID *int
}

type RoundRepository interface {
	FindBySeasonID(ctx context.Context, seasonID int) ([]Round, error)
	FindByID(ctx context.Context, id int) (Round, error)
	FindByAFLRoundID(ctx context.Context, aflRoundID int) (Round, error)
}
