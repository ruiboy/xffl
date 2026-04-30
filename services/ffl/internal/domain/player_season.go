package domain

import "context"

type PlayerSeason struct {
	ID                 int
	PlayerID           int
	ClubSeasonID       int
	AFLPlayerSeasonID  *int
}

type PlayerSeasonRepository interface {
	FindByClubSeasonID(ctx context.Context, clubSeasonID int) ([]PlayerSeason, error)
	FindByID(ctx context.Context, id int) (PlayerSeason, error)
	FindByAFLPlayerSeasonID(ctx context.Context, aflPlayerSeasonID int) ([]PlayerSeason, error)
	FindPlayersForPlayerSeasonIDs(ctx context.Context, ids []int) (map[int]Player, error)
	Create(ctx context.Context, playerID int, clubSeasonID int, fromRoundID *int, aflPlayerSeasonID *int) (PlayerSeason, error)
	SetEndRound(ctx context.Context, id int, toRoundID int) error
	Delete(ctx context.Context, id int) error
}
