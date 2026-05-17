package application

import (
	"context"
	"fmt"
	"log/slog"
)

// ProcessAFLMatchFinalized derives and persists the match result, then recalculates
// the AFL ladder for the season.
func (c *Commands) ProcessAFLMatchFinalized(ctx context.Context, matchID, seasonID int) error {
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
