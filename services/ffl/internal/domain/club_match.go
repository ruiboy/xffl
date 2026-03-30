package domain

import (
	"context"
	"strings"
)

type ClubMatch struct {
	ID            int
	MatchID       int
	ClubSeasonID  int
	StoredScore   int
	PlayerMatches []PlayerMatch
}

// Score computes the total fantasy score for this club match.
//
// A player is a starter if they occupy a position slot (no BackupPositions or
// InterchangePosition). A player is on the bench if they have either field set.
//
// Rules:
//  1. Start with the sum of all starters' scores.
//  2. Substitution: if a starter has status DNP, a bench player whose
//     BackupPositions includes the starter's position fills in.
//  3. Interchange: if a bench player's InterchangePosition targets a starter
//     and the bench player's score exceeds the starter's, swap them.
//
// A bench player can only be used once (sub or interchange, not both).
// Substitution is evaluated before interchange.
func (cm ClubMatch) Score() int {
	starters := make(map[Position]*PlayerMatch)
	var bench []*PlayerMatch

	for i := range cm.PlayerMatches {
		pm := &cm.PlayerMatches[i]
		if pm.isBench() {
			bench = append(bench, pm)
		} else {
			starters[pm.Position] = pm
		}
	}

	used := make(map[int]bool) // bench player indices already consumed

	// Substitution: bench replaces DNP starters.
	for pos, starter := range starters {
		if starter.Status != PlayerMatchStatusDNP {
			continue
		}
		for i, bp := range bench {
			if used[i] || bp.BackupPositions == nil {
				continue
			}
			if containsPosition(*bp.BackupPositions, pos) {
				starters[pos] = bp
				used[i] = true
				break
			}
		}
	}

	// Interchange: bench outscores starter at the interchange position.
	for i, bp := range bench {
		if used[i] || bp.InterchangePosition == nil {
			continue
		}
		targetPos := Position(*bp.InterchangePosition)
		starter, ok := starters[targetPos]
		if !ok {
			continue
		}
		if bp.Score > starter.Score {
			starters[targetPos] = bp
			used[i] = true
		}
	}

	total := 0
	for _, pm := range starters {
		total += pm.Score
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
}
