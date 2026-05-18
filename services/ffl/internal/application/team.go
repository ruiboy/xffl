package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"xffl/contracts/events"
	"xffl/services/ffl/internal/domain"
)

// SetTeamParams are the inputs to SetTeam.
type SetTeamParams struct {
	ClubMatchID int
	Entries     []SetTeamEntry
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
	if err := c.RecalculateScore(ctx, params.ClubMatchID); err != nil {
		slog.WarnContext(ctx, "recalculate club match score failed after SetTeam", slog.Int("club_match_id", params.ClubMatchID), slog.Any("error", err))
	}

	// Reload player_matches after score recalculation so the snapshot is current.
	latest, err := c.playerMatches.FindByClubMatchID(ctx, params.ClubMatchID)
	if err != nil {
		slog.WarnContext(ctx, "reload player_matches failed after SetTeam", slog.Int("club_match_id", params.ClubMatchID), slog.Any("error", err))
		latest = result
	}

	match, err := c.matches.FindByID(ctx, matchID)
	if err != nil {
		slog.WarnContext(ctx, "failed to load match for FflClubMatchUpdated event", slog.Int("match_id", matchID), slog.Any("error", err))
		return result, nil
	}

	cm, err := c.clubMatches.FindByID(ctx, params.ClubMatchID)
	if err != nil {
		slog.WarnContext(ctx, "failed to load club_match for FflClubMatchUpdated event", slog.Int("club_match_id", params.ClubMatchID), slog.Any("error", err))
		return result, nil
	}

	b, err := json.Marshal(events.FflClubMatchUpdatedPayload{
		ClubMatchID:   params.ClubMatchID,
		MatchID:       matchID,
		RoundID:       match.RoundID,
		DataStatus:    string(cm.DataStatus),
		PlayerMatches: buildPlayerMatchMap(latest),
	})
	if err == nil {
		if err := c.dispatcher.Publish(ctx, events.FflClubMatchUpdated, b); err != nil {
			slog.WarnContext(ctx, "publish FflClubMatchUpdated failed", slog.Int("club_match_id", params.ClubMatchID), slog.Any("error", err))
		}
	}
	return result, nil
}

// DeclareSubs records explicit TM substitution and interchange decisions for a club match.
// Validation and status assignment are delegated to ClubMatch.DeclareSubs.
// Triggers a score recalculation after writing.
func (c *Commands) DeclareSubs(ctx context.Context, clubMatchID int, subs []domain.SubPairing, interchange *domain.SubPairing) ([]domain.PlayerMatch, error) {
	err := c.tx.WithTx(ctx, func(repos WriteRepos) error {
		cm, err := repos.ClubMatches.FindByID(ctx, clubMatchID)
		if err != nil {
			return fmt.Errorf("find club match: %w", err)
		}
		pms, err := repos.PlayerMatches.FindByClubMatchID(ctx, clubMatchID)
		if err != nil {
			return fmt.Errorf("find player matches: %w", err)
		}
		cm.PlayerMatches = pms

		updated, err := cm.DeclareSubs(subs, interchange)
		if err != nil {
			return err
		}

		oldStatus := make(map[int]*domain.PlayerMatchStatus, len(pms))
		for _, pm := range pms {
			oldStatus[pm.ID] = pm.Status
		}

		for _, pm := range updated {
			newStatus := pm.Status
			if old := oldStatus[pm.ID]; statusEqual(old, newStatus) {
				continue
			}
			if newStatus == nil {
				continue
			}
			if err := repos.PlayerMatches.UpdateStatus(ctx, pm.ID, *newStatus); err != nil {
				return fmt.Errorf("update status for player_match %d: %w", pm.ID, err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := c.RecalculateScore(ctx, clubMatchID); err != nil {
		slog.WarnContext(ctx, "recalculate score failed after DeclareSubs", slog.Int("club_match_id", clubMatchID), slog.Any("error", err))
	}

	pms, err := c.playerMatches.FindByClubMatchID(ctx, clubMatchID)
	if err != nil {
		return nil, err
	}

	// Publish FFL.ClubMatchUpdated and FFL.SubsDeclared.
	cm, _ := c.clubMatches.FindByID(ctx, clubMatchID)
	m, _ := c.matches.FindByID(ctx, cm.MatchID)

	if b, err := json.Marshal(events.FflClubMatchUpdatedPayload{
		ClubMatchID:   clubMatchID,
		MatchID:       cm.MatchID,
		RoundID:       m.RoundID,
		DataStatus:    string(cm.DataStatus),
		PlayerMatches: buildPlayerMatchMap(pms),
	}); err == nil {
		if err := c.dispatcher.Publish(ctx, events.FflClubMatchUpdated, b); err != nil {
			slog.WarnContext(ctx, "publish FflClubMatchUpdated failed after DeclareSubs", slog.Int("club_match_id", clubMatchID), slog.Any("error", err))
		}
	}

	return pms, nil
}

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

// buildPlayerMatchMap builds the FflPlayerMatchInfo snapshot from a slice of player_matches.
func buildPlayerMatchMap(pms []domain.PlayerMatch) map[int]events.FflPlayerMatchInfo {
	m := make(map[int]events.FflPlayerMatchInfo, len(pms))
	for _, pm := range pms {
		info := events.FflPlayerMatchInfo{}
		if pm.Position != nil {
			info.Position = string(*pm.Position)
		}
		if pm.Status != nil {
			info.Status = string(*pm.Status)
		}
		if pm.BackupPositions != nil {
			info.BackupPositions = *pm.BackupPositions
		}
		if pm.InterchangePosition != nil {
			info.InterchangePosition = *pm.InterchangePosition
		}
		m[pm.ID] = info
	}
	return m
}

func statusEqual(a, b *domain.PlayerMatchStatus) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

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
