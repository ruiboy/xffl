package domain

import (
	"context"
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

// Score computes the total fantasy score for this club match.
//
// A player is a starter if they occupy a position slot (BackupPositions == nil). Multiple players can occupy the same position (e.g. 3
// goal kickers). Each starter slot is scored independently.
//
// Rules applied in order:
//  1. Substitution: if a starter's status is DNP, the first eligible bench player
//     whose BackupPositions includes that position fills the slot. A player who
//     played but scored 0 is NOT eligible for substitution.
//  2. Interchange: if a bench player's InterchangePosition targets a starter slot
//     and the bench player's score strictly exceeds that starter's, they swap.
//     For multi-slot positions, the bench player replaces the lowest-scoring
//     starter they can beat.
//
// A bench player can only be used once (sub or interchange, not both).
// Substitution is evaluated before interchange.
func (cm ClubMatch) Score() int {
	// starters holds all starter slots per position (multiple per position allowed).
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

	used := make(map[int]bool) // bench indices already consumed

	// Substitution: replace each DNP starter slot with the first eligible bench player.
	for pos, slots := range starters {
		for si, starter := range slots {
			if starter.Status == nil || *starter.Status != PlayerMatchStatusDNP {
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
		// Replace the slot where the bench player produces the greatest gain.
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

// containsPosition checks whether a comma-separated positions string contains pos.
func containsPosition(positions string, pos Position) bool {
	for _, p := range strings.Split(positions, ",") {
		if Position(strings.TrimSpace(p)) == pos {
			return true
		}
	}
	return false
}

type ClubMatchRepository interface {
	FindByMatchID(ctx context.Context, matchID int) ([]ClubMatch, error)
	FindByID(ctx context.Context, id int) (ClubMatch, error)
	UpdateScore(ctx context.Context, id int, score int) error
	UpdateDataStatus(ctx context.Context, id int, status ClubMatchDataStatus) error
}
