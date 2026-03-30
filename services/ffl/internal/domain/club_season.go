package domain

import "context"

type ClubSeason struct {
	ID       int
	ClubID   int
	SeasonID int
	Played   int
	Won      int
	Lost     int
	Drawn    int
	For      int
	Against  int
}

// Percentage returns the club's season percentage (For / Against * 100).
// Returns 0 when Against is zero.
func (cs ClubSeason) Percentage() float64 {
	if cs.Against == 0 {
		return 0
	}
	return float64(cs.For) / float64(cs.Against) * 100
}

type ClubSeasonRepository interface {
	FindBySeasonID(ctx context.Context, seasonID int) ([]ClubSeason, error)
	FindByID(ctx context.Context, id int) (ClubSeason, error)
}
