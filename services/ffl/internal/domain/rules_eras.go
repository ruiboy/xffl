package domain

// Historical rules, one per scoring era. Each era is a standalone value — no
// era is derived from another, so editing one can never silently shift another.
// What has actually varied across eras is small (goal/tackle points, the star's
// stat-set, interchange), so the parts that have *never* varied are supplied by
// the constructors below rather than copy-pasted into every era.
//
//	2015 — goal 5, tackle 4, star excludes hitouts, interchange
//	2010 — goal 5, tackle 4, star excludes hitouts, no interchange
//	2005 — goal 4, tackle 4, star excludes hitouts, no interchange
//	1999 — goal 4, tackle 3, star excludes hitouts, no interchange
//	1998 — goal 4, tackle 3, star includes hitouts, no interchange
//
// Years are era-start labels; which season uses which era is data
// (ffl.season.rules_id), so exact boundaries can be pinned later.

// statPoints builds the per-stat point map. Only goals and tackles have varied
// across eras; kicks, handballs, marks, and hitouts have always been 1, 1, 2, 1.
func statPoints(goals, tackles int) map[Stat]int {
	return map[Stat]int{
		StatGoals:     goals,
		StatKicks:     1,
		StatHandballs: 1,
		StatMarks:     2,
		StatTackles:   tackles,
		StatHitouts:   1,
	}
}

// standardPositions builds the position list shared by all known eras. The six
// single-stat positions and their slot counts have never varied; only the star's
// stat-set has (it once included hitouts), so that is the sole parameter. A fresh
// slice is returned each call, so eras never alias one another.
func standardPositions(starStats []Stat) []PositionRule {
	return []PositionRule{
		{Position: PositionGoals, Stats: []Stat{StatGoals}, Slots: 3},
		{Position: PositionKicks, Stats: []Stat{StatKicks}, Slots: 4},
		{Position: PositionHandballs, Stats: []Stat{StatHandballs}, Slots: 4},
		{Position: PositionMarks, Stats: []Stat{StatMarks}, Slots: 2},
		{Position: PositionTackles, Stats: []Stat{StatTackles}, Slots: 2},
		{Position: PositionHitouts, Stats: []Stat{StatHitouts}, Slots: 2},
		{Position: PositionStar, Stats: starStats, Slots: 1},
	}
}

// The star scores every counting stat; before 1999 it also scored hitouts.
func starStats() []Stat {
	return []Stat{StatGoals, StatKicks, StatHandballs, StatMarks, StatTackles}
}
func starStatsWithHitouts() []Stat {
	return []Stat{StatGoals, StatKicks, StatHandballs, StatMarks, StatTackles, StatHitouts}
}

var (
	rules2015 = Rules{
		ID:          "2015",
		Scoring:     Scoring{Points: statPoints(5, 4)},
		Composition: Composition{Positions: standardPositions(starStats()), BenchSize: 4, Interchange: true},
	}
	rules2010 = Rules{
		ID:          "2010",
		Scoring:     Scoring{Points: statPoints(5, 4)},
		Composition: Composition{Positions: standardPositions(starStats()), BenchSize: 4, Interchange: false},
	}
	rules2005 = Rules{
		ID:          "2005",
		Scoring:     Scoring{Points: statPoints(4, 4)},
		Composition: Composition{Positions: standardPositions(starStats()), BenchSize: 4, Interchange: false},
	}
	rules1999 = Rules{
		ID:          "1999",
		Scoring:     Scoring{Points: statPoints(4, 3)},
		Composition: Composition{Positions: standardPositions(starStats()), BenchSize: 4, Interchange: false},
	}
	rules1998 = Rules{
		ID:          "1998",
		Scoring:     Scoring{Points: statPoints(4, 3)},
		Composition: Composition{Positions: standardPositions(starStatsWithHitouts()), BenchSize: 4, Interchange: false},
	}
)

func init() {
	for _, r := range []Rules{rules1998, rules1999, rules2005, rules2010, rules2015} {
		rulesByID[r.ID] = r
	}
}
