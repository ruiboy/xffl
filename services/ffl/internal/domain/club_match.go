package domain

import (
	"context"
	"fmt"
	"strings"
)

// ClubMatchDataStatus tracks whether a team has been submitted and scored.
type ClubMatchDataStatus string

const (
	ClubMatchDataNoData    ClubMatchDataStatus = "no_data"
	ClubMatchDataSubmitted ClubMatchDataStatus = "submitted"
	ClubMatchDataFinal     ClubMatchDataStatus = "final"
)

type ClubMatch struct {
	ID            int
	MatchID       int
	ClubSeasonID  int
	DataStatus    ClubMatchDataStatus
	StoredScore   int
	PlayerMatches []PlayerMatch
}

// TeamSubmitted is the domain event raised when a club match team is submitted.
// Yeah, this is not adding a lot of value right now, but it demonstrates the
// concept of domain events (cf. integration events).
type TeamSubmitted struct {
	ClubMatchID int
}

// SubmitTeam validates the player list against team composition rules, replaces
// the club match's player matches, and transitions data_status to submitted.
// Returns a TeamSubmitted domain event on success.
func (cm *ClubMatch) SubmitTeam(players []PlayerMatch) (TeamSubmitted, error) {
	if err := validateTeam(players); err != nil {
		return TeamSubmitted{}, err
	}
	cm.PlayerMatches = players
	cm.DataStatus = ClubMatchDataSubmitted
	return TeamSubmitted{ClubMatchID: cm.ID}, nil
}

// validateTeam enforces team composition rules against a set of player matches.
// It returns a descriptive error if any rule is violated, or nil if the team is valid.
// Teams need not be full — all constraints are upper bounds, not minimums.
func validateTeam(entries []PlayerMatch) error {
	starterCounts := make(map[Position]int)
	var benchPlayers []PlayerMatch
	interchangeCount := 0

	for _, e := range entries {
		if e.InterchangePosition != nil && e.BackupPositions == nil {
			return fmt.Errorf("team: interchange position requires backup positions to be set")
		}
		if e.BackupPositions != nil {
			benchPlayers = append(benchPlayers, e)
			if e.InterchangePosition != nil {
				interchangeCount++
			}
		} else {
			if e.Position == nil {
				return fmt.Errorf("team: starter must have a position")
			}
			starterCounts[*e.Position]++
		}
	}

	// Rule 1: starter count per position ≤ PositionSlots[pos].
	for pos, count := range starterCounts {
		max, ok := PositionSlots[pos]
		if !ok {
			return fmt.Errorf("team: unknown position %q", pos)
		}
		if count > max {
			return fmt.Errorf("team: position %q has %d players, maximum is %d", pos, count, max)
		}
	}

	// Rule 2: total bench ≤ 4.
	if len(benchPlayers) > 4 {
		return fmt.Errorf("team: bench has %d players, maximum is 4", len(benchPlayers))
	}

	benchStarCount := 0
	coveredPositions := make(map[Position]bool)

	for _, bp := range benchPlayers {
		if bp.BackupPositions == nil {
			continue
		}
		positions := parsePositions(*bp.BackupPositions)
		isBenchStar := len(positions) == 1 && positions[0] == PositionStar

		if isBenchStar {
			// Rule 3: at most 1 backup star.
			benchStarCount++
			if benchStarCount > 1 {
				return fmt.Errorf("team: at most 1 backup star allowed on the bench")
			}
		} else {
			// Rule 4: non-star bench players have exactly 2 backup positions, none "star".
			if len(positions) != 2 {
				return fmt.Errorf("team: non-star bench player must have exactly 2 backup positions, got %d", len(positions))
			}
			for _, pos := range positions {
				if pos == PositionStar {
					return fmt.Errorf("team: non-star bench player cannot list star as a backup position")
				}
				if _, ok := PositionSlots[pos]; !ok {
					return fmt.Errorf("team: unknown backup position %q", pos)
				}
				// Rule 5: each non-star position covered by at most one bench player.
				if coveredPositions[pos] {
					return fmt.Errorf("team: position %q is already covered by another bench player", pos)
				}
				coveredPositions[pos] = true
			}
		}
	}

	// Rule 6: at most 1 interchange position across all bench players.
	if interchangeCount > 1 {
		return fmt.Errorf("team: at most 1 interchange position allowed, got %d", interchangeCount)
	}

	// Rule 7: interchange position must be a recognised Position and one of the player's own backup positions.
	for _, bp := range benchPlayers {
		if bp.InterchangePosition != nil {
			pos := Position(*bp.InterchangePosition)
			if _, ok := PositionSlots[pos]; !ok {
				return fmt.Errorf("team: interchange position %q is not a valid position", pos)
			}
			if !containsPosition(*bp.BackupPositions, pos) {
				return fmt.Errorf("team: interchange position %q is not one of this player's backup positions", pos)
			}
		}
	}

	return nil
}

// containsPosition checks whether a comma-separated positions string contains pos.
func containsPosition(positions string, pos Position) bool {
	for _, p := range strings.Split(positions, ",") {
		if Position(strings.TrimSpace(p)) == pos {
			return true
		}
	}
	return false
}

// SubPairing is an explicit TM declaration pairing a starter being replaced with the bench
// player covering them.
type SubPairing struct {
	ReplacedPMID  int // starter going out (must be DNP for subs; any status for interchange)
	ReplacingPMID int // bench player coming in
}

// DeclareSubs records explicit TM sub/interchange decisions for this club match.
// subs lists starter→bench pairings for substitutions (replaced starter must be DNP).
// interchange is an optional single starter→bench pairing for the interchange slot.
// All prior sub/interchange statuses are reset to named before applying new declarations.
// Returns the full player match slice with updated statuses.
func (cm ClubMatch) DeclareSubs(subs []SubPairing, interchange *SubPairing) ([]PlayerMatch, error) {
	if cm.DataStatus == ClubMatchDataFinal {
		return nil, fmt.Errorf("club match %d is already final", cm.ID)
	}

	pmByID := make(map[int]*PlayerMatch, len(cm.PlayerMatches))
	updated := make([]PlayerMatch, len(cm.PlayerMatches))
	copy(updated, cm.PlayerMatches)
	for i := range updated {
		pmByID[updated[i].ID] = &updated[i]
	}

	// Validate subs before mutating anything.
	for _, pair := range subs {
		replaced, ok := pmByID[pair.ReplacedPMID]
		if !ok {
			return nil, fmt.Errorf("player_match %d not found in club match", pair.ReplacedPMID)
		}
		if replaced.BackupPositions != nil {
			return nil, fmt.Errorf("player_match %d is a bench player, cannot be subbed out", pair.ReplacedPMID)
		}
		if replaced.AFLStatus == nil || *replaced.AFLStatus != AFLStatusDNP {
			return nil, fmt.Errorf("player_match %d is not DNP, cannot be subbed out", pair.ReplacedPMID)
		}
		replacing, ok := pmByID[pair.ReplacingPMID]
		if !ok {
			return nil, fmt.Errorf("player_match %d not found in club match", pair.ReplacingPMID)
		}
		if replacing.BackupPositions == nil {
			return nil, fmt.Errorf("player_match %d is not a bench player, cannot be subbed in", pair.ReplacingPMID)
		}
	}
	if interchange != nil {
		replaced, ok := pmByID[interchange.ReplacedPMID]
		if !ok {
			return nil, fmt.Errorf("player_match %d not found in club match", interchange.ReplacedPMID)
		}
		if replaced.BackupPositions != nil {
			return nil, fmt.Errorf("player_match %d is a bench player, cannot be interchanged out", interchange.ReplacedPMID)
		}
		replacing, ok := pmByID[interchange.ReplacingPMID]
		if !ok {
			return nil, fmt.Errorf("player_match %d not found in club match", interchange.ReplacingPMID)
		}
		if replacing.BackupPositions == nil {
			return nil, fmt.Errorf("player_match %d is not a bench player, cannot be interchanged in", interchange.ReplacingPMID)
		}
	}

	// Reset all prior sub/interchange statuses to named.
	named := PlayerMatchStatusNamed
	for i := range updated {
		if s := updated[i].Status; s != nil {
			switch *s {
			case PlayerMatchStatusSubbedOut, PlayerMatchStatusSubbedIn,
				PlayerMatchStatusInterchangedOut, PlayerMatchStatusInterchangedIn:
				updated[i].Status = &named
			}
		}
	}

	// Apply new declarations.
	for _, pair := range subs {
		subbedOut := PlayerMatchStatusSubbedOut
		subbedIn := PlayerMatchStatusSubbedIn
		pmByID[pair.ReplacedPMID].Status = &subbedOut
		pmByID[pair.ReplacingPMID].Status = &subbedIn
	}
	if interchange != nil {
		interchangedOut := PlayerMatchStatusInterchangedOut
		interchangedIn := PlayerMatchStatusInterchangedIn
		pmByID[interchange.ReplacedPMID].Status = &interchangedOut
		pmByID[interchange.ReplacingPMID].Status = &interchangedIn
	}

	return updated, nil
}

// Score computes the total fantasy score for this club match from explicit TM declarations.
// Starters with status named score; subbed_out and interchanged_out starters do not.
// Bench players score only when status is subbed_in or interchanged_in.
// A DNP starter with no TM declaration scores zero.
func (cm ClubMatch) Score() int {
	total := 0
	for _, pm := range cm.PlayerMatches {
		if pm.isBench() {
			if pm.Status != nil && (*pm.Status == PlayerMatchStatusSubbedIn || *pm.Status == PlayerMatchStatusInterchangedIn) {
				total += pm.Score
			}
		} else {
			if pm.Status == nil || *pm.Status == PlayerMatchStatusNamed {
				total += pm.Score
			}
		}
	}
	return total
}

type ClubMatchRepository interface {
	FindByMatchID(ctx context.Context, matchID int) ([]ClubMatch, error)
	FindByID(ctx context.Context, id int) (ClubMatch, error)
	UpdateScore(ctx context.Context, id int, score int) error
	UpdateDataStatus(ctx context.Context, id int, status ClubMatchDataStatus) error
	CountFinalByMatchID(ctx context.Context, matchID int) (int, error)
}
