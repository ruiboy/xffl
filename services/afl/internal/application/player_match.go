package application

import (
	"context"
	"encoding/json"
	"log/slog"

	"xffl/contracts/events"
	"xffl/services/afl/internal/domain"
)

// UpdatePlayerMatch upserts a player match and recalculates the club match score
// using domain logic. Publishes an AFL.PlayerMatchUpdated event on success.
func (c *Commands) UpdatePlayerMatch(ctx context.Context, params domain.UpsertPlayerMatchParams) (domain.PlayerMatch, error) {
	var result domain.PlayerMatch
	var roundID int
	var matchID int
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

		matchID = clubMatch.MatchID
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

	result.MatchDataStatus = string(domain.MatchDataPartial)
	if match, err := c.matches.FindByID(ctx, matchID); err == nil {
		result.MatchDataStatus = string(match.DataStatus)
	}

	payload, err := json.Marshal(events.PlayerMatchUpdatedPayload{
		PlayerMatchID:  result.ID,
		PlayerSeasonID: result.PlayerSeasonID,
		ClubMatchID:    result.ClubMatchID,
		RoundID:        roundID,
		Status:         result.AFLPlayerMatchStatus(),
		Kicks:          result.Kicks,
		Handballs:      result.Handballs,
		Marks:          result.Marks,
		Hitouts:        result.Hitouts,
		Tackles:        result.Tackles,
		Goals:          result.Goals,
		Behinds:        result.Behinds,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to marshal PlayerMatchUpdated event", slog.Any("error", err))
		return result, nil
	}
	if err := c.dispatcher.Publish(ctx, events.PlayerMatchUpdated, payload); err != nil {
		slog.ErrorContext(ctx, "failed to publish PlayerMatchUpdated event", slog.Any("error", err))
	}

	return result, nil
}
