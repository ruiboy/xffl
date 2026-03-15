package application

import (
	"context"

	"xffl/services/afl/internal/domain"
)

// Queries handles all read operations for the AFL service.
type Queries struct {
	clubs        domain.ClubRepository
	seasons      domain.SeasonRepository
	rounds       domain.RoundRepository
	matches      domain.MatchRepository
	clubSeasons  domain.ClubSeasonRepository
	clubMatches  domain.ClubMatchRepository
	players      domain.PlayerRepository
	playerMatches domain.PlayerMatchRepository
}

func NewQueries(
	clubs domain.ClubRepository,
	seasons domain.SeasonRepository,
	rounds domain.RoundRepository,
	matches domain.MatchRepository,
	clubSeasons domain.ClubSeasonRepository,
	clubMatches domain.ClubMatchRepository,
	players domain.PlayerRepository,
	playerMatches domain.PlayerMatchRepository,
) *Queries {
	return &Queries{
		clubs:         clubs,
		seasons:       seasons,
		rounds:        rounds,
		matches:       matches,
		clubSeasons:   clubSeasons,
		clubMatches:   clubMatches,
		players:       players,
		playerMatches: playerMatches,
	}
}

func (q *Queries) GetClubs(ctx context.Context) ([]domain.Club, error) {
	return q.clubs.FindAll(ctx)
}

func (q *Queries) GetClub(ctx context.Context, id int) (domain.Club, error) {
	return q.clubs.FindByID(ctx, id)
}

func (q *Queries) GetSeasons(ctx context.Context) ([]domain.Season, error) {
	return q.seasons.FindAll(ctx)
}

func (q *Queries) GetSeason(ctx context.Context, id int) (domain.Season, error) {
	return q.seasons.FindByID(ctx, id)
}

func (q *Queries) GetRounds(ctx context.Context, seasonID int) ([]domain.Round, error) {
	return q.rounds.FindBySeasonID(ctx, seasonID)
}

func (q *Queries) GetMatch(ctx context.Context, id int) (domain.Match, error) {
	return q.matches.FindByID(ctx, id)
}

func (q *Queries) GetMatches(ctx context.Context, roundID int) ([]domain.Match, error) {
	return q.matches.FindByRoundID(ctx, roundID)
}

func (q *Queries) GetClubSeasons(ctx context.Context, seasonID int) ([]domain.ClubSeason, error) {
	return q.clubSeasons.FindBySeasonID(ctx, seasonID)
}

func (q *Queries) GetClubMatches(ctx context.Context, matchID int) ([]domain.ClubMatch, error) {
	return q.clubMatches.FindByMatchID(ctx, matchID)
}

func (q *Queries) GetPlayerMatches(ctx context.Context, clubMatchID int) ([]domain.PlayerMatch, error) {
	return q.playerMatches.FindByClubMatchID(ctx, clubMatchID)
}

func (q *Queries) GetPlayers(ctx context.Context, clubID int) ([]domain.Player, error) {
	return q.players.FindByClubID(ctx, clubID)
}
