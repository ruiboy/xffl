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

// ScoreCommands handles FFL score and ladder calculation use cases.
type ScoreCommands struct {
	matches       domain.MatchRepository
	clubMatches   domain.ClubMatchRepository
	clubSeasons   domain.ClubSeasonRepository
	rounds        domain.RoundRepository
	playerMatches domain.PlayerMatchRepository
	dispatcher    sharedevents.Dispatcher
}

func NewScoreCommands(
	matches domain.MatchRepository,
	clubMatches domain.ClubMatchRepository,
	clubSeasons domain.ClubSeasonRepository,
	rounds domain.RoundRepository,
	playerMatches domain.PlayerMatchRepository,
	dispatcher sharedevents.Dispatcher,
) *ScoreCommands {
	return &ScoreCommands{
		matches:       matches,
		clubMatches:   clubMatches,
		clubSeasons:   clubSeasons,
		rounds:        rounds,
		playerMatches: playerMatches,
		dispatcher:    dispatcher,
	}
}

// ProcessAFLRoundFinalized reacts to AFL.MatchFinalized: for each FFL club_match in the
// corresponding round, recalculates provisional scores. For those already at data_status
// 'final', emits FFL.ClubMatchScoreFinalized.
func (c *ScoreCommands) ProcessAFLRoundFinalized(ctx context.Context, aflRoundID int) error {
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
			c.inferPlayerMatchStatuses(ctx, cm.ID)
		}
	}
	return nil
}

// inferPlayerMatchStatuses sets ffl.player_match.status based on AFL link presence:
// - linked (afl_player_match_id set) → played
// - unlinked and named (no AFL stats found for this player) → dnp
func (c *ScoreCommands) inferPlayerMatchStatuses(ctx context.Context, clubMatchID int) {
	pms, err := c.playerMatches.FindByClubMatchID(ctx, clubMatchID)
	if err != nil {
		slog.WarnContext(ctx, "load player_matches for status inference failed", slog.Int("club_match_id", clubMatchID), slog.Any("error", err))
		return
	}
	for _, pm := range pms {
		var newStatus domain.PlayerMatchStatus
		if pm.AFLPlayerMatchID != nil {
			newStatus = domain.PlayerMatchStatusPlayed
		} else if pm.Status != nil && *pm.Status == domain.PlayerMatchStatusNamed {
			newStatus = domain.PlayerMatchStatusDNP
		} else {
			continue
		}
		if err := c.playerMatches.UpdateStatus(ctx, pm.ID, newStatus); err != nil {
			slog.WarnContext(ctx, "update player_match status failed", slog.Int("player_match_id", pm.ID), slog.Any("error", err))
		}
	}
}

// ProcessFflTeamFinalized reacts to FFL.TeamFinalized: recalculates the club_match score and
// emits FFL.ClubMatchScoreFinalized. Assumes AFL stats are final by the time this fires (the
// data ops workflow enforces this order; RecalculateFflLadder is the safety net).
func (c *ScoreCommands) ProcessFflTeamFinalized(ctx context.Context, clubMatchID, matchID int) error {
	if err := c.emitClubMatchScoreFinalized(ctx, clubMatchID, matchID); err != nil {
		slog.WarnContext(ctx, "emit ClubMatchScoreFinalized failed", slog.Int("club_match_id", clubMatchID), slog.Any("error", err))
	}
	return nil
}

// ProcessFflClubMatchScoreFinalized reacts to FFL.ClubMatchScoreFinalized: if both club_matches
// for the match are now final, emits FFL.MatchFinalized.
func (c *ScoreCommands) ProcessFflClubMatchScoreFinalized(ctx context.Context, clubMatchID, matchID int) error {
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
func (c *ScoreCommands) ProcessFflMatchFinalized(ctx context.Context, matchID, roundID int) error {
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

// RecalculateFflLadder rebuilds FFL ladder standings for the given season from all final matches.
// Idempotent — safe to call multiple times.
func (c *ScoreCommands) RecalculateFflLadder(ctx context.Context, seasonID int) error {
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

// emitClubMatchScoreFinalized publishes FFL.ClubMatchScoreFinalized for a given club_match.
func (c *ScoreCommands) emitClubMatchScoreFinalized(ctx context.Context, clubMatchID, matchID int) error {
	b, err := json.Marshal(events.FflClubMatchScoreFinalizedPayload{
		ClubMatchID: clubMatchID,
		MatchID:     matchID,
	})
	if err != nil {
		return err
	}
	return c.dispatcher.Publish(ctx, events.FflClubMatchScoreFinalized, b)
}
