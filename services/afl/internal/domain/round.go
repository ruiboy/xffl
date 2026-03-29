package domain

import "context"

type Round struct {
	ID       int
	Name     string
	SeasonID int
}

type RoundRepository interface {
	FindBySeasonID(ctx context.Context, seasonID int) ([]Round, error)
	FindByID(ctx context.Context, id int) (Round, error)
	FindLatest(ctx context.Context) (Round, error)
}
