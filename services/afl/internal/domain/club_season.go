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

type ClubSeasonRepository interface {
	FindBySeasonID(ctx context.Context, seasonID int) ([]ClubSeason, error)
	FindByID(ctx context.Context, id int) (ClubSeason, error)
}
