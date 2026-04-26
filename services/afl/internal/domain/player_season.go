package domain

import "context"

type PlayerSeason struct {
	ID           int
	PlayerID     int
	ClubSeasonID int
	FromRoundID  *int
	ToRoundID    *int
}

type PlayerSeasonRepository interface {
	FindByID(ctx context.Context, id int) (PlayerSeason, error)
	FindPlayersForPlayerSeasonIDs(ctx context.Context, ids []int) (map[int]Player, error)
}
