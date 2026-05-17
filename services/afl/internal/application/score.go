package application

import (
	"context"
	"fmt"
	"log/slog"

	"xffl/services/afl/internal/domain"
)

// RecalculateAFLLadder rebuilds AFL ladder standings for the given season
// from all final matches. Idempotent — safe to call multiple times.
func (c *Commands) RecalculateAFLLadder(ctx context.Context, seasonID int) error {
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
