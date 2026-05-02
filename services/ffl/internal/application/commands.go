package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"xffl/contracts/events"
	"xffl/services/ffl/internal/domain"
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

// EventRepos provides read-only access for event handler lookups.
type EventRepos struct {
	Rounds        domain.RoundRepository
	PlayerSeasons domain.PlayerSeasonRepository
	PlayerMatches domain.PlayerMatchRepository
}

// Commands handles all write operations for the FFL service.
type Commands struct {
	tx           TxManager
	dispatcher   sharedevents.Dispatcher
	eventRepos   EventRepos
	playerLookup PlayerLookup
}

func NewCommands(tx TxManager, dispatcher sharedevents.Dispatcher, eventRepos EventRepos, playerLookup PlayerLookup) *Commands {
	return &Commands{tx: tx, dispatcher: dispatcher, eventRepos: eventRepos, playerLookup: playerLookup}
}

// CreatePlayer creates a new player linked to an AFL player.
func (c *Commands) CreatePlayer(ctx context.Context, name string, aflPlayerID int) (domain.Player, error) {
	var result domain.Player
	err := c.tx.WithTx(ctx, func(repos WriteRepos) error {
		p, err := repos.Players.Create(ctx, name, aflPlayerID)
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

// AddPlayerToSeason adds a player to a club season squad. The AFL player_season
// ID is the only cross-service handle the caller needs to provide; the FFL
// service resolves it to the underlying afl.player.id via Twirp and find-or-
// creates the ffl.player row. Player names are not written into ffl.player
// (drv_name is retired); they are read via federation traversal at query time.
func (c *Commands) AddPlayerToSeason(ctx context.Context, clubSeasonID, aflPlayerSeasonID int, fromRoundID, costCents *int) (domain.PlayerSeason, error) {
	aflPlayerID, err := c.playerLookup.LookupPlayerSeason(ctx, aflPlayerSeasonID)
	if err != nil {
		return domain.PlayerSeason{}, fmt.Errorf("lookup AFL player season: %w", err)
	}
	var result domain.PlayerSeason
	err = c.tx.WithTx(ctx, func(repos WriteRepos) error {
		player, err := repos.Players.FindByAFLPlayerID(ctx, aflPlayerID)
		if err != nil {
			// drv_name is set to "" — it's a denormalized field being retired
			// and will be dropped once UI traverses FFLPlayer.aflPlayer.name.
			player, err = repos.Players.Create(ctx, "", aflPlayerID)
			if err != nil {
				return err
			}
		}
		ps, err := repos.PlayerSeasons.Create(ctx, player.ID, clubSeasonID, fromRoundID, &aflPlayerSeasonID, costCents)
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

// SetTeamEntry represents a single player assignment in a team.
type SetTeamEntry struct {
	PlayerSeasonID      int
	Position            string
	BackupPositions     *string
	InterchangePosition *string
}

// SetTeam upserts all player match entries for a club match (the weekly team).
// Returns an error if the team violates team composition rules.
func (c *Commands) SetTeam(ctx context.Context, clubMatchID int, entries []SetTeamEntry) ([]domain.PlayerMatch, error) {
	// Validate composition rules before touching the database.
	params := make([]domain.UpsertPlayerMatchParams, len(entries))
	for i, e := range entries {
		pos := domain.Position(e.Position)
		params[i] = domain.UpsertPlayerMatchParams{
			Position:            &pos,
			BackupPositions:     e.BackupPositions,
			InterchangePosition: e.InterchangePosition,
		}
	}
	if err := domain.ValidateTeam(params); err != nil {
		return nil, err
	}

	var result []domain.PlayerMatch
	err := c.tx.WithTx(ctx, func(repos WriteRepos) error {
		// Replace the team: delete all existing entries first, then insert fresh.
		if err := repos.PlayerMatches.DeleteByClubMatchID(ctx, clubMatchID); err != nil {
			return err
		}
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

// HandlePlayerMatchUpdated processes an AFL.PlayerMatchUpdated event.
// It finds all FFL player matches for the given AFL player in the matching round
// and recalculates their fantasy scores.
func (c *Commands) HandlePlayerMatchUpdated(ctx context.Context, payload []byte) error {
	var event events.PlayerMatchUpdatedPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("unmarshal PlayerMatchUpdated: %w", err)
	}

	slog.DebugContext(ctx, "event received",
		slog.String("event_type", events.PlayerMatchUpdated),
		slog.Int("player_match_id", event.PlayerMatchID),
		slog.Int("player_season_id", event.PlayerSeasonID),
		slog.Int("round_id", event.RoundID),
	)

	// Find the FFL round that corresponds to this AFL round.
	fflRound, err := c.eventRepos.Rounds.FindByAFLRoundID(ctx, event.RoundID)
	if err != nil {
		return nil
	}

	// Find all FFL player seasons linked to this AFL player season.
	fflPlayerSeasons, err := c.eventRepos.PlayerSeasons.FindByAFLPlayerSeasonID(ctx, event.PlayerSeasonID)
	if err != nil {
		return fmt.Errorf("find FFL player seasons for AFL player_season %d: %w", event.PlayerSeasonID, err)
	}
	if len(fflPlayerSeasons) == 0 {
		return nil // player not in any FFL squad
	}

	stats := domain.AFLStats{
		Goals:     event.Goals,
		Kicks:     event.Kicks,
		Handballs: event.Handballs,
		Marks:     event.Marks,
		Tackles:   event.Tackles,
		Hitouts:   event.Hitouts,
	}

	for _, ps := range fflPlayerSeasons {
		// Find the FFL player match for this player season in the matching round.
		pm, err := c.eventRepos.PlayerMatches.FindByPlayerSeasonAndRound(ctx, ps.ID, fflRound.ID)
		if err != nil {
			slog.DebugContext(ctx, "no player_match for player_season in round, skipping", slog.Int("player_season_id", ps.ID), slog.Int("round_id", fflRound.ID))
			continue
		}

		// Link to the AFL player match if not already set.
		if pm.AFLPlayerMatchID == nil {
			if err := c.eventRepos.PlayerMatches.UpdateAFLPlayerMatchID(ctx, pm.ID, event.PlayerMatchID); err != nil {
				slog.ErrorContext(ctx, "failed to set afl_player_match_id on player_match", slog.Int("player_match_id", pm.ID), slog.Any("error", err))
			}
		}

		// Calculate and store the fantasy score.
		scored, err := c.CalculateFantasyScore(ctx, pm.ID, stats)
		if err != nil {
			slog.ErrorContext(ctx, "failed to calculate score for player_match", slog.Int("player_match_id", pm.ID), slog.Any("error", err))
			continue
		}

		// Publish FFL.FantasyScoreCalculated.
		fflPayload, err := json.Marshal(events.FantasyScoreCalculatedPayload{
			PlayerMatchID: scored.ID,
			Score:         scored.Score,
		})
		if err != nil {
			slog.ErrorContext(ctx, "failed to marshal FantasyScoreCalculated event", slog.Any("error", err))
			continue
		}
		if err := c.dispatcher.Publish(ctx, events.FantasyScoreCalculated, fflPayload); err != nil {
			slog.ErrorContext(ctx, "failed to publish FantasyScoreCalculated event", slog.Any("error", err))
		}
	}

	return nil
}
