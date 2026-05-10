package application

import (
	"context"
	"fmt"
	"log/slog"

	"xffl/services/afl/internal/domain"
	sharedevents "xffl/shared/events"
)

// ScoreCommands handles AFL score and ladder calculation use cases.
type ScoreCommands struct {
	matches     domain.MatchRepository
	clubMatches domain.ClubMatchRepository
	clubSeasons domain.ClubSeasonRepository
	rounds      domain.RoundRepository
	dispatcher  sharedevents.Dispatcher
}

func NewScoreCommands(
	matches domain.MatchRepository,
	clubMatches domain.ClubMatchRepository,
	clubSeasons domain.ClubSeasonRepository,
	rounds domain.RoundRepository,
	dispatcher sharedevents.Dispatcher,
) *ScoreCommands {
	return &ScoreCommands{
		matches:     matches,
		clubMatches: clubMatches,
		clubSeasons: clubSeasons,
		rounds:      rounds,
		dispatcher:  dispatcher,
	}
}

// ProcessAFLMatchFinalized derives and persists the match result, then recalculates
// the AFL ladder for the season.
func (c *ScoreCommands) ProcessAFLMatchFinalized(ctx context.Context, matchID, seasonID int) error {
	match, err := c.matches.FindByID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("load match %d: %w", matchID, err)
	}

	home, err := c.clubMatches.FindByID(ctx, match.Home.ID)
	if err != nil {
		return fmt.Errorf("load home club_match: %w", err)
	}
	away, err := c.clubMatches.FindByID(ctx, match.Away.ID)
	if err != nil {
		return fmt.Errorf("load away club_match: %w", err)
	}
	match.Home = home
	match.Away = away

	if err := c.matches.UpdateResult(ctx, matchID, match.DeriveResult()); err != nil {
		slog.WarnContext(ctx, "update match result failed", slog.Int("match_id", matchID), slog.Any("error", err))
	}

	if err := c.RecalculateAFLLadder(ctx, seasonID); err != nil {
		slog.WarnContext(ctx, "recalculate AFL ladder failed", slog.Int("season_id", seasonID), slog.Any("error", err))
	}

	return nil
}

// RecalculateAFLLadder rebuilds AFL ladder standings for the given season
// from all final matches. Idempotent — safe to call multiple times.
func (c *ScoreCommands) RecalculateAFLLadder(ctx context.Context, seasonID int) error {
	matches, err := c.matches.FindFinalBySeasonID(ctx, seasonID)
	if err != nil {
		return fmt.Errorf("load final matches: %w", err)
	}
	for _, cs := range domain.CalculateLadder(matches) {
		if err := c.clubSeasons.Update(ctx, cs); err != nil {
			slog.WarnContext(ctx, "update club season failed",
				slog.Int("club_season_id", cs.ID), slog.Any("error", err))
		}
	}
	return nil
}
