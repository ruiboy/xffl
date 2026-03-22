package domain

import "context"

type PlayerSeason struct {
	ID           int
	PlayerID     int
	ClubSeasonID int
}

type PlayerSeasonRepository interface {
	FindByID(ctx context.Context, id int) (PlayerSeason, error)
}
