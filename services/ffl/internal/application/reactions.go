package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"xffl/contracts/events"
	"xffl/services/ffl/internal/domain"
)

// PlayerMatchUpdate carries AFL player performance data for a single player in a round.
type PlayerMatchUpdate struct {
	AFLPlayerMatchID  int
	AFLPlayerSeasonID int
	ClubMatchID       int
	RoundID           int
	Goals             int
	Kicks             int
	Handballs         int
	Marks             int
	Tackles           int
	Hitouts           int
}

// ProcessPlayerMatchUpdated finds all FFL player matches for the given AFL player in the
// matching round, links them to the AFL player match, and recalculates scores.
// Status is no longer synced here — it arrives via ProcessAFLMatchUpdated.
func (c *Commands) ProcessPlayerMatchUpdated(ctx context.Context, update PlayerMatchUpdate) error {
	slog.DebugContext(ctx, "ProcessPlayerMatchUpdated",
		slog.Int("afl_player_match_id", update.AFLPlayerMatchID),
		slog.Int("afl_player_season_id", update.AFLPlayerSeasonID),
		slog.Int("round_id", update.RoundID),
	)

	fflRound, err := c.rounds.FindByAFLRoundID(ctx, update.RoundID)
	if err != nil {
		return nil
	}

	fflPlayerSeasons, err := c.playerSeasons.FindByAFLPlayerSeasonID(ctx, update.AFLPlayerSeasonID)
	if err != nil {
		return fmt.Errorf("find FFL player seasons for AFL player_season %d: %w", update.AFLPlayerSeasonID, err)
	}
	if len(fflPlayerSeasons) == 0 {
		return nil
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
		pm, err := c.playerMatches.FindByPlayerSeasonAndRound(ctx, ps.ID, fflRound.ID)
		if err != nil {
			slog.DebugContext(ctx, "no player_match for player_season in round, skipping",
				slog.Int("player_season_id", ps.ID), slog.Int("round_id", fflRound.ID))
			continue
		}

		if pm.AFLPlayerMatchID == nil {
			if err := c.playerMatches.UpdateAFLPlayerMatchID(ctx, pm.ID, update.AFLPlayerMatchID); err != nil {
				slog.ErrorContext(ctx, "failed to set afl_player_match_id", slog.Int("player_match_id", pm.ID), slog.Any("error", err))
			}
		}

		scored, err := c.CalculateFantasyScore(ctx, pm.ID, stats)
		if err != nil {
			slog.ErrorContext(ctx, "failed to calculate score", slog.Int("player_match_id", pm.ID), slog.Any("error", err))
			continue
		}

		fflPayload, err := json.Marshal(events.FflPlayerMatchUpdatedPayload{
			PlayerMatchID: scored.ID,
			ClubMatchID:   scored.ClubMatchID,
			Score:         scored.Score,
		})
		if err != nil {
			slog.ErrorContext(ctx, "failed to marshal FflPlayerMatchUpdated", slog.Any("error", err))
			continue
		}
		if err := c.dispatcher.Publish(ctx, events.FflPlayerMatchUpdated, fflPayload); err != nil {
			slog.ErrorContext(ctx, "failed to publish FflPlayerMatchUpdated", slog.Any("error", err))
		}

		// Post-final stat correction cascade: if both axes are already final, recalculate ladder.
		if err := c.recalculateLadderIfBothFinal(ctx, scored.ClubMatchID); err != nil {
			slog.WarnContext(ctx, "ladder cascade failed", slog.Int("club_match_id", scored.ClubMatchID), slog.Any("error", err))
		}
	}

	return nil
}

// ProcessAFLMatchUpdated reacts to AFL.MatchUpdated: applies the PlayerSeasonIDStatusMap to
// matching FFL player_matches, recalculates scores, and emits FFL.ClubMatchScoreFinalized
// for club_matches where both axes are final.
func (c *Commands) ProcessAFLMatchUpdated(ctx context.Context, payload events.AflMatchUpdatedPayload) error {
	fflRound, err := c.rounds.FindByAFLRoundID(ctx, payload.RoundID)
	if err != nil {
		return nil // no FFL round linked to this AFL round
	}

	fflMatches, err := c.matches.FindByRoundID(ctx, fflRound.ID)
	if err != nil {
		return fmt.Errorf("load FFL matches for round %d: %w", fflRound.ID, err)
	}

	for _, m := range fflMatches {
		clubMatches, err := c.clubMatches.FindByMatchID(ctx, m.ID)
		if err != nil {
			slog.WarnContext(ctx, "load club_matches failed", slog.Int("match_id", m.ID), slog.Any("error", err))
			continue
		}
		for _, cm := range clubMatches {
			if err := c.applyAFLStatusMap(ctx, cm.ID, payload.PlayerSeasonIDStatusMap); err != nil {
				slog.WarnContext(ctx, "apply AFL status map failed", slog.Int("club_match_id", cm.ID), slog.Any("error", err))
			}

			if err := c.RecalculateScore(ctx, cm.ID); err != nil {
				slog.WarnContext(ctx, "recalculate score failed", slog.Int("club_match_id", cm.ID), slog.Any("error", err))
			}

			if cm.DataStatus == domain.ClubMatchDataFinal {
				allFinal, err := c.playerMatches.AllAFLStatusesFinal(ctx, cm.ID)
				if err == nil && allFinal {
					if err := c.emitClubMatchScoreFinalized(ctx, cm.ID, m.ID); err != nil {
						slog.WarnContext(ctx, "emit ClubMatchScoreFinalized failed", slog.Int("club_match_id", cm.ID), slog.Any("error", err))
					}
				}
			}
		}
	}
	return nil
}

// applyAFLStatusMap updates drv_afl_status for player_matches in a club_match whose
// afl_player_season_id is in the map. Players not in the map (other AFL matches) are unaffected.
func (c *Commands) applyAFLStatusMap(ctx context.Context, clubMatchID int, statusMap map[int]string) error {
	pms, err := c.playerMatches.FindByClubMatchID(ctx, clubMatchID)
	if err != nil {
		return fmt.Errorf("load player_matches: %w", err)
	}

	psIDs := make([]int, 0, len(pms))
	for _, pm := range pms {
		psIDs = append(psIDs, pm.PlayerSeasonID)
	}

	playerSeasons, err := c.playerSeasons.FindByIDs(ctx, psIDs)
	if err != nil {
		return fmt.Errorf("load player_seasons: %w", err)
	}

	// Build ffl_player_season_id → PlayerSeason map for lookup.
	psByID := make(map[int]domain.PlayerSeason, len(playerSeasons))
	for _, ps := range playerSeasons {
		psByID[ps.ID] = ps
	}

	// Build afl_player_season_id → player_match_id map.
	aflPSIDToPMID := make(map[int]int, len(pms))
	for _, pm := range pms {
		if ps, ok := psByID[pm.PlayerSeasonID]; ok && ps.AFLPlayerSeasonID != 0 {
			aflPSIDToPMID[ps.AFLPlayerSeasonID] = pm.ID
		}
	}

	for aflPSID, status := range statusMap {
		pmID, ok := aflPSIDToPMID[aflPSID]
		if !ok {
			continue // player not in this FFL club_match
		}
		if err := c.playerMatches.UpdateAFLStatus(ctx, pmID, domain.AFLStatus(status)); err != nil {
			slog.WarnContext(ctx, "UpdateAFLStatus failed",
				slog.Int("player_match_id", pmID), slog.String("status", status), slog.Any("error", err))
		}
	}
	return nil
}

// ProcessFflClubMatchUpdated reacts to FFL.ClubMatchUpdated: recalculates the club_match score
// and emits FFL.ClubMatchScoreFinalized if both axes are final.
func (c *Commands) ProcessFflClubMatchUpdated(ctx context.Context, clubMatchID, matchID int, dataStatus domain.ClubMatchDataStatus) error {
	if err := c.RecalculateScore(ctx, clubMatchID); err != nil {
		slog.WarnContext(ctx, "recalculate score failed", slog.Int("club_match_id", clubMatchID), slog.Any("error", err))
	}

	if dataStatus == domain.ClubMatchDataFinal {
		allFinal, err := c.playerMatches.AllAFLStatusesFinal(ctx, clubMatchID)
		if err == nil && allFinal {
			if err := c.emitClubMatchScoreFinalized(ctx, clubMatchID, matchID); err != nil {
				slog.WarnContext(ctx, "emit ClubMatchScoreFinalized failed", slog.Int("club_match_id", clubMatchID), slog.Any("error", err))
			}
		}
	}
	return nil
}

// ProcessFflClubMatchScoreFinalized reacts to FFL.ClubMatchScoreFinalized: if both club_matches
// for the match are finalized, emits FFL.MatchScoreFinalized.
func (c *Commands) ProcessFflClubMatchScoreFinalized(ctx context.Context, clubMatchID, matchID int) error {
	count, err := c.clubMatches.CountFinalByMatchID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("count final club_matches for match %d: %w", matchID, err)
	}
	if count < 2 {
		return nil
	}

	m, err := c.matches.FindByID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("load match %d: %w", matchID, err)
	}

	fflPayload, err := json.Marshal(events.FflMatchScoreFinalizedPayload{
		MatchID: matchID,
		RoundID: m.RoundID,
	})
	if err != nil {
		return fmt.Errorf("marshal FflMatchScoreFinalized: %w", err)
	}
	if err := c.dispatcher.Publish(ctx, events.FflMatchScoreFinalized, fflPayload); err != nil {
		slog.WarnContext(ctx, "publish FflMatchScoreFinalized failed", slog.Int("match_id", matchID), slog.Any("error", err))
	}
	return nil
}

// ProcessFflMatchScoreFinalized reacts to FFL.MatchScoreFinalized: derives and persists the
// match result, then recalculates the FFL ladder for the season.
func (c *Commands) ProcessFflMatchScoreFinalized(ctx context.Context, matchID, roundID int) error {
	clubMatches, err := c.clubMatches.FindByMatchID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("load club_matches for match %d: %w", matchID, err)
	}

	m, err := c.matches.FindByID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("load match %d: %w", matchID, err)
	}
	for _, cm := range clubMatches {
		if cm.ID == m.Home.ID {
			m.Home = cm
		} else if cm.ID == m.Away.ID {
			m.Away = cm
		}
	}

	if err := c.matches.UpdateResult(ctx, matchID, m.DeriveResult()); err != nil {
		slog.WarnContext(ctx, "update match result failed", slog.Int("match_id", matchID), slog.Any("error", err))
	}

	round, err := c.rounds.FindByID(ctx, roundID)
	if err != nil {
		return fmt.Errorf("load round %d: %w", roundID, err)
	}
	if err := c.RecalculateFflLadder(ctx, round.SeasonID); err != nil {
		slog.WarnContext(ctx, "recalculate FFL ladder failed", slog.Int("season_id", round.SeasonID), slog.Any("error", err))
	}
	return nil
}

// recalculateLadderIfBothFinal triggers FFL ladder recalculation when a post-final stat
// correction arrives (AFL.PlayerMatchUpdated after both axes are already final).
func (c *Commands) recalculateLadderIfBothFinal(ctx context.Context, clubMatchID int) error {
	cm, err := c.clubMatches.FindByID(ctx, clubMatchID)
	if err != nil {
		return nil
	}
	if cm.DataStatus != domain.ClubMatchDataFinal {
		return nil
	}
	allFinal, err := c.playerMatches.AllAFLStatusesFinal(ctx, clubMatchID)
	if err != nil || !allFinal {
		return nil
	}
	m, err := c.matches.FindByID(ctx, cm.MatchID)
	if err != nil {
		return err
	}
	r, err := c.rounds.FindByID(ctx, m.RoundID)
	if err != nil {
		return err
	}
	return c.RecalculateFflLadder(ctx, r.SeasonID)
}
