package application

import (
	"context"
	"fmt"

	"xffl/services/ffl/internal/domain"
)

const confidenceThreshold = 0.85

// ResolvedPlayer is a parsed player row that has been matched to an ffl.player_season record.
type ResolvedPlayer struct {
	Parsed         ParsedPlayerRow
	PlayerSeasonID int
	BestMatch      PlayerMatch
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

// DataOpsCommands handles data import operations.
type DataOpsCommands struct {
	tx            TxManager
	playerLookup  PlayerLookup
	playerResolver PlayerResolver
	teamParser    TeamParser
}

func NewDataOpsCommands(tx TxManager, lookup PlayerLookup, resolver PlayerResolver, parser TeamParser) *DataOpsCommands {
	return &DataOpsCommands{
		tx:             tx,
		playerLookup:   lookup,
		playerResolver: resolver,
		teamParser:     parser,
	}
}

// LookupCandidates fetches player names from the AFL service and returns a candidate pool.
// aflIDToPlayerSeasonID maps afl_player_id → player_season_id (built by the caller who
// already has both ffl.player and player_season records available).
func (c *DataOpsCommands) LookupCandidates(ctx context.Context, aflIDToPlayerSeasonID map[int]int) ([]PlayerCandidate, error) {
	aflPlayerIDs := make([]int, 0, len(aflIDToPlayerSeasonID))
	for aflID := range aflIDToPlayerSeasonID {
		aflPlayerIDs = append(aflPlayerIDs, aflID)
	}

	fetched, err := c.playerLookup.LookupPlayers(ctx, aflPlayerIDs)
	if err != nil {
		return nil, err
	}

	candidates := make([]PlayerCandidate, 0, len(fetched))
	for _, f := range fetched {
		psID := aflIDToPlayerSeasonID[f.AFLPlayerID]
		candidates = append(candidates, PlayerCandidate{
			PlayerID:    psID, // player_season_id in squad context
			AFLPlayerID: f.AFLPlayerID,
			Name:        f.Name,
			Club:        f.Club,
		})
	}
	return candidates, nil
}

// ParseTeamSubmission parses a raw forum post and resolves each player against the squad.
// No DB writes occur. The caller reviews the result and calls ImportRoundTeams to confirm.
func (c *DataOpsCommands) ParseTeamSubmission(ctx context.Context, params ParseTeamSubmissionParams, playerSeasons []domain.PlayerSeason, candidates []PlayerCandidate) (ParseTeamSubmissionResult, error) {
	rows, err := c.teamParser.Parse(ctx, params.TeamName, params.Post)
	if err != nil {
		return ParseTeamSubmissionResult{}, fmt.Errorf("parse forum post: %w", err)
	}

	// Build a lookup from AFLPlayerID → candidate (includes PlayerSeasonID from the caller).
	candidateByAFLID := make(map[int]PlayerCandidate, len(candidates))
	for _, cand := range candidates {
		candidateByAFLID[cand.AFLPlayerID] = cand
	}

	resolved := make([]ResolvedPlayer, 0, len(rows))
	var needsReview []int

	for _, row := range rows {
		matches, err := c.playerResolver.Resolve(ctx, row.Name, row.ClubHint, candidates)
		if err != nil {
			return ParseTeamSubmissionResult{}, fmt.Errorf("resolve %q: %w", row.Name, err)
		}

		rp := ResolvedPlayer{Parsed: row}
		if len(matches) > 0 {
			rp.BestMatch = matches[0]
			rp.Confident = matches[0].Confidence >= confidenceThreshold
			rp.PlayerSeasonID = matches[0].Candidate.PlayerID // PlayerID is the player_season_id in squad context
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

// ImportRoundTeams writes the confirmed team submission to the database.
// Each ResolvedPlayer must have a valid PlayerSeasonID set before calling.
func (c *DataOpsCommands) ImportRoundTeams(ctx context.Context, params ImportRoundTeamsParams) ([]domain.PlayerMatch, error) {
	var result []domain.PlayerMatch

	err := c.tx.WithTx(ctx, func(repos WriteRepos) error {
		// Remove any existing player_match records for this club_match.
		if err := repos.PlayerMatches.DeleteByClubMatchID(ctx, params.ClubMatchID); err != nil {
			return fmt.Errorf("clear existing player matches: %w", err)
		}

		result = make([]domain.PlayerMatch, 0, len(params.ResolvedPlayers))
		for _, rp := range params.ResolvedPlayers {
			if rp.PlayerSeasonID == 0 {
				continue // skip unresolved players
			}

			upsertParams := buildUpsertParams(params.ClubMatchID, rp)
			pm, err := repos.PlayerMatches.Upsert(ctx, upsertParams)
			if err != nil {
				return fmt.Errorf("upsert player_match for player_season %d: %w", rp.PlayerSeasonID, err)
			}
			result = append(result, pm)
		}
		return nil
	})

	return result, err
}

func buildUpsertParams(clubMatchID int, rp ResolvedPlayer) domain.UpsertPlayerMatchParams {
	p := rp.Parsed
	status := domain.PlayerMatchStatusNamed

	params := domain.UpsertPlayerMatchParams{
		ClubMatchID:    clubMatchID,
		PlayerSeasonID: rp.PlayerSeasonID,
		Status:         &status,
	}

	if p.BackupPositions != "" || p.InterchangePosition != "" {
		// bench player
		if p.BackupPositions != "" {
			params.BackupPositions = &p.BackupPositions
		}
		if p.InterchangePosition != "" {
			params.InterchangePosition = &p.InterchangePosition
		}
	} else {
		pos := domain.Position(p.Position)
		params.Position = &pos
	}

	if p.Score != nil {
		params.Score = p.Score
	}

	return params
}
