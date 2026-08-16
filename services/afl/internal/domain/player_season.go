package domain

import "context"

type PlayerSeason struct {
	ID           int
	PlayerID     int
	ClubSeasonID int
	FromRoundID  *int
	ToRoundID    *int
}

// PlayerSeasonWithPlayer combines a player_season record with its player's name.
// Used to build candidate pools for fuzzy name matching during stats import.
type PlayerSeasonWithPlayer struct {
	PlayerSeasonID int
	PlayerID       int
	ClubSeasonID   int
	Name           string
}

// PlayerSeasonWithClub combines a player_season record with its player's name and
// club name. Used to build candidate pools for fuzzy name matching against a whole
// season's players (e.g. the FFL squad importer).
type PlayerSeasonWithClub struct {
	PlayerSeasonID int
	PlayerID       int
	Name           string
	ClubName       string
}

type PlayerSeasonRepository interface {
	Create(ctx context.Context, playerID, clubSeasonID int) (PlayerSeason, error)
	FindByID(ctx context.Context, id int) (PlayerSeason, error)
	FindByIDs(ctx context.Context, ids []int) (map[int]PlayerSeason, error)
	FindPlayersForPlayerSeasonIDs(ctx context.Context, ids []int) (map[int]Player, error)
	FindByClubSeasonIDWithPlayer(ctx context.Context, clubSeasonID int) ([]PlayerSeasonWithPlayer, error)
	FindBySeasonIDWithClub(ctx context.Context, seasonID int) ([]PlayerSeasonWithClub, error)
	FindIDsBySeasonID(ctx context.Context, seasonID int, nameQuery *string) ([]int, error)
	FindLatestByPlayerID(ctx context.Context, playerID int) (PlayerSeason, bool, error)
	FindByPlayerID(ctx context.Context, playerID int) ([]PlayerSeason, error)
}
