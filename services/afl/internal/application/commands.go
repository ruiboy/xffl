package application

import (
	"context"

	"xffl/services/afl/internal/domain"
	sharedevents "xffl/shared/events"
)

// WriteRepos provides repository access within a transaction.
type WriteRepos struct {
	Players       domain.PlayerRepository
	PlayerSeasons domain.PlayerSeasonRepository
	PlayerMatches domain.PlayerMatchRepository
	ClubMatches   domain.ClubMatchRepository
}

// TxManager abstracts transactional execution.
type TxManager interface {
	WithTx(ctx context.Context, fn func(repos WriteRepos) error) error
}

// Commands handles all write operations for the AFL service.
type Commands struct {
	tx          TxManager
	matches     domain.MatchRepository
	clubMatches domain.ClubMatchRepository
	clubSeasons domain.ClubSeasonRepository
	rounds      domain.RoundRepository
	dispatcher  sharedevents.Dispatcher
}

func NewCommands(
	tx TxManager,
	matches domain.MatchRepository,
	clubMatches domain.ClubMatchRepository,
	clubSeasons domain.ClubSeasonRepository,
	rounds domain.RoundRepository,
	dispatcher sharedevents.Dispatcher,
) *Commands {
	return &Commands{
		tx:          tx,
		matches:     matches,
		clubMatches: clubMatches,
		clubSeasons: clubSeasons,
		rounds:      rounds,
		dispatcher:  dispatcher,
	}
}
