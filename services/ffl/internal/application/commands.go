package application

import (
	"context"

	"xffl/services/ffl/internal/domain"
	sharedevents "xffl/shared/events"
)

// WriteRepos provides repository access within a transaction.
type WriteRepos struct {
	Leagues       domain.LeagueRepository
	Seasons       domain.SeasonRepository
	Rounds        domain.RoundRepository
	Matches       domain.MatchRepository
	Clubs         domain.ClubRepository
	ClubSeasons   domain.ClubSeasonRepository
	ClubMatches   domain.ClubMatchRepository
	Players       domain.PlayerRepository
	PlayerSeasons domain.PlayerSeasonRepository
	PlayerMatches domain.PlayerMatchRepository
}

// TxManager abstracts transactional execution.
type TxManager interface {
	WithTx(ctx context.Context, fn func(repos WriteRepos) error) error
}

// Commands handles all write and event-handling operations for the FFL service.
type Commands struct {
	tx            TxManager
	dispatcher    sharedevents.Dispatcher
	playerLookup  PlayerLookup
	matches       domain.MatchRepository
	clubMatches   domain.ClubMatchRepository
	clubSeasons   domain.ClubSeasonRepository
	rounds        domain.RoundRepository
	playerMatches domain.PlayerMatchRepository
	playerSeasons domain.PlayerSeasonRepository
}

// FindPlayerMatchesByClubMatch returns the player_matches for a club_match.
// Exposed so the dataops package can build event snapshots without reaching
// into Commands' private repositories.
func (c *Commands) FindPlayerMatchesByClubMatch(ctx context.Context, clubMatchID int) ([]domain.PlayerMatch, error) {
	return c.playerMatches.FindByClubMatchID(ctx, clubMatchID)
}

func NewCommands(
	tx TxManager,
	dispatcher sharedevents.Dispatcher,
	playerLookup PlayerLookup,
	matches domain.MatchRepository,
	clubMatches domain.ClubMatchRepository,
	clubSeasons domain.ClubSeasonRepository,
	rounds domain.RoundRepository,
	playerMatches domain.PlayerMatchRepository,
	playerSeasons domain.PlayerSeasonRepository,
) *Commands {
	return &Commands{
		tx:            tx,
		dispatcher:    dispatcher,
		playerLookup:  playerLookup,
		matches:       matches,
		clubMatches:   clubMatches,
		clubSeasons:   clubSeasons,
		rounds:        rounds,
		playerMatches: playerMatches,
		playerSeasons: playerSeasons,
	}
}
