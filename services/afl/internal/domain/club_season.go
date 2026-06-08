package domain

import "context"

type ClubSeason struct {
	ID                int
	ClubID            int
	SeasonID          int
	Played            int
	Won               int
	Lost              int
	Drawn             int
	For               int
	Against           int
	PremiershipPoints int
}

// Percentage is the AFL ladder tie-breaker: points scored as a percentage of points conceded.
// Undefined (e.g. before a club has conceded anything) is reported as 0.
func (cs ClubSeason) Percentage() float64 {
	if cs.Against == 0 {
		return 0
	}
	return float64(cs.For) / float64(cs.Against) * 100
}

type ClubSeasonRepository interface {
	FindBySeasonID(ctx context.Context, seasonID int) ([]ClubSeason, error)
	FindByID(ctx context.Context, id int) (ClubSeason, error)
	Update(ctx context.Context, cs ClubSeason) error
}
