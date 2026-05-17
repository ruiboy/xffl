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
	fflRound, err := c.rounds.FindByAFLRoundID(ctx, update.RoundID)
	if err != nil {
		return nil
	}

	// Find all FFL player seasons linked to this AFL player season.
	fflPlayerSeasons, err := c.playerSeasons.FindByAFLPlayerSeasonID(ctx, update.AFLPlayerSeasonID)
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
		pm, err := c.playerMatches.FindByPlayerSeasonAndRound(ctx, ps.ID, fflRound.ID)
		if err != nil {
			slog.DebugContext(ctx, "no player_match for player_season in round, skipping", slog.Int("player_season_id", ps.ID), slog.Int("round_id", fflRound.ID))
			continue
		}

		// Link to the AFL player match if not already set.
		if pm.AFLPlayerMatchID == nil {
			if err := c.playerMatches.UpdateAFLPlayerMatchID(ctx, pm.ID, update.AFLPlayerMatchID); err != nil {
				slog.ErrorContext(ctx, "failed to set afl_player_match_id on player_match", slog.Int("player_match_id", pm.ID), slog.Any("error", err))
			}
		}

		// Sync the AFL participation status onto the FFL record.
		if update.Status != "" {
			if err := c.playerMatches.UpdateAFLStatus(ctx, pm.ID, domain.AFLStatus(update.Status)); err != nil {
				slog.WarnContext(ctx, "failed to update drv_afl_status", slog.Int("player_match_id", pm.ID), slog.Any("error", err))
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

// ProcessAFLMatchFinalized reacts to AFL.MatchFinalized: for each FFL club_match in the
// corresponding round, sets drv_afl_status=dnp for unlinked players and emits
// FFL.ClubMatchScoreFinalized for club_matches already at data_status 'final'.
func (c *Commands) ProcessAFLMatchFinalized(ctx context.Context, aflRoundID int) error {
	fflRound, err := c.rounds.FindByAFLRoundID(ctx, aflRoundID)
	if err != nil {
		// No FFL round linked to this AFL round — nothing to do.
		return nil
	}

	fflMatches, err := c.matches.FindByRoundID(ctx, fflRound.ID)
	if err != nil {
		return fmt.Errorf("load FFL matches for round %d: %w", fflRound.ID, err)
	}

	for _, m := range fflMatches {
		clubMatches, err := c.clubMatches.FindByMatchID(ctx, m.ID)
		if err != nil {
			slog.WarnContext(ctx, "load club_matches for match failed", slog.Int("match_id", m.ID), slog.Any("error", err))
			continue
		}
		for _, cm := range clubMatches {
			if cm.DataStatus == domain.ClubMatchDataFinal {
				if err := c.emitClubMatchScoreFinalized(ctx, cm.ID, m.ID); err != nil {
					slog.WarnContext(ctx, "emit ClubMatchScoreFinalized failed", slog.Int("club_match_id", cm.ID), slog.Any("error", err))
				}
			}
			// Set drv_afl_status=dnp for any player_match whose drv_afl_status is still NULL
			// (no AFL stats row exists for them → they did not play).
			if err := c.playerMatches.SetAFLStatusDNP(ctx, cm.ID); err != nil {
				slog.WarnContext(ctx, "set drv_afl_status dnp failed", slog.Int("club_match_id", cm.ID), slog.Any("error", err))
			}
		}
	}
	return nil
}

// ProcessFflTeamFinalized reacts to FFL.TeamFinalized: recalculates the club_match score and
// emits FFL.ClubMatchScoreFinalized. Assumes AFL stats are final by the time this fires (the
// data ops workflow enforces this order; RecalculateFflLadder is the safety net).
func (c *Commands) ProcessFflTeamFinalized(ctx context.Context, clubMatchID, matchID int) error {
	if err := c.emitClubMatchScoreFinalized(ctx, clubMatchID, matchID); err != nil {
		slog.WarnContext(ctx, "emit ClubMatchScoreFinalized failed", slog.Int("club_match_id", clubMatchID), slog.Any("error", err))
	}
	return nil
}

// ProcessFflClubMatchScoreFinalized reacts to FFL.ClubMatchScoreFinalized: if both club_matches
// for the match are now final, emits FFL.MatchFinalized.
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

	fflPayload, err := json.Marshal(events.FflMatchFinalizedPayload{
		MatchID: matchID,
		RoundID: m.RoundID,
	})
	if err != nil {
		return fmt.Errorf("marshal FflMatchFinalized: %w", err)
	}
	if err := c.dispatcher.Publish(ctx, events.FflMatchFinalized, fflPayload); err != nil {
		slog.WarnContext(ctx, "publish FflMatchFinalized failed", slog.Int("match_id", matchID), slog.Any("error", err))
	}
	return nil
}

// ProcessFflMatchFinalized reacts to FFL.MatchFinalized: derives and persists the match result,
// then recalculates the FFL ladder for the season.
func (c *Commands) ProcessFflMatchFinalized(ctx context.Context, matchID, roundID int) error {
	// Load the match with its stored club_match scores to derive the result.
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
