package dataops

import (
	"context"
	"fmt"

	"xffl/services/ffl/internal/application"
	"xffl/services/ffl/internal/domain"
)

// ResolvedSquadMember is a parsed squad member matched to an AFL player_season.
type ResolvedSquadMember struct {
	Parsed            application.ParsedSquadMember
	AFLPlayerSeasonID int
	BestMatch         application.PlayerNameMatch
	Confident         bool // true if best-match confidence >= threshold
}

// ResolvedSquad is one club's parsed squad with each member resolved against the
// candidate pool. ClubName is the FFL club name exactly as written in the thread
// header; assigning it to a club_season is the reviewer's job and is supplied to
// ImportSquad, not inferred here.
type ResolvedSquad struct {
	ClubName    string
	Members     []ResolvedSquadMember
	NeedsReview []int // indices into Members with confidence < threshold
}

// ParseSquadThreadResult is returned to the caller for review before ImportSquad
// confirms. No DB writes have occurred.
type ParseSquadThreadResult struct {
	Squads []ResolvedSquad
}

// ParseSquadThreadForSeason sources the AFL season's players as the candidate pool
// (via the cross-service LookupPlayerSeasonsBySeasonID lookup) and parses+resolves
// the thread against it. This is the entry point a caller drives with just the AFL
// season and the pasted thread; ParseSquadThread is the lower level that takes an
// explicit pool.
func (c *DataOpsCommands) ParseSquadThreadForSeason(ctx context.Context, aflSeasonID int, thread string) (ParseSquadThreadResult, error) {
	candidates, err := c.playerLookup.LookupPlayerSeasonsBySeasonID(ctx, aflSeasonID)
	if err != nil {
		return ParseSquadThreadResult{}, fmt.Errorf("source season candidates: %w", err)
	}
	return c.ParseSquadThread(ctx, thread, candidates)
}

// ParseSquadThread parses a squads thread and resolves every member against the
// AFL candidate pool. No DB writes occur; the caller reviews the result and calls
// ImportSquad once per club to confirm. The candidate pool is supplied by the
// caller (same contract as ParseTeamSubmission) and its AFLPlayerSeasonID is the
// handle ImportSquad needs.
func (c *DataOpsCommands) ParseSquadThread(ctx context.Context, thread string, candidates []application.PlayerCandidate) (ParseSquadThreadResult, error) {
	parsed, err := c.squadParser.ParseSquads(ctx, thread)
	if err != nil {
		return ParseSquadThreadResult{}, fmt.Errorf("parse squads thread: %w", err)
	}

	result := ParseSquadThreadResult{Squads: make([]ResolvedSquad, 0, len(parsed))}
	for _, squad := range parsed {
		rs := ResolvedSquad{ClubName: squad.ClubName, Members: make([]ResolvedSquadMember, 0, len(squad.Members))}
		for _, m := range squad.Members {
			matches, err := c.playerResolver.Resolve(ctx, m.Name, m.ClubHint, candidates)
			if err != nil {
				return ParseSquadThreadResult{}, fmt.Errorf("resolve %q: %w", m.Name, err)
			}
			rm := ResolvedSquadMember{Parsed: m}
			if len(matches) > 0 {
				rm.BestMatch = matches[0]
				rm.Confident = matches[0].Confidence >= confidenceThreshold
				rm.AFLPlayerSeasonID = matches[0].Candidate.AFLPlayerSeasonID
			}
			if !rm.Confident {
				rs.NeedsReview = append(rs.NeedsReview, len(rs.Members))
			}
			rs.Members = append(rs.Members, rm)
		}
		result.Squads = append(result.Squads, rs)
	}
	return result, nil
}

// ImportSquadParams are the confirmed inputs to ImportSquad: one club's members,
// assigned by the reviewer to a club_season.
type ImportSquadParams struct {
	ClubSeasonID int
	FromRoundID  *int // round the squad takes effect from; nil for season start
	Members      []ResolvedSquadMember
}

// ImportSquad adds each resolved member to the club_season squad via
// AddPlayerToSeason (which resolves the AFL player_season, find-or-creates the
// ffl.player, and creates the player_season). Members that did not resolve to an
// AFL player_season (AFLPlayerSeasonID == 0) are skipped — the reviewer resolves
// those before confirming.
func (c *DataOpsCommands) ImportSquad(ctx context.Context, params ImportSquadParams) ([]domain.PlayerSeason, error) {
	added := make([]domain.PlayerSeason, 0, len(params.Members))
	for _, m := range params.Members {
		if m.AFLPlayerSeasonID == 0 {
			continue
		}
		ps, err := c.commands.AddPlayerToSeason(ctx, params.ClubSeasonID, m.AFLPlayerSeasonID, params.FromRoundID, m.Parsed.CostCents)
		if err != nil {
			return added, fmt.Errorf("add player %q to season: %w", m.Parsed.Name, err)
		}
		added = append(added, ps)
	}
	return added, nil
}
