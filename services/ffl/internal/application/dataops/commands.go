package dataops

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"xffl/contracts/events"
	"xffl/services/ffl/internal/application"
	"xffl/services/ffl/internal/domain"
	sharedevents "xffl/shared/events"
)

const confidenceThreshold = 0.85

// ResolvedPlayer is a parsed player row that has been matched to an ffl.player_season record.
type ResolvedPlayer struct {
	Parsed         application.ParsedPlayerRow
	PlayerSeasonID int
	BestMatch      application.PlayerNameMatch
	Confident      bool // true if confidence >= threshold
}

// ParseTeamSubmissionParams are the inputs to ParseTeamSubmission.
type ParseTeamSubmissionParams struct {
	ClubSeasonID int
	TeamName     string // FFL team name as written in posts (e.g. "Ruiboys")
	Post         string // raw pasted forum post text
}

// ParseTeamSubmissionResult is returned to the caller for review before confirming.
type ParseTeamSubmissionResult struct {
	ClubSeasonID    int
	ResolvedPlayers []ResolvedPlayer
	// NeedsReview contains indices into ResolvedPlayers where confidence < threshold
	NeedsReview []int
}

// ImportRoundTeamsParams are the confirmed inputs to ImportRoundTeams.
type ImportRoundTeamsParams struct {
	ClubMatchID     int
	ResolvedPlayers []ResolvedPlayer
}

// MarkTeamFinalParams are the inputs to MarkTeamFinal.
type MarkTeamFinalParams struct {
	ClubMatchID int
	MatchID     int
	RoundID     int
}

// DataOpsCommands handles data import operations.
type DataOpsCommands struct {
	tx             application.TxManager
	playerLookup   application.PlayerLookup
	playerResolver application.PlayerResolver
	teamParser     application.TeamParser
	squadParser    application.SquadThreadParser
	dispatcher     sharedevents.Dispatcher
	commands       *application.Commands
}

func NewDataOpsCommands(tx application.TxManager, lookup application.PlayerLookup, resolver application.PlayerResolver, parser application.TeamParser, dispatcher sharedevents.Dispatcher, commands *application.Commands, squadParser application.SquadThreadParser) *DataOpsCommands {
	return &DataOpsCommands{
		tx:             tx,
		playerLookup:   lookup,
		playerResolver: resolver,
		teamParser:     parser,
		squadParser:    squadParser,
		dispatcher:     dispatcher,
		commands:       commands,
	}
}

// LookupSquadCandidates builds a squad's resolution pool from the FFL season's AFL
// season, so each member's name and club come from the season being imported (via
// the linked afl_player_season) — not from the player's whole career, which would
// surface a former club (e.g. a 2025 Port player showing as his 2021 Bulldogs).
// squadByAFLPlayerSeasonID maps afl_player_season_id → ffl player_season_id.
func (c *DataOpsCommands) LookupSquadCandidates(ctx context.Context, aflSeasonID int, squadByAFLPlayerSeasonID map[int]int) ([]application.PlayerCandidate, error) {
	seasonPlayers, err := c.playerLookup.LookupPlayerSeasonsBySeasonID(ctx, aflSeasonID)
	if err != nil {
		return nil, err
	}
	candidates := make([]application.PlayerCandidate, 0, len(squadByAFLPlayerSeasonID))
	for _, sp := range seasonPlayers {
		psID, ok := squadByAFLPlayerSeasonID[sp.AFLPlayerSeasonID]
		if !ok {
			continue
		}
		candidates = append(candidates, application.PlayerCandidate{
			PlayerID:          psID, // ffl player_season_id in squad context
			AFLPlayerID:       sp.AFLPlayerID,
			AFLPlayerSeasonID: sp.AFLPlayerSeasonID,
			Name:              sp.Name,
			Club:              sp.Club,
		})
	}
	return candidates, nil
}

// ParseTeamSubmission parses a raw forum post and resolves each player against the squad.
// No DB writes occur. The caller reviews the result and calls ImportRoundTeams to confirm.
func (c *DataOpsCommands) ParseTeamSubmission(ctx context.Context, params ParseTeamSubmissionParams, playerSeasons []domain.PlayerSeason, candidates []application.PlayerCandidate) (ParseTeamSubmissionResult, error) {
	rows, err := c.teamParser.Parse(ctx, params.TeamName, params.Post)
	if err != nil {
		return ParseTeamSubmissionResult{}, fmt.Errorf("parse forum post: %w", err)
	}

	// Build a lookup from AFLPlayerID → candidate (includes PlayerSeasonID from the caller).
	candidateByAFLID := make(map[int]application.PlayerCandidate, len(candidates))
	for _, cand := range candidates {
		candidateByAFLID[cand.AFLPlayerID] = cand
	}

	resolved := make([]ResolvedPlayer, 0, len(rows))
	var needsReview []int

	for _, row := range rows {
		nameMatches, err := c.playerResolver.Resolve(ctx, row.Name, row.ClubHint, candidates)
		if err != nil {
			return ParseTeamSubmissionResult{}, fmt.Errorf("resolve %q: %w", row.Name, err)
		}

		rp := ResolvedPlayer{Parsed: row}
		if len(nameMatches) > 0 {
			rp.BestMatch = nameMatches[0]
			rp.Confident = nameMatches[0].Confidence >= confidenceThreshold
			rp.PlayerSeasonID = nameMatches[0].Candidate.PlayerID // PlayerID is the player_season_id in squad context
		}
		if !rp.Confident {
			needsReview = append(needsReview, len(resolved))
		}
		resolved = append(resolved, rp)
	}

	return ParseTeamSubmissionResult{
		ClubSeasonID:    params.ClubSeasonID,
		ResolvedPlayers: resolved,
		NeedsReview:     needsReview,
	}, nil
}

// ImportRoundTeams converts resolved players to team entries and delegates to teamSubmitter.SetTeam,
// which handles validation, diff-based persistence, scoring, and event publishing.
// display_order is auto-assigned: starters and bench are numbered within each position group by parse order.
// When a player has a posted score, "posted:NN" is written to player_match.notes.
// The sum of all posted player scores is written to club_match.notes as "posted:NN".
func (c *DataOpsCommands) ImportRoundTeams(ctx context.Context, params ImportRoundTeamsParams) ([]domain.PlayerMatch, error) {
	positionCount := make(map[string]int)
	entries := make([]application.SetTeamEntry, 0, len(params.ResolvedPlayers))
	postedTotal := 0
	anyPosted := false
	for _, rp := range params.ResolvedPlayers {
		if rp.PlayerSeasonID == 0 {
			continue
		}
		groupKey := rp.Parsed.Position
		if rp.Parsed.BackupPositions != "" {
			groupKey = "bench"
		}
		positionCount[groupKey]++
		e := application.SetTeamEntry{
			PlayerSeasonID: rp.PlayerSeasonID,
			Position:       rp.Parsed.Position,
			DisplayOrder:   positionCount[groupKey],
			Score:          rp.Parsed.Score,
		}
		if rp.Parsed.Score != nil {
			note := fmt.Sprintf("posted:%d", *rp.Parsed.Score)
			e.Notes = &note
			postedTotal += *rp.Parsed.Score
			anyPosted = true
		}
		if rp.Parsed.BackupPositions != "" {
			e.BackupPositions = &rp.Parsed.BackupPositions
		}
		if rp.Parsed.InterchangePosition != "" {
			e.InterchangePosition = &rp.Parsed.InterchangePosition
		}
		entries = append(entries, e)
	}
	sp := application.SetTeamParams{
		ClubMatchID: params.ClubMatchID,
		Entries:     entries,
	}
	if anyPosted {
		sp.ClubMatchPostedScore = &postedTotal
	}
	return c.commands.SetTeam(ctx, sp)
}

// MarkTeamSubmitted reverts the club_match data_status to 'submitted' and publishes FFL.ClubMatchUpdated(submitted).
func (c *DataOpsCommands) MarkTeamSubmitted(ctx context.Context, params MarkTeamFinalParams) error {
	err := c.tx.WithTx(ctx, func(repos application.WriteRepos) error {
		return repos.ClubMatches.UpdateDataStatus(ctx, params.ClubMatchID, domain.ClubMatchDataSubmitted)
	})
	if err != nil {
		return err
	}

	pms, err := c.commands.FindPlayerMatchesByClubMatch(ctx, params.ClubMatchID)
	if err != nil {
		slog.WarnContext(ctx, "load player_matches failed for FflClubMatchUpdated", slog.Int("club_match_id", params.ClubMatchID), slog.Any("error", err))
	}

	b, err := json.Marshal(events.FflClubMatchUpdatedPayload{
		ClubMatchID:   params.ClubMatchID,
		MatchID:       params.MatchID,
		RoundID:       params.RoundID,
		DataStatus:    string(domain.ClubMatchDataSubmitted),
		PlayerMatches: application.BuildPlayerMatchMap(pms),
	})
	if err == nil {
		if err := c.dispatcher.Publish(ctx, events.FflClubMatchUpdated, b); err != nil {
			slog.WarnContext(ctx, "publish FflClubMatchUpdated(submitted) failed", slog.Int("club_match_id", params.ClubMatchID), slog.Any("error", err))
		}
	}
	return nil
}

// MarkTeamFinal sets the club_match data_status to 'final' and publishes FFL.ClubMatchUpdated(final).
func (c *DataOpsCommands) MarkTeamFinal(ctx context.Context, params MarkTeamFinalParams) error {
	err := c.tx.WithTx(ctx, func(repos application.WriteRepos) error {
		return repos.ClubMatches.UpdateDataStatus(ctx, params.ClubMatchID, domain.ClubMatchDataFinal)
	})
	if err != nil {
		return err
	}

	pms, err := c.commands.FindPlayerMatchesByClubMatch(ctx, params.ClubMatchID)
	if err != nil {
		slog.WarnContext(ctx, "load player_matches failed for FflClubMatchUpdated", slog.Int("club_match_id", params.ClubMatchID), slog.Any("error", err))
	}

	b, err := json.Marshal(events.FflClubMatchUpdatedPayload{
		ClubMatchID:   params.ClubMatchID,
		MatchID:       params.MatchID,
		RoundID:       params.RoundID,
		DataStatus:    string(domain.ClubMatchDataFinal),
		PlayerMatches: application.BuildPlayerMatchMap(pms),
	})
	if err == nil {
		if err := c.dispatcher.Publish(ctx, events.FflClubMatchUpdated, b); err != nil {
			slog.WarnContext(ctx, "publish FflClubMatchUpdated(final) failed", slog.Int("club_match_id", params.ClubMatchID), slog.Any("error", err))
		}
	}
	return nil
}
