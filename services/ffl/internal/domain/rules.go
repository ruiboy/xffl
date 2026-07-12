package domain

import (
	"fmt"
	"math"
)

// Rules is a season's complete set of rules, selected via ffl.season.rules_id.
// It composes independent facets; today scoring and team composition, with room
// for others (e.g. draft, salary cap, trading) as they are needed. What a Rules
// applies to is up to whatever holds its ID.
type Rules struct {
	ID          string
	Scoring     Scoring
	Composition Composition
}

// Scoring is the facet governing how AFL stats convert to fantasy points.
type Scoring struct {
	Points map[Stat]int // points awarded per unit of each stat
}

// Composition is the facet governing how a team may be built: the available
// positions (and what each scores), how many bench players, and whether the
// season has an interchange slot.
type Composition struct {
	Positions   []PositionRule
	BenchSize   int
	Interchange bool
}

// PositionRule describes one fantasy position: how many starter slots it has and
// which AFL stats its score sums. Single-stat positions (goals, kicks, …) list
// one stat; the star lists several.
type PositionRule struct {
	Position Position
	Stats    []Stat
	Slots    int
}

// position returns the PositionRule for pos, if these rules include it.
func (r Rules) position(pos Position) (PositionRule, bool) {
	for _, p := range r.Composition.Positions {
		if p.Position == pos {
			return p, true
		}
	}
	return PositionRule{}, false
}

// Slots returns the starter slot count for pos and whether pos exists in these rules.
func (r Rules) Slots(pos Position) (int, bool) {
	p, ok := r.position(pos)
	if !ok {
		return 0, false
	}
	return p.Slots, true
}

// Score computes the fantasy score for a player at pos given their AFL stats:
// the sum, over the position's stats, of stat value × the rules' point value.
func (r Rules) Score(pos Position, s AFLStats) int {
	p, ok := r.position(pos)
	if !ok {
		return 0
	}
	total := 0
	for _, st := range p.Stats {
		total += s.Value(st) * r.Scoring.Points[st]
	}
	return total
}

// ByeScore computes the bye score from season-average stats. Each stat is floored
// individually before its point value is applied, then summed (matching the
// per-component floor of the live bye calculation).
func (r Rules) ByeScore(pos Position, avg AFLAvgStats) int {
	p, ok := r.position(pos)
	if !ok {
		return 0
	}
	total := 0
	for _, st := range p.Stats {
		total += int(math.Floor(avg.Value(st))) * r.Scoring.Points[st]
	}
	return total
}

// Validate enforces the rules' team composition rules against a set of player
// matches. It returns a descriptive error if any rule is violated, or nil if the
// team is valid. Teams need not be full — all constraints are upper bounds.
func (r Rules) Validate(entries []PlayerMatch) error {
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

	// Rule 1: starter count per position ≤ the rules' slot count for that position.
	for pos, count := range starterCounts {
		max, ok := r.Slots(pos)
		if !ok {
			return fmt.Errorf("team: unknown position %q", pos)
		}
		if count > max {
			return fmt.Errorf("team: position %q has %d players, maximum is %d", pos, count, max)
		}
	}

	// Rule 2: total bench ≤ the rules' bench size.
	if len(benchPlayers) > r.Composition.BenchSize {
		return fmt.Errorf("team: bench has %d players, maximum is %d", len(benchPlayers), r.Composition.BenchSize)
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
				if _, ok := r.Slots(pos); !ok {
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

	// Rule 6: interchange positions are capped by the rules — 1 when interchange
	// is enabled, 0 when the season has no interchange.
	maxInterchange := 0
	if r.Composition.Interchange {
		maxInterchange = 1
	}
	if interchangeCount > maxInterchange {
		if !r.Composition.Interchange {
			return fmt.Errorf("team: interchange is not permitted in this season")
		}
		return fmt.Errorf("team: at most %d interchange position allowed, got %d", maxInterchange, interchangeCount)
	}

	// Rule 7: interchange position must be a recognised Position and one of the player's own backup positions.
	for _, bp := range benchPlayers {
		if bp.InterchangePosition != nil {
			pos := Position(*bp.InterchangePosition)
			if _, ok := r.Slots(pos); !ok {
				return fmt.Errorf("team: interchange position %q is not a valid position", pos)
			}
			if !containsPosition(*bp.BackupPositions, pos) {
				return fmt.Errorf("team: interchange position %q is not one of this player's backup positions", pos)
			}
		}
	}

	return nil
}

// rulesByID is the registry of era rules keyed by ID, populated from the
// full per-era definitions in rules_eras.go.
var rulesByID = map[string]Rules{}

// RulesFor returns the rules for an ID. An unknown ID is an error; seasons
// always carry a valid rules_id (the column is NOT NULL with a default), so
// there is no implicit fallback.
func RulesFor(id string) (Rules, error) {
	r, ok := rulesByID[id]
	if !ok {
		return Rules{}, fmt.Errorf("unknown rules id %q", id)
	}
	return r, nil
}
