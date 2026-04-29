package application

import (
	"context"

	"xffl/services/afl/internal/domain"
	"xffl/shared/clock"
)

// Queries handles all read operations for the AFL service.
type Queries struct {
	clock         clock.Clock
	clubs         domain.ClubRepository
	seasons       domain.SeasonRepository
	rounds        domain.RoundRepository
	matches       domain.MatchRepository
	clubSeasons   domain.ClubSeasonRepository
	clubMatches   domain.ClubMatchRepository
	players       domain.PlayerRepository
	playerMatches domain.PlayerMatchRepository
	playerSeasons domain.PlayerSeasonRepository
}

func NewQueries(
	clk clock.Clock,
	clubs domain.ClubRepository,
	seasons domain.SeasonRepository,
	rounds domain.RoundRepository,
	matches domain.MatchRepository,
	clubSeasons domain.ClubSeasonRepository,
	clubMatches domain.ClubMatchRepository,
	players domain.PlayerRepository,
	playerMatches domain.PlayerMatchRepository,
	playerSeasons domain.PlayerSeasonRepository,
) *Queries {
	return &Queries{
		clock:         clk,
		clubs:         clubs,
		seasons:       seasons,
		rounds:        rounds,
		matches:       matches,
		clubSeasons:   clubSeasons,
		clubMatches:   clubMatches,
		players:       players,
		playerMatches: playerMatches,
		playerSeasons: playerSeasons,
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

func (q *Queries) GetRound(ctx context.Context, id int) (domain.Round, error) {
	return q.rounds.FindByID(ctx, id)
}

func (q *Queries) GetRounds(ctx context.Context, seasonID int) ([]domain.Round, error) {
	return q.rounds.FindBySeasonID(ctx, seasonID)
}

func (q *Queries) GetMatch(ctx context.Context, id int) (domain.Match, error) {
	return q.matches.FindByID(ctx, id)
}

func (q *Queries) GetMatchWithDetails(ctx context.Context, id int) (domain.Match, error) {
	return q.matches.FindByIDWithDetails(ctx, id)
}

func (q *Queries) GetMatches(ctx context.Context, roundID int) ([]domain.Match, error) {
	return q.matches.FindByRoundID(ctx, roundID)
}

func (q *Queries) GetClubSeasons(ctx context.Context, seasonID int) ([]domain.ClubSeason, error) {
	return q.clubSeasons.FindBySeasonID(ctx, seasonID)
}

func (q *Queries) GetClubMatch(ctx context.Context, id int) (domain.ClubMatch, error) {
	return q.clubMatches.FindByID(ctx, id)
}

func (q *Queries) GetClubMatches(ctx context.Context, matchID int) ([]domain.ClubMatch, error) {
	return q.clubMatches.FindByMatchID(ctx, matchID)
}

func (q *Queries) GetPlayerMatches(ctx context.Context, clubMatchID int) ([]domain.PlayerMatch, error) {
	return q.playerMatches.FindByClubMatchID(ctx, clubMatchID)
}

func (q *Queries) GetPlayer(ctx context.Context, id int) (domain.Player, error) {
	return q.players.FindByID(ctx, id)
}

func (q *Queries) SearchPlayers(ctx context.Context, query string) ([]domain.Player, error) {
	return q.players.Search(ctx, query)
}

func (q *Queries) GetClubsByIDs(ctx context.Context, ids []int) (map[int]domain.Club, error) {
	return q.clubs.FindByIDs(ctx, ids)
}

func (q *Queries) GetMatchesByIDs(ctx context.Context, ids []int) (map[int]domain.Match, error) {
	return q.matches.FindByIDs(ctx, ids)
}

func (q *Queries) GetPlayersForPlayerSeasonIDs(ctx context.Context, ids []int) (map[int]domain.Player, error) {
	return q.playerSeasons.FindPlayersForPlayerSeasonIDs(ctx, ids)
}

// GetClubForClubSeason resolves the club for a club_season record.
func (q *Queries) GetClubForClubSeason(ctx context.Context, clubSeasonID int) (domain.Club, error) {
	cs, err := q.clubSeasons.FindByID(ctx, clubSeasonID)
	if err != nil {
		return domain.Club{}, err
	}
	return q.clubs.FindByID(ctx, cs.ClubID)
}

func (q *Queries) GetPlayerSeasonByID(ctx context.Context, id int) (domain.PlayerSeason, error) {
	return q.playerSeasons.FindByID(ctx, id)
}

func (q *Queries) GetPlayerSeasonsByIDs(ctx context.Context, ids []int) (map[int]domain.PlayerSeason, error) {
	return q.playerSeasons.FindByIDs(ctx, ids)
}

func (q *Queries) GetPlayerMatchByID(ctx context.Context, id int) (domain.PlayerMatch, error) {
	return q.playerMatches.FindByID(ctx, id)
}

func (q *Queries) GetPlayerMatchesByPlayerSeasonID(ctx context.Context, playerSeasonID int) ([]domain.PlayerMatch, error) {
	return q.playerMatches.FindByPlayerSeasonID(ctx, playerSeasonID)
}

func (q *Queries) GetClubSeasonByID(ctx context.Context, id int) (domain.ClubSeason, error) {
	return q.clubSeasons.FindByID(ctx, id)
}

func (q *Queries) GetPlayerSeasonStats(ctx context.Context, ids []int) ([]domain.PlayerSeasonStats, error) {
	return q.playerMatches.FindStatsByPlayerSeasonIDs(ctx, ids)
}

// GetPlayerForPlayerSeason resolves the player for a player_season record.
func (q *Queries) GetPlayerForPlayerSeason(ctx context.Context, playerSeasonID int) (domain.Player, error) {
	ps, err := q.playerSeasons.FindByID(ctx, playerSeasonID)
	if err != nil {
		return domain.Player{}, err
	}
	return q.players.FindByID(ctx, ps.PlayerID)
}
