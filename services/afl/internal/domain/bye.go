package domain

import "context"

type Bye struct {
	ID           int
	RoundID      int
	ClubSeasonID int
}

type ByeRepository interface {
	FindByRoundID(ctx context.Context, roundID int) ([]Bye, error)
	FindByRoundAndClub(ctx context.Context, roundID int, clubSeasonID int) (Bye, error)
	Upsert(ctx context.Context, roundID int, clubSeasonID int) (Bye, error)
}
