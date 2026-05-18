package application

import (
	"context"
	"fmt"

	"xffl/services/ffl/internal/domain"
)

// AddPlayerToSeason adds a player to a club season squad. The AFL player_season
// ID is the only cross-service handle the caller needs to provide; the FFL
// service resolves it to the underlying afl.player.id via Twirp and find-or-
// creates the ffl.player row.
func (c *Commands) AddPlayerToSeason(ctx context.Context, clubSeasonID, aflPlayerSeasonID int, fromRoundID, costCents *int) (domain.PlayerSeason, error) {
	aflPlayerID, err := c.playerLookup.LookupPlayerSeason(ctx, aflPlayerSeasonID)
	if err != nil {
		return domain.PlayerSeason{}, fmt.Errorf("lookup AFL player season: %w", err)
	}
	var result domain.PlayerSeason
	err = c.tx.WithTx(ctx, func(repos WriteRepos) error {
		player, err := repos.Players.FindByAFLPlayerID(ctx, aflPlayerID)
		if err != nil {
			player, err = repos.Players.Create(ctx, aflPlayerID)
			if err != nil {
				return err
			}
		}
		ps, err := repos.PlayerSeasons.Create(ctx, player.ID, clubSeasonID, fromRoundID, aflPlayerSeasonID, costCents)
		if err != nil {
			return err
		}
		result = ps
		return nil
	})
	return result, err
}

// UpdatePlayerSeasonDetails updates the notes for a player season.
func (c *Commands) UpdatePlayerSeasonDetails(ctx context.Context, id int, notes *string) (domain.PlayerSeason, error) {
	var result domain.PlayerSeason
	err := c.tx.WithTx(ctx, func(repos WriteRepos) error {
		ps, err := repos.PlayerSeasons.UpdateDetails(ctx, id, notes)
		if err != nil {
			return err
		}
		result = ps
		return nil
	})
	return result, err
}

// RemovePlayerFromSeason records the last round a player was in the squad, preserving history.
func (c *Commands) RemovePlayerFromSeason(ctx context.Context, playerSeasonID int, toRoundID int) error {
	return c.tx.WithTx(ctx, func(repos WriteRepos) error {
		return repos.PlayerSeasons.SetEndRound(ctx, playerSeasonID, toRoundID)
	})
}
