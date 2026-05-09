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
	matches     domain.MatchRepository
	clubMatches domain.ClubMatchRepository
	clubSeasons domain.ClubSeasonRepository
	rounds      domain.RoundRepository
	dispatcher  sharedevents.Dispatcher
}

func NewScoreCommands(
	matches domain.MatchRepository,
	clubMatches domain.ClubMatchRepository,
	clubSeasons domain.ClubSeasonRepository,
	rounds domain.RoundRepository,
	dispatcher sharedevents.Dispatcher,
) *ScoreCommands {
	return &ScoreCommands{
		matches:     matches,
		clubMatches: clubMatches,
		clubSeasons: clubSeasons,
		rounds:      rounds,
		dispatcher:  dispatcher,
	}
}

// HandleAflMatchFinalized reacts to AFL.MatchFinalized: for each FFL club_match in the
// corresponding round, recalculates provisional scores. For those already at data_status
// 'final', emits FFL.ClubMatchScoreFinalized.
func (c *ScoreCommands) HandleAflMatchFinalized(ctx context.Context, payload []byte) error {
	var p events.AflMatchFinalizedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal AflMatchFinalized: %w", err)
	}

	fflRound, err := c.rounds.FindByAFLRoundID(ctx, p.RoundID)
	if err != nil {
		// No FFL round linked to this AFL round — nothing to do.
		return nil
	}

	fflMatches, err := c.matches.FindByRoundID(ctx, fflRound.ID)
	if err != nil {
		return fmt.Errorf("load FFL matches for round %d: %w", fflRound.ID, err)
	}

	for _, m := range fflMatches {
		for _, cmID := range []int{m.Home.ID, m.Away.ID} {
			if cmID == 0 {
				continue
			}
			cm, err := c.clubMatches.FindByID(ctx, cmID)
			if err != nil {
				slog.WarnContext(ctx, "load club_match failed", slog.Int("id", cmID), slog.Any("error", err))
				continue
			}
			if cm.DataStatus == domain.ClubMatchDataFinal {
				if err := c.emitClubMatchScoreFinalized(ctx, cm.ID, m.ID); err != nil {
					slog.WarnContext(ctx, "emit ClubMatchScoreFinalized failed", slog.Int("club_match_id", cm.ID), slog.Any("error", err))
				}
			}
		}
	}
	return nil
}

// HandleFflTeamSubmitted reacts to FFL.TeamSubmitted. Currently a no-op at the score level
// because player scores are already updated via AFL.PlayerMatchUpdated → CalculateFantasyScore.
func (c *ScoreCommands) HandleFflTeamSubmitted(ctx context.Context, payload []byte) error {
	return nil
}

// HandleFflTeamFinalized reacts to FFL.TeamFinalized: recalculates the club_match score and
// emits FFL.ClubMatchScoreFinalized. Assumes AFL stats are final by the time this fires (the
// data ops workflow enforces this order; RecalculateFflLadder is the safety net).
func (c *ScoreCommands) HandleFflTeamFinalized(ctx context.Context, payload []byte) error {
	var p events.FflTeamFinalizedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal FflTeamFinalized: %w", err)
	}

	if err := c.emitClubMatchScoreFinalized(ctx, p.ClubMatchID, p.MatchID); err != nil {
		slog.WarnContext(ctx, "emit ClubMatchScoreFinalized failed", slog.Int("club_match_id", p.ClubMatchID), slog.Any("error", err))
	}
	return nil
}

// HandleFflClubMatchScoreFinalized reacts to FFL.ClubMatchScoreFinalized: if both club_matches
// for the match are now final, emits FFL.MatchFinalized.
func (c *ScoreCommands) HandleFflClubMatchScoreFinalized(ctx context.Context, payload []byte) error {
	var p events.FflClubMatchScoreFinalizedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal FflClubMatchScoreFinalized: %w", err)
	}

	count, err := c.clubMatches.CountFinalByMatchID(ctx, p.MatchID)
	if err != nil {
		return fmt.Errorf("count final club_matches for match %d: %w", p.MatchID, err)
	}
	if count < 2 {
		return nil
	}

	m, err := c.matches.FindByID(ctx, p.MatchID)
	if err != nil {
		return fmt.Errorf("load match %d: %w", p.MatchID, err)
	}

	fflPayload, err := json.Marshal(events.FflMatchFinalizedPayload{
		MatchID: p.MatchID,
		RoundID: m.RoundID,
	})
	if err != nil {
		return fmt.Errorf("marshal FflMatchFinalized: %w", err)
	}
	if err := c.dispatcher.Publish(ctx, events.FflMatchFinalized, fflPayload); err != nil {
		slog.WarnContext(ctx, "publish FflMatchFinalized failed", slog.Int("match_id", p.MatchID), slog.Any("error", err))
	}
	return nil
}

// HandleFflMatchFinalized reacts to FFL.MatchFinalized: derives and persists the match result,
// then recalculates the FFL ladder for the season.
func (c *ScoreCommands) HandleFflMatchFinalized(ctx context.Context, payload []byte) error {
	var p events.FflMatchFinalizedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal FflMatchFinalized: %w", err)
	}

	// Load the match with its stored club_match scores to derive the result.
	clubMatches, err := c.clubMatches.FindByMatchID(ctx, p.MatchID)
	if err != nil {
		return fmt.Errorf("load club_matches for match %d: %w", p.MatchID, err)
	}

	m, err := c.matches.FindByID(ctx, p.MatchID)
	if err != nil {
		return fmt.Errorf("load match %d: %w", p.MatchID, err)
	}
	for _, cm := range clubMatches {
		if cm.ID == m.Home.ID {
			m.Home = cm
		} else if cm.ID == m.Away.ID {
			m.Away = cm
		}
	}

	if err := c.matches.UpdateResult(ctx, p.MatchID, m.DeriveResult()); err != nil {
		slog.WarnContext(ctx, "update match result failed", slog.Int("match_id", p.MatchID), slog.Any("error", err))
	}

	round, err := c.rounds.FindByID(ctx, m.RoundID)
	if err != nil {
		return fmt.Errorf("load round %d: %w", m.RoundID, err)
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
