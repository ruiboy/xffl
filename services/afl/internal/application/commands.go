package application

import (
	"context"
	"encoding/json"
	"log"

	"xffl/contracts/events"
	"xffl/services/afl/internal/domain"
	sharedevents "xffl/shared/events"
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
	tx         TxManager
	dispatcher sharedevents.Dispatcher
}

func NewCommands(tx TxManager, dispatcher sharedevents.Dispatcher) *Commands {
	return &Commands{tx: tx, dispatcher: dispatcher}
}

// UpdatePlayerMatch upserts a player match and recalculates the club match score
// using domain logic. Publishes an AFL.PlayerMatchUpdated event on success.
func (c *Commands) UpdatePlayerMatch(ctx context.Context, params domain.UpsertPlayerMatchParams) (domain.PlayerMatch, error) {
	var result domain.PlayerMatch
	var roundID int
	err := c.tx.WithTx(ctx, func(repos WriteRepos) error {
		pm, err := repos.PlayerMatches.Upsert(ctx, params)
		if err != nil {
			return err
		}
		result = pm

		playerMatches, err := repos.PlayerMatches.FindByClubMatchID(ctx, params.ClubMatchID)
		if err != nil {
			return err
		}
		clubMatch, err := repos.ClubMatches.FindByID(ctx, params.ClubMatchID)
		if err != nil {
			return err
		}

		roundID, err = repos.ClubMatches.FindRoundID(ctx, params.ClubMatchID)
		if err != nil {
			return err
		}

		clubMatch.PlayerMatches = playerMatches
		return repos.ClubMatches.UpdateScore(ctx, params.ClubMatchID, clubMatch.Score())
	})
	if err != nil {
		return result, err
	}

	payload, err := json.Marshal(events.PlayerMatchUpdatedPayload{
		PlayerMatchID:  result.ID,
		PlayerSeasonID: result.PlayerSeasonID,
		ClubMatchID:    result.ClubMatchID,
		RoundID:        roundID,
		Kicks:          result.Kicks,
		Handballs:      result.Handballs,
		Marks:          result.Marks,
		Hitouts:        result.Hitouts,
		Tackles:        result.Tackles,
		Goals:          result.Goals,
		Behinds:        result.Behinds,
	})
	if err != nil {
		log.Printf("AFL: failed to marshal PlayerMatchUpdated event: %v", err)
		return result, nil
	}
	if err := c.dispatcher.Publish(ctx, events.PlayerMatchUpdated, payload); err != nil {
		log.Printf("AFL: failed to publish PlayerMatchUpdated event: %v", err)
	}

	return result, nil
}
