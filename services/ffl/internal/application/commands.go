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
	Matches       domain.MatchRepository
	ClubMatches   domain.ClubMatchRepository
}

// TeamSubmitter is the narrow interface DataOpsCommands uses to delegate team persistence.
type TeamSubmitter interface {
	SetTeam(ctx context.Context, params SetTeamParams) ([]domain.PlayerMatch, error)
}

// SetTeamParams are the inputs to SetTeam.
type SetTeamParams struct {
	ClubMatchID int
	Entries     []SetTeamEntry
}

// CommandsDeps bundles the dependencies Commands needs beyond the transaction
// manager and event dispatcher. Grouping them in a struct keeps callsites
// labelled and stable as new dependencies are added.
type CommandsDeps struct {
	EventRepos   EventRepos
	PlayerLookup PlayerLookup
}

// Commands handles all write operations for the FFL service.
type Commands struct {
	tx           TxManager
	dispatcher   sharedevents.Dispatcher
	eventRepos   EventRepos
	playerLookup PlayerLookup
}

func NewCommands(tx TxManager, dispatcher sharedevents.Dispatcher, deps CommandsDeps) *Commands {
	return &Commands{tx: tx, dispatcher: dispatcher, eventRepos: deps.EventRepos, playerLookup: deps.PlayerLookup}
}

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
	Score               *int // optional seed score for new players (AFL events are authoritative once set)
}

// SetTeam persists a complete team for a club match using diff-based persistence to
// preserve afl_player_match_id links for returning players. It validates team composition
// via the domain, computes a provisional score, updates data_status, and publishes
// FFL.TeamSubmitted.
func (c *Commands) SetTeam(ctx context.Context, params SetTeamParams) ([]domain.PlayerMatch, error) {
	var result []domain.PlayerMatch
	var matchID int

	err := c.tx.WithTx(ctx, func(repos WriteRepos) error {
		// load the ClubMatch
		cm, err := repos.ClubMatches.FindByID(ctx, params.ClubMatchID)
		if err != nil {
			return fmt.Errorf("find club match: %w", err)
		}
		matchID = cm.MatchID

		// load the PlayerMatches - key by PlayerSeasonID to ease lookup when updating with
		// the incoming changes in the next step
		existing, err := repos.PlayerMatches.FindByClubMatchID(ctx, params.ClubMatchID)
		if err != nil {
			return fmt.Errorf("find existing player matches: %w", err)
		}
		existingByPS := make(map[int]domain.PlayerMatch, len(existing))
		for _, pm := range existing {
			existingByPS[pm.PlayerSeasonID] = pm
		}

		// build the new list of PlayerMatches, copy over existing data - ID, AFL Link etc -  for
		// any PMs that already exist
		newPlayers := make([]domain.PlayerMatch, 0, len(params.Entries))
		inNewTeam := make(map[int]bool)
		for _, e := range params.Entries {
			if e.PlayerSeasonID == 0 {
				continue
			}
			inNewTeam[e.PlayerSeasonID] = true
			newPlayers = append(newPlayers, entryToPlayerMatch(e, params.ClubMatchID, existingByPS))
		}

		// validate and submit the team
		if _, err := cm.SubmitTeam(newPlayers); err != nil {
			return err
		}

		// do diff-based persistence: delete any PlayerMatches no longer needed, and upsert the rest
		for _, pm := range existing {
			if !inNewTeam[pm.PlayerSeasonID] {
				if err := repos.PlayerMatches.DeleteByID(ctx, pm.ID); err != nil {
					return fmt.Errorf("delete removed player_match %d: %w", pm.ID, err)
				}
			}
		}

		result = make([]domain.PlayerMatch, 0, len(cm.PlayerMatches))
		for _, pm := range cm.PlayerMatches {
			upserted, err := repos.PlayerMatches.Upsert(ctx, upsertParamsFromPlayerMatch(pm))
			if err != nil {
				return fmt.Errorf("upsert player_match for player_season %d: %w", pm.PlayerSeasonID, err)
			}
			result = append(result, upserted)
		}

		// compute the provisional score and set status
		cm.PlayerMatches = result
		if err := repos.ClubMatches.UpdateScore(ctx, cm.ID, cm.Score()); err != nil {
			return fmt.Errorf("update club match score: %w", err)
		}
		if err := repos.ClubMatches.UpdateDataStatus(ctx, cm.ID, cm.DataStatus); err != nil {
			return fmt.Errorf("update club match data status: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Recalculate scores now that the team is persisted and AFL stats may already be available.
	if err := c.RecalculateClubMatchScore(ctx, params.ClubMatchID); err != nil {
		slog.WarnContext(ctx, "recalculate club match score failed after SetTeam", slog.Int("club_match_id", params.ClubMatchID), slog.Any("error", err))
	}

	// Publish integration event for external subscribers.
	match, err := c.eventRepos.Matches.FindByID(ctx, matchID)
	if err != nil {
		slog.WarnContext(ctx, "failed to load match for FflTeamSubmitted event", slog.Int("match_id", matchID), slog.Any("error", err))
		return result, nil
	}
	b, err := json.Marshal(events.FflTeamSubmittedPayload{
		ClubMatchID: params.ClubMatchID,
		MatchID:     matchID,
		RoundID:     match.RoundID,
	})
	if err == nil {
		if err := c.dispatcher.Publish(ctx, events.FflTeamSubmitted, b); err != nil {
			slog.WarnContext(ctx, "publish FflTeamSubmitted failed", slog.Int("club_match_id", params.ClubMatchID), slog.Any("error", err))
		}
	}
	return result, nil
}

// RecalculateClubMatchScore re-applies AFL stats to all player_matches for a club_match,
// then re-sums the club_match total via ClubMatch.Score().
//
// Two lookup paths are used:
//   - Linked: player_matches that already have afl_player_match_id → looked up by that ID.
//   - Unlinked: freshly submitted rows with no afl_player_match_id yet → looked up by
//     (afl_player_season_id, afl_round_id) and the link is established as a side effect.
func (c *Commands) RecalculateClubMatchScore(ctx context.Context, clubMatchID int) error {
	pms, err := c.eventRepos.PlayerMatches.FindByClubMatchID(ctx, clubMatchID)
	if err != nil {
		return fmt.Errorf("load player matches for club_match %d: %w", clubMatchID, err)
	}

	// Partition into linked (have AFL player_match_id) and unlinked.
	var aflMatchIDs []int
	var unlinkedPSIDs []int
	for _, pm := range pms {
		if pm.AFLPlayerMatchID != nil {
			aflMatchIDs = append(aflMatchIDs, *pm.AFLPlayerMatchID)
		} else {
			unlinkedPSIDs = append(unlinkedPSIDs, pm.PlayerSeasonID)
		}
	}

	// Network call 1: fetch stats for linked player_matches by AFL player_match_id.
	statsByAFLMatchID := make(map[int]PlayerMatchStats)
	if len(aflMatchIDs) > 0 {
		linked, err := c.playerLookup.LookupPlayerMatch(ctx, aflMatchIDs)
		if err != nil {
			return fmt.Errorf("lookup player match stats: %w", err)
		}
		for _, s := range linked {
			statsByAFLMatchID[s.ID] = s
		}
	}

	// Network call 2: fetch stats for unlinked player_matches by (AFL player_season_id, AFL round_id).
	// statsByAFLSeasonID maps AFL player_season_id → stats (includes the AFL player_match_id for linking).
	statsByAFLSeasonID := make(map[int]PlayerMatchStats)
	if len(unlinkedPSIDs) > 0 {
		// Resolve AFL player_season_ids from ffl.player_season records.
		playerSeasons, err := c.eventRepos.PlayerSeasons.FindByIDs(ctx, unlinkedPSIDs)
		if err != nil {
			return fmt.Errorf("load player_seasons for unlinked player_matches: %w", err)
		}
		var aflPSIDs []int
		for _, ps := range playerSeasons {
			if ps.AFLPlayerSeasonID != nil {
				aflPSIDs = append(aflPSIDs, *ps.AFLPlayerSeasonID)
			}
		}

		if len(aflPSIDs) > 0 {
			// Traverse clubMatchID → match → round to get the AFL round ID.
			cm, err := c.eventRepos.ClubMatches.FindByID(ctx, clubMatchID)
			if err != nil {
				return fmt.Errorf("load club_match %d: %w", clubMatchID, err)
			}
			m, err := c.eventRepos.Matches.FindByID(ctx, cm.MatchID)
			if err != nil {
				return fmt.Errorf("load match %d: %w", cm.MatchID, err)
			}
			r, err := c.eventRepos.Rounds.FindByID(ctx, m.RoundID)
			if err != nil {
				return fmt.Errorf("load round %d: %w", m.RoundID, err)
			}

			if r.AFLRoundID != nil {
				unlinked, err := c.playerLookup.LookupPlayerMatchBySeasonRound(ctx, aflPSIDs, *r.AFLRoundID)
				if err != nil {
					return fmt.Errorf("lookup player match stats by season/round: %w", err)
				}
				for _, s := range unlinked {
					statsByAFLSeasonID[s.PlayerSeasonID] = s
				}
			}
		}
	}

	// Apply stats and re-sum inside a transaction.
	return c.tx.WithTx(ctx, func(repos WriteRepos) error {
		pms, err := repos.PlayerMatches.FindByClubMatchID(ctx, clubMatchID)
		if err != nil {
			return err
		}

		for _, pm := range pms {
			var s PlayerMatchStats
			var found bool

			if pm.AFLPlayerMatchID != nil {
				s, found = statsByAFLMatchID[*pm.AFLPlayerMatchID]
			} else {
				// Look up by AFL player_season_id via the unlinked path.
				ps, err := repos.PlayerSeasons.FindByID(ctx, pm.PlayerSeasonID)
				if err == nil && ps.AFLPlayerSeasonID != nil {
					s, found = statsByAFLSeasonID[*ps.AFLPlayerSeasonID]
					if found && s.ID != 0 {
						// Establish the AFL player_match link for future calls.
						if linkErr := repos.PlayerMatches.UpdateAFLPlayerMatchID(ctx, pm.ID, s.ID); linkErr != nil {
							slog.WarnContext(ctx, "update afl_player_match_id failed", slog.Int("player_match_id", pm.ID), slog.Any("error", linkErr))
						}
					}
				}
			}

			if !found {
				continue
			}

			score := pm.CalculateScore(domain.AFLStats{
				Goals:     s.Goals,
				Kicks:     s.Kicks,
				Handballs: s.Handballs,
				Marks:     s.Marks,
				Tackles:   s.Tackles,
				Hitouts:   s.Hitouts,
			})
			status := domain.PlayerMatchStatus(s.Status)
			if _, err := repos.PlayerMatches.Upsert(ctx, domain.UpsertPlayerMatchParams{
				ClubMatchID:         pm.ClubMatchID,
				PlayerSeasonID:      pm.PlayerSeasonID,
				Position:            pm.Position,
				Status:              &status,
				BackupPositions:     pm.BackupPositions,
				InterchangePosition: pm.InterchangePosition,
				Score:               &score,
			}); err != nil {
				return fmt.Errorf("upsert player_match %d: %w", pm.ID, err)
			}
		}

		// Re-load updated player_matches to compute the new club total.
		updated, err := repos.PlayerMatches.FindByClubMatchID(ctx, clubMatchID)
		if err != nil {
			return err
		}
		cm, err := repos.ClubMatches.FindByID(ctx, clubMatchID)
		if err != nil {
			return err
		}
		cm.PlayerMatches = updated
		return repos.ClubMatches.UpdateScore(ctx, clubMatchID, cm.Score())
	})
}

// entryToPlayerMatch converts a SetTeamEntry to a domain.PlayerMatch, enriching it
// with the ID, score, and AFL link from the existing DB record for returning players.
func entryToPlayerMatch(e SetTeamEntry, clubMatchID int, existing map[int]domain.PlayerMatch) domain.PlayerMatch {
	status := domain.PlayerMatchStatusNamed
	pm := domain.PlayerMatch{
		ClubMatchID:    clubMatchID,
		PlayerSeasonID: e.PlayerSeasonID,
		Status:         &status,
	}
	if e.BackupPositions != nil || e.InterchangePosition != nil {
		pm.BackupPositions = e.BackupPositions
		pm.InterchangePosition = e.InterchangePosition
	} else {
		pos := domain.Position(e.Position)
		pm.Position = &pos
	}
	if ex, ok := existing[e.PlayerSeasonID]; ok {
		pm.ID = ex.ID
		pm.Score = ex.Score
		pm.AFLPlayerMatchID = ex.AFLPlayerMatchID
	} else if e.Score != nil {
		pm.Score = *e.Score
	}
	return pm
}

// upsertParamsFromPlayerMatch converts a domain.PlayerMatch to UpsertPlayerMatchParams.
// Score is passed as nil when zero so COALESCE in the upsert preserves any existing DB value.
func upsertParamsFromPlayerMatch(pm domain.PlayerMatch) domain.UpsertPlayerMatchParams {
	params := domain.UpsertPlayerMatchParams{
		ClubMatchID:         pm.ClubMatchID,
		PlayerSeasonID:      pm.PlayerSeasonID,
		Position:            pm.Position,
		Status:              pm.Status,
		BackupPositions:     pm.BackupPositions,
		InterchangePosition: pm.InterchangePosition,
	}
	if pm.Score != 0 {
		s := pm.Score
		params.Score = &s
	}
	return params
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

// PlayerMatchUpdate carries AFL player performance data for a single player in a round.
type PlayerMatchUpdate struct {
	AFLPlayerMatchID  int
	AFLPlayerSeasonID int
	ClubMatchID       int
	RoundID           int
	Status            string
	Goals             int
	Kicks             int
	Handballs         int
	Marks             int
	Tackles           int
	Hitouts           int
}

// ProcessPlayerMatchUpdated finds all FFL player matches for the given AFL player in the
// matching round, links them to the AFL player match, syncs status, and recalculates scores.
func (c *Commands) ProcessPlayerMatchUpdated(ctx context.Context, update PlayerMatchUpdate) error {
	slog.DebugContext(ctx, "ProcessPlayerMatchUpdated",
		slog.Int("afl_player_match_id", update.AFLPlayerMatchID),
		slog.Int("afl_player_season_id", update.AFLPlayerSeasonID),
		slog.Int("round_id", update.RoundID),
	)

	// Find the FFL round that corresponds to this AFL round.
	fflRound, err := c.eventRepos.Rounds.FindByAFLRoundID(ctx, update.RoundID)
	if err != nil {
		return nil
	}

	// Find all FFL player seasons linked to this AFL player season.
	fflPlayerSeasons, err := c.eventRepos.PlayerSeasons.FindByAFLPlayerSeasonID(ctx, update.AFLPlayerSeasonID)
	if err != nil {
		return fmt.Errorf("find FFL player seasons for AFL player_season %d: %w", update.AFLPlayerSeasonID, err)
	}
	if len(fflPlayerSeasons) == 0 {
		return nil // player not in any FFL squad
	}

	stats := domain.AFLStats{
		Goals:     update.Goals,
		Kicks:     update.Kicks,
		Handballs: update.Handballs,
		Marks:     update.Marks,
		Tackles:   update.Tackles,
		Hitouts:   update.Hitouts,
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
			if err := c.eventRepos.PlayerMatches.UpdateAFLPlayerMatchID(ctx, pm.ID, update.AFLPlayerMatchID); err != nil {
				slog.ErrorContext(ctx, "failed to set afl_player_match_id on player_match", slog.Int("player_match_id", pm.ID), slog.Any("error", err))
			}
		}

		// Sync the AFL player match status onto the FFL record (named during partial import).
		if update.Status != "" {
			if err := c.eventRepos.PlayerMatches.UpdateStatus(ctx, pm.ID, domain.PlayerMatchStatus(update.Status)); err != nil {
				slog.WarnContext(ctx, "failed to update player_match status", slog.Int("player_match_id", pm.ID), slog.Any("error", err))
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
