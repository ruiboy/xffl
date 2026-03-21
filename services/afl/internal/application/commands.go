package application

import (
	"context"

	"xffl/services/afl/internal/domain"
)

// WriteRepos provides repository access within a transaction.
type WriteRepos struct {
	PlayerMatches domain.PlayerMatchRepository
	ClubMatches   domain.ClubMatchRepository
}

// TxManager abstracts transactional execution.
type TxManager interface {
	WithTx(ctx context.Context, fn func(repos WriteRepos) error) error
}

// Commands handles all write operations for the AFL service.
type Commands struct {
	tx TxManager
}

func NewCommands(tx TxManager) *Commands {
	return &Commands{tx: tx}
}

// UpdatePlayerMatch upserts a player match and recalculates the club match score.
func (c *Commands) UpdatePlayerMatch(ctx context.Context, params domain.UpsertPlayerMatchParams) (domain.PlayerMatch, error) {
	var result domain.PlayerMatch
	err := c.tx.WithTx(ctx, func(repos WriteRepos) error {
		pm, err := repos.PlayerMatches.Upsert(ctx, params)
		if err != nil {
			return err
		}
		result = pm
		return repos.ClubMatches.RecalculateScore(ctx, params.ClubMatchID)
	})
	return result, err
}
