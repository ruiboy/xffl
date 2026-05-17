package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"xffl/contracts/events"
	"xffl/services/ffl/internal/domain"
)

// RecalculateScore re-applies AFL stats to all player_matches for a club_match,
// then re-sums the club_match total via ClubMatch.Score().
//
// Two lookup paths are used:
//   - Linked: player_matches that already have afl_player_match_id → looked up by that ID.
//   - Unlinked: freshly submitted rows with no afl_player_match_id yet → looked up by
//     (afl_player_season_id, afl_round_id) and the link is established as a side effect.
func (c *Commands) RecalculateScore(ctx context.Context, clubMatchID int) error {
	pms, err := c.playerMatches.FindByClubMatchID(ctx, clubMatchID)
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
		playerSeasons, err := c.playerSeasons.FindByIDs(ctx, unlinkedPSIDs)
		if err != nil {
			return fmt.Errorf("load player_seasons for unlinked player_matches: %w", err)
		}
		var aflPSIDs []int
		for _, ps := range playerSeasons {
			if ps.AFLPlayerSeasonID != 0 {
				aflPSIDs = append(aflPSIDs, ps.AFLPlayerSeasonID)
			}
		}

		if len(aflPSIDs) > 0 {
			// Traverse clubMatchID → match → round to get the AFL round ID.
			cm, err := c.clubMatches.FindByID(ctx, clubMatchID)
			if err != nil {
				return fmt.Errorf("load club_match %d: %w", clubMatchID, err)
			}
			m, err := c.matches.FindByID(ctx, cm.MatchID)
			if err != nil {
				return fmt.Errorf("load match %d: %w", cm.MatchID, err)
			}
			r, err := c.rounds.FindByID(ctx, m.RoundID)
			if err != nil {
				return fmt.Errorf("load round %d: %w", m.RoundID, err)
			}

			if r.AFLRoundID != 0 {
				unlinked, err := c.playerLookup.LookupPlayerMatchBySeasonRound(ctx, aflPSIDs, r.AFLRoundID)
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
				if err == nil && ps.AFLPlayerSeasonID != 0 {
					s, found = statsByAFLSeasonID[ps.AFLPlayerSeasonID]
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
			upsertParams := domain.UpsertPlayerMatchParams{
				ClubMatchID:         pm.ClubMatchID,
				PlayerSeasonID:      pm.PlayerSeasonID,
				Position:            pm.Position,
				BackupPositions:     pm.BackupPositions,
				InterchangePosition: pm.InterchangePosition,
				Score:               &score,
			}
			if s.Status != "" {
				aflStatus := domain.AFLStatus(s.Status)
				upsertParams.AFLStatus = &aflStatus
			}
			if _, err := repos.PlayerMatches.Upsert(ctx, upsertParams); err != nil {
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

// RecalculateFflLadder rebuilds FFL ladder standings for the given season from all final matches.
// Idempotent — safe to call multiple times.
func (c *Commands) RecalculateFflLadder(ctx context.Context, seasonID int) error {
	matches, err := c.matches.FindFinalBySeasonID(ctx, seasonID)
	if err != nil {
		return fmt.Errorf("load final FFL matches: %w", err)
	}
	for _, cs := range domain.CalculateLadder(matches) {
		if err := c.clubSeasons.Update(ctx, cs); err != nil {
			slog.WarnContext(ctx, "update club season failed",
				slog.Int("club_season_id", cs.ID), slog.Any("error", err))
		}
	}
	return nil
}

// AllAFLStatusesFinal returns true when every player_match in the club_match has
// drv_afl_status ∈ {played, dnp} — none are null or playing.
func (c *Commands) AllAFLStatusesFinal(ctx context.Context, clubMatchID int) (bool, error) {
	return c.playerMatches.AllAFLStatusesFinal(ctx, clubMatchID)
}

func (c *Commands) emitClubMatchScoreFinalized(ctx context.Context, clubMatchID, matchID int) error {
	b, err := json.Marshal(events.FflClubMatchScoreFinalizedPayload{
		ClubMatchID: clubMatchID,
		MatchID:     matchID,
	})
	if err != nil {
		return err
	}
	return c.dispatcher.Publish(ctx, events.FflClubMatchScoreFinalized, b)
}
