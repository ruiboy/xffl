package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"xffl/contracts/events"
	"xffl/services/ffl/internal/domain"
)

// ByeIneligibleError is returned when a named player's AFL club has a bye
// but they did not play in their club's most recent non-bye match.
type ByeIneligibleError struct {
	PlayerSeasonID int
}

func (e ByeIneligibleError) Error() string {
	return fmt.Sprintf("player_season %d is on a bye but did not play in their club's most recent match: ineligible to be named", e.PlayerSeasonID)
}

// SetTeamParams are the inputs to SetTeam.
type SetTeamParams struct {
	ClubMatchID     int
	Entries         []SetTeamEntry
	ClubMatchNotes  *string // optional notes to write on the club_match row
}

// SetTeamEntry represents a single player assignment in a team.
type SetTeamEntry struct {
	PlayerSeasonID      int
	Position            string
	BackupPositions     *string
	InterchangePosition *string
	DisplayOrder        int     // display position within the player's position group (or bench)
	Notes               *string // optional notes to write on the player_match row
	Score               *int    // optional seed score for new players (AFL events are authoritative once set)
}

// SetTeam persists a complete team for a club match using diff-based persistence to
// preserve afl_player_match_id links for returning players. It validates team composition
// via the domain, computes a provisional score, updates data_status, and publishes
// FFL.TeamSubmitted.
func (c *Commands) SetTeam(ctx context.Context, params SetTeamParams) ([]domain.PlayerMatch, error) {
	var result []domain.PlayerMatch
	var matchID int

	// Resolve bye info before the transaction — this is a network call to the AFL service.
	byeByFFFLPS, err := c.lookupByeInfoForTeam(ctx, params)
	if err != nil {
		return nil, err
	}

	err = c.tx.WithTx(ctx, func(repos WriteRepos) error {
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

		// Apply bye status and scores. Starters on a bye get drv_afl_status = "bye" and
		// drv_score computed from season average. Bench players on a bye get the status only
		// (score is set at activation via DeclareSubs).
		for i, pm := range newPlayers {
			bi, hasBye := byeByFFFLPS[pm.PlayerSeasonID]
			if !hasBye {
				continue
			}
			if !bi.PlayedLast {
				return ByeIneligibleError{PlayerSeasonID: pm.PlayerSeasonID}
			}
			status := domain.AFLStatusBye
			newPlayers[i].AFLStatus = &status
			if pm.Position != nil {
				score := pm.CalculateByeScore(domain.AFLAvgStats{
					Goals: bi.AvgGoals, Kicks: bi.AvgKicks, Handballs: bi.AvgHandballs,
					Marks: bi.AvgMarks, Tackles: bi.AvgTackles, Hitouts: bi.AvgHitouts,
				})
				newPlayers[i].Score = score
			}
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
		if params.ClubMatchNotes != nil {
			if err := repos.ClubMatches.UpdateNotes(ctx, cm.ID, *params.ClubMatchNotes); err != nil {
				return fmt.Errorf("update club match notes: %w", err)
			}
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
		oldPosition := make(map[int]*domain.Position, len(pms))
		for _, pm := range pms {
			oldStatus[pm.ID] = pm.Status
			oldPosition[pm.ID] = pm.Position
		}

		for _, pm := range updated {
			statusChanged := !statusEqual(oldStatus[pm.ID], pm.Status)
			posChanged := !positionEqual(oldPosition[pm.ID], pm.Position)
			if !statusChanged && !posChanged {
				continue
			}
			if statusChanged && pm.Status != nil {
				if err := repos.PlayerMatches.UpdateStatus(ctx, pm.ID, *pm.Status); err != nil {
					return fmt.Errorf("update status for player_match %d: %w", pm.ID, err)
				}
			}
			if posChanged {
				if err := repos.PlayerMatches.UpdatePosition(ctx, pm.ID, pm.Position); err != nil {
					return fmt.Errorf("update position for player_match %d: %w", pm.ID, err)
				}
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

	// For any newly activated bench player whose AFL club has a bye, RecalculateScore
	// won't find AFL stats. Compute and persist their bye score explicitly.
	if err := c.applyByeScoresForActivated(ctx, clubMatchID); err != nil {
		slog.WarnContext(ctx, "apply bye scores after DeclareSubs failed", slog.Int("club_match_id", clubMatchID), slog.Any("error", err))
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
	pm.DisplayOrder = e.DisplayOrder
	pm.Notes = e.Notes
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

func positionEqual(a, b *domain.Position) bool {
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
		AFLStatus:           pm.AFLStatus,
		BackupPositions:     pm.BackupPositions,
		InterchangePosition: pm.InterchangePosition,
		DisplayOrder:        pm.DisplayOrder,
		Notes:               pm.Notes,
	}
	if pm.Score != 0 {
		s := pm.Score
		params.Score = &s
	}
	return params
}

// lookupByeInfoForTeam resolves bye status and season averages for all players in a SetTeam
// submission. Returns a map of FFL player_season_id → ByePlayerInfo (only for players whose
// AFL club has a bye in this round). Returns nil if the round has no AFL link.
func (c *Commands) lookupByeInfoForTeam(ctx context.Context, params SetTeamParams) (map[int]ByePlayerInfo, error) {
	cm, err := c.clubMatches.FindByID(ctx, params.ClubMatchID)
	if err != nil {
		return nil, fmt.Errorf("find club match for bye lookup: %w", err)
	}
	m, err := c.matches.FindByID(ctx, cm.MatchID)
	if err != nil {
		return nil, fmt.Errorf("find match for bye lookup: %w", err)
	}
	r, err := c.rounds.FindByID(ctx, m.RoundID)
	if err != nil {
		return nil, fmt.Errorf("find round for bye lookup: %w", err)
	}
	if r.AFLRoundID == 0 {
		return nil, nil
	}

	var psIDs []int
	for _, e := range params.Entries {
		if e.PlayerSeasonID != 0 {
			psIDs = append(psIDs, e.PlayerSeasonID)
		}
	}
	if len(psIDs) == 0 {
		return nil, nil
	}

	playerSeasons, err := c.playerSeasons.FindByIDs(ctx, psIDs)
	if err != nil {
		return nil, fmt.Errorf("find player seasons for bye lookup: %w", err)
	}

	fflToAFLPS := make(map[int]int, len(playerSeasons))
	var aflPSIDs []int
	for _, ps := range playerSeasons {
		if ps.AFLPlayerSeasonID != 0 {
			fflToAFLPS[ps.ID] = ps.AFLPlayerSeasonID
			aflPSIDs = append(aflPSIDs, ps.AFLPlayerSeasonID)
		}
	}
	if len(aflPSIDs) == 0 {
		return nil, nil
	}

	byeSlice, err := c.playerLookup.LookupByeInfo(ctx, aflPSIDs, r.AFLRoundID)
	if err != nil {
		return nil, fmt.Errorf("lookup bye info from AFL service: %w", err)
	}

	byeByAFLPS := make(map[int]ByePlayerInfo, len(byeSlice))
	for _, bi := range byeSlice {
		byeByAFLPS[bi.PlayerSeasonID] = bi
	}

	result := make(map[int]ByePlayerInfo)
	for fflPSID, aflPSID := range fflToAFLPS {
		if bi, ok := byeByAFLPS[aflPSID]; ok && bi.HasBye {
			result[fflPSID] = bi
		}
	}
	return result, nil
}

// applyByeScoresForActivated sets drv_score for any bench player who was just activated
// (subbed_in or interchanged_in) and whose AFL club has a bye. RecalculateScore won't
// find AFL stats for these players, so we compute from their season average.
func (c *Commands) applyByeScoresForActivated(ctx context.Context, clubMatchID int) error {
	pms, err := c.playerMatches.FindByClubMatchID(ctx, clubMatchID)
	if err != nil {
		return err
	}

	activated := make([]domain.PlayerMatch, 0)
	for _, pm := range pms {
		isActivated := pm.Status != nil &&
			(*pm.Status == domain.PlayerMatchStatusSubbedIn || *pm.Status == domain.PlayerMatchStatusInterchangedIn)
		isBye := pm.AFLStatus != nil && *pm.AFLStatus == domain.AFLStatusBye
		if isActivated && isBye && pm.Position != nil && pm.Score == 0 {
			activated = append(activated, pm)
		}
	}
	if len(activated) == 0 {
		return nil
	}

	// Resolve AFL round ID.
	cm, err := c.clubMatches.FindByID(ctx, clubMatchID)
	if err != nil {
		return err
	}
	m, err := c.matches.FindByID(ctx, cm.MatchID)
	if err != nil {
		return err
	}
	r, err := c.rounds.FindByID(ctx, m.RoundID)
	if err != nil {
		return err
	}
	if r.AFLRoundID == 0 {
		return nil
	}

	// Collect AFL player_season_ids for activated bye bench players.
	psIDs := make([]int, len(activated))
	for i, pm := range activated {
		psIDs[i] = pm.PlayerSeasonID
	}
	playerSeasons, err := c.playerSeasons.FindByIDs(ctx, psIDs)
	if err != nil {
		return err
	}
	fflToAFLPS := make(map[int]int, len(playerSeasons))
	var aflPSIDs []int
	for _, ps := range playerSeasons {
		if ps.AFLPlayerSeasonID != 0 {
			fflToAFLPS[ps.ID] = ps.AFLPlayerSeasonID
			aflPSIDs = append(aflPSIDs, ps.AFLPlayerSeasonID)
		}
	}
	if len(aflPSIDs) == 0 {
		return nil
	}

	byeSlice, err := c.playerLookup.LookupByeInfo(ctx, aflPSIDs, r.AFLRoundID)
	if err != nil {
		return fmt.Errorf("lookup bye info for activated bench player: %w", err)
	}
	byeByAFLPS := make(map[int]ByePlayerInfo, len(byeSlice))
	for _, bi := range byeSlice {
		byeByAFLPS[bi.PlayerSeasonID] = bi
	}

	for _, pm := range activated {
		aflPSID, ok := fflToAFLPS[pm.PlayerSeasonID]
		if !ok {
			continue
		}
		bi, ok := byeByAFLPS[aflPSID]
		if !ok {
			continue
		}
		score := pm.CalculateByeScore(domain.AFLAvgStats{
			Goals: bi.AvgGoals, Kicks: bi.AvgKicks, Handballs: bi.AvgHandballs,
			Marks: bi.AvgMarks, Tackles: bi.AvgTackles, Hitouts: bi.AvgHitouts,
		})
		s := score
		if _, err := c.playerMatches.Upsert(ctx, domain.UpsertPlayerMatchParams{
			ClubMatchID:         pm.ClubMatchID,
			PlayerSeasonID:      pm.PlayerSeasonID,
			Position:            pm.Position,
			Status:              pm.Status,
			BackupPositions:     pm.BackupPositions,
			InterchangePosition: pm.InterchangePosition,
			DisplayOrder:        pm.DisplayOrder,
			Score:               &s,
		}); err != nil {
			slog.WarnContext(ctx, "upsert bye score for activated bench player failed",
				slog.Int("player_match_id", pm.ID), slog.Any("error", err))
		}
	}

	// Re-sum the club match total.
	updated, err := c.playerMatches.FindByClubMatchID(ctx, clubMatchID)
	if err != nil {
		return err
	}
	cm.PlayerMatches = updated
	return c.clubMatches.UpdateScore(ctx, clubMatchID, cm.Score())
}

// ReorderDirection indicates which direction to move a player match within its position group.
type ReorderDirection string

const (
	ReorderUp   ReorderDirection = "UP"
	ReorderDown ReorderDirection = "DOWN"
)

// ReorderPlayerMatch swaps the display_order of a player match with its neighbour in the
// same position group (or bench group), then returns all player matches for the club match.
func (c *Commands) ReorderPlayerMatch(ctx context.Context, id int, direction ReorderDirection) ([]domain.PlayerMatch, error) {
	var result []domain.PlayerMatch
	err := c.tx.WithTx(ctx, func(repos WriteRepos) error {
		target, err := repos.PlayerMatches.FindByID(ctx, id)
		if err != nil {
			return fmt.Errorf("find player match: %w", err)
		}

		all, err := repos.PlayerMatches.FindByClubMatchID(ctx, target.ClubMatchID)
		if err != nil {
			return fmt.Errorf("find player matches: %w", err)
		}

		// Build the ordered group: players in the same position slot (starters share position,
		// bench players share nil position). Already sorted by display_order from the query.
		group := make([]domain.PlayerMatch, 0)
		for _, pm := range all {
			sameGroup := (pm.Position == nil && target.Position == nil) ||
				(pm.Position != nil && target.Position != nil && *pm.Position == *target.Position)
			// Bench: both have BackupPositions set; starter: both have Position set.
			isBench := pm.BackupPositions != nil
			targetIsBench := target.BackupPositions != nil
			if isBench != targetIsBench {
				continue
			}
			if sameGroup {
				group = append(group, pm)
			}
		}

		// Find the target's index in the group.
		idx := -1
		for i, pm := range group {
			if pm.ID == id {
				idx = i
				break
			}
		}
		if idx < 0 {
			return fmt.Errorf("player match %d not found in its position group", id)
		}

		// Find the neighbour to swap with.
		var neighbourIdx int
		if direction == ReorderUp {
			if idx == 0 {
				return nil // already first, no-op
			}
			neighbourIdx = idx - 1
		} else {
			if idx == len(group)-1 {
				return nil // already last, no-op
			}
			neighbourIdx = idx + 1
		}

		a, b := group[idx], group[neighbourIdx]
		if err := repos.PlayerMatches.UpdateDisplayOrder(ctx, a.ID, b.DisplayOrder); err != nil {
			return fmt.Errorf("update display order for %d: %w", a.ID, err)
		}
		if err := repos.PlayerMatches.UpdateDisplayOrder(ctx, b.ID, a.DisplayOrder); err != nil {
			return fmt.Errorf("update display order for %d: %w", b.ID, err)
		}

		result, err = repos.PlayerMatches.FindByClubMatchID(ctx, target.ClubMatchID)
		return err
	})
	return result, err
}
