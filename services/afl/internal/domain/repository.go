package domain

import "context"

type ClubRepository interface {
	FindAll(ctx context.Context) ([]Club, error)
	FindByID(ctx context.Context, id int) (Club, error)
}

type SeasonRepository interface {
	FindAll(ctx context.Context) ([]Season, error)
	FindByID(ctx context.Context, id int) (Season, error)
}

type RoundRepository interface {
	FindBySeasonID(ctx context.Context, seasonID int) ([]Round, error)
	FindByID(ctx context.Context, id int) (Round, error)
}

type MatchRepository interface {
	FindByRoundID(ctx context.Context, roundID int) ([]Match, error)
	FindByID(ctx context.Context, id int) (Match, error)
}

type ClubMatchRepository interface {
	FindByMatchID(ctx context.Context, matchID int) ([]ClubMatch, error)
	FindByID(ctx context.Context, id int) (ClubMatch, error)
}

type PlayerMatchRepository interface {
	FindByClubMatchID(ctx context.Context, clubMatchID int) ([]PlayerMatch, error)
	FindByID(ctx context.Context, id int) (PlayerMatch, error)
}

type PlayerRepository interface {
	FindByClubID(ctx context.Context, clubID int) ([]Player, error)
	FindByID(ctx context.Context, id int) (Player, error)
}

type ClubSeasonRepository interface {
	FindBySeasonID(ctx context.Context, seasonID int) ([]ClubSeason, error)
	FindByID(ctx context.Context, id int) (ClubSeason, error)
}

type PlayerSeasonRepository interface {
	FindByID(ctx context.Context, id int) (PlayerSeason, error)
}
