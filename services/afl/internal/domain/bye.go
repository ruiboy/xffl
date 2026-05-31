package domain

import "context"

type Bye struct {
	ID           int
	RoundID      int
	ClubSeasonID int
}

// ByeWithClub extends Bye with the club name for display purposes.
type ByeWithClub struct {
	Bye
	ClubID   int
	ClubName string
}

type ByeRepository interface {
	FindByRoundID(ctx context.Context, roundID int) ([]Bye, error)
	FindByRoundIDWithClub(ctx context.Context, roundID int) ([]ByeWithClub, error)
	FindByRoundAndClub(ctx context.Context, roundID int, clubSeasonID int) (Bye, error)
	Upsert(ctx context.Context, roundID int, clubSeasonID int) (Bye, error)
}
