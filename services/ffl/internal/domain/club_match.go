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

// Score computes the total fantasy score for this club match.
//
// Two modes:
//   - Auto mode (all starters named): substitutes all DNP starters with first eligible bench
//     player; applies interchange if bench player score exceeds the starter's.
//   - TM mode (any starter has status subbed or interchanged): uses explicit TM decisions —
//     subbed starters are covered via BackupPositions; interchanged starters are swapped with
//     the interchange bench player. Bench players stay named in both modes.
func (cm ClubMatch) Score() int {
	if cm.isTMMode() {
		return cm.scoreTM()
	}
	return cm.scoreAuto()
}

// isTMMode returns true if any starter has an explicit TM decision recorded.
func (cm ClubMatch) isTMMode() bool {
	for _, pm := range cm.PlayerMatches {
		if pm.isBench() {
			continue
		}
		if pm.Status != nil && (*pm.Status == PlayerMatchStatusSubbed || *pm.Status == PlayerMatchStatusInterchange) {
			return true
		}
	}
	return false
}

// scoreAuto substitutes all DNP starters and applies interchange where beneficial.
func (cm ClubMatch) scoreAuto() int {
	starters := make(map[Position][]*PlayerMatch)
	var bench []*PlayerMatch

	for i := range cm.PlayerMatches {
		pm := &cm.PlayerMatches[i]
		if pm.isBench() {
			bench = append(bench, pm)
		} else if pm.Position != nil {
			starters[*pm.Position] = append(starters[*pm.Position], pm)
		}
	}

	used := make(map[int]bool)

	// Substitution: replace each DNP starter with the first eligible bench player.
	for pos, slots := range starters {
		for si, starter := range slots {
			if starter.AFLStatus == nil || *starter.AFLStatus != AFLStatusDNP {
				continue
			}
			for i, bp := range bench {
				if used[i] || bp.BackupPositions == nil {
					continue
				}
				if containsPosition(*bp.BackupPositions, pos) {
					starters[pos][si] = bp
					used[i] = true
					break
				}
			}
		}
	}

	// Interchange: bench player beats a starter at their designated interchange position.
	for i, bp := range bench {
		if used[i] || bp.InterchangePosition == nil {
			continue
		}
		targetPos := Position(*bp.InterchangePosition)
		slots, ok := starters[targetPos]
		if !ok {
			continue
		}
		bestSlot := -1
		bestGain := 0
		for si, starter := range slots {
			if gain := bp.Score - starter.Score; gain > bestGain {
				bestGain = gain
				bestSlot = si
			}
		}
		if bestSlot >= 0 {
			starters[targetPos][bestSlot] = bp
			used[i] = true
		}
	}

	total := 0
	for _, slots := range starters {
		for _, pm := range slots {
			total += pm.Score
		}
	}
	return total
}

// scoreTM applies explicit TM decisions: interchanged starters swap with the interchange
// bench player; subbed starters are covered via BackupPositions.
func (cm ClubMatch) scoreTM() int {
	starters := make(map[Position][]*PlayerMatch)
	var bench []*PlayerMatch

	for i := range cm.PlayerMatches {
		pm := &cm.PlayerMatches[i]
		if pm.isBench() {
			bench = append(bench, pm)
		} else if pm.Position != nil {
			starters[*pm.Position] = append(starters[*pm.Position], pm)
		}
	}

	used := make(map[int]bool)

	// Interchange: swap the starter marked interchanged with the interchange bench player.
	for i, bp := range bench {
		if used[i] || bp.InterchangePosition == nil {
			continue
		}
		targetPos := Position(*bp.InterchangePosition)
		for si, starter := range starters[targetPos] {
			if starter.Status != nil && *starter.Status == PlayerMatchStatusInterchange {
				starters[targetPos][si] = bp
				used[i] = true
				break
			}
		}
	}

	// Substitution: replace each subbed starter via BackupPositions.
	for pos, slots := range starters {
		for si, starter := range slots {
			if starter.Status == nil || *starter.Status != PlayerMatchStatusSubbed {
				continue
			}
			for i, bp := range bench {
				if used[i] || bp.BackupPositions == nil {
					continue
				}
				if containsPosition(*bp.BackupPositions, pos) {
					starters[pos][si] = bp
					used[i] = true
					break
				}
			}
		}
	}

	total := 0
	for _, slots := range starters {
		for _, pm := range slots {
			total += pm.Score
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
