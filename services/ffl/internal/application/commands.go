package application

import (
	"context"

	"xffl/services/ffl/internal/domain"
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

// Commands handles all write operations for the FFL service.
type Commands struct {
	tx TxManager
}

func NewCommands(tx TxManager) *Commands {
	return &Commands{tx: tx}
}

// CreatePlayer creates a new player.
func (c *Commands) CreatePlayer(ctx context.Context, name string) (domain.Player, error) {
	var result domain.Player
	err := c.tx.WithTx(ctx, func(repos WriteRepos) error {
		p, err := repos.Players.Create(ctx, name)
		if err != nil {
			return err
		}
		result = p
		return nil
	})
	return result, err
}

// UpdatePlayer updates an existing player's name.
func (c *Commands) UpdatePlayer(ctx context.Context, id int, name string) (domain.Player, error) {
	var result domain.Player
	err := c.tx.WithTx(ctx, func(repos WriteRepos) error {
		p, err := repos.Players.Update(ctx, id, name)
		if err != nil {
			return err
		}
		result = p
		return nil
	})
	return result, err
}

// DeletePlayer removes a player.
func (c *Commands) DeletePlayer(ctx context.Context, id int) error {
	return c.tx.WithTx(ctx, func(repos WriteRepos) error {
		return repos.Players.Delete(ctx, id)
	})
}

// AddPlayerToSeason assigns a player to a club season roster.
func (c *Commands) AddPlayerToSeason(ctx context.Context, playerID int, clubSeasonID int) (domain.PlayerSeason, error) {
	var result domain.PlayerSeason
	err := c.tx.WithTx(ctx, func(repos WriteRepos) error {
		ps, err := repos.PlayerSeasons.Create(ctx, playerID, clubSeasonID)
		if err != nil {
			return err
		}
		result = ps
		return nil
	})
	return result, err
}

// RemovePlayerFromSeason removes a player from a club season roster.
func (c *Commands) RemovePlayerFromSeason(ctx context.Context, playerSeasonID int) error {
	return c.tx.WithTx(ctx, func(repos WriteRepos) error {
		return repos.PlayerSeasons.Delete(ctx, playerSeasonID)
	})
}

// SetLineupEntry represents a single player assignment in a lineup.
type SetLineupEntry struct {
	PlayerSeasonID      int
	Position            string
	BackupPositions     *string
	InterchangePosition *string
}

// SetLineup upserts all player match entries for a club match (the weekly lineup).
func (c *Commands) SetLineup(ctx context.Context, clubMatchID int, entries []SetLineupEntry) ([]domain.PlayerMatch, error) {
	var result []domain.PlayerMatch
	err := c.tx.WithTx(ctx, func(repos WriteRepos) error {
		result = make([]domain.PlayerMatch, len(entries))
		for i, e := range entries {
			pos := domain.Position(e.Position)
			status := domain.PlayerMatchStatusNamed
			pm, err := repos.PlayerMatches.Upsert(ctx, domain.UpsertPlayerMatchParams{
				ClubMatchID:         clubMatchID,
				PlayerSeasonID:      e.PlayerSeasonID,
				Position:            &pos,
				Status:              &status,
				BackupPositions:     e.BackupPositions,
				InterchangePosition: e.InterchangePosition,
			})
			if err != nil {
				return err
			}
			result[i] = pm
		}
		return nil
	})
	return result, err
}

// CalculateFantasyScore calculates and stores the fantasy score for a player match
// based on AFL stats, then recalculates the club match total.
func (c *Commands) CalculateFantasyScore(ctx context.Context, playerMatchID int, stats domain.AFLStats) (domain.PlayerMatch, error) {
	var result domain.PlayerMatch
	err := c.tx.WithTx(ctx, func(repos WriteRepos) error {
		pm, err := repos.PlayerMatches.FindByID(ctx, playerMatchID)
		if err != nil {
			return err
		}

		score := pm.CalculateScore(stats)
		updated, err := repos.PlayerMatches.Upsert(ctx, domain.UpsertPlayerMatchParams{
			ClubMatchID:         pm.ClubMatchID,
			PlayerSeasonID:      pm.PlayerSeasonID,
			Position:            pm.Position,
			Status:              pm.Status,
			BackupPositions:     pm.BackupPositions,
			InterchangePosition: pm.InterchangePosition,
			Score:               &score,
		})
		if err != nil {
			return err
		}
		result = updated

		playerMatches, err := repos.PlayerMatches.FindByClubMatchID(ctx, pm.ClubMatchID)
		if err != nil {
			return err
		}
		clubMatch, err := repos.ClubMatches.FindByID(ctx, pm.ClubMatchID)
		if err != nil {
			return err
		}

		clubMatch.PlayerMatches = playerMatches
		return repos.ClubMatches.UpdateScore(ctx, pm.ClubMatchID, clubMatch.Score())
	})
	return result, err
}
