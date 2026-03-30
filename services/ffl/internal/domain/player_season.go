package domain

import "context"

type PlayerSeason struct {
	ID           int
	PlayerID     int
	ClubSeasonID int
}

type PlayerSeasonRepository interface {
	FindByClubSeasonID(ctx context.Context, clubSeasonID int) ([]PlayerSeason, error)
	FindByID(ctx context.Context, id int) (PlayerSeason, error)
}
