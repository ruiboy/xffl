package domain

// Stat is a raw AFL statistic that contributes to a fantasy score. It is the one
// canonical enum for the six counting stats; a position (see PositionRule) is
// defined by which stats it scores.
type Stat string

const (
	StatGoals     Stat = "goals"
	StatKicks     Stat = "kicks"
	StatHandballs Stat = "handballs"
	StatMarks     Stat = "marks"
	StatTackles   Stat = "tackles"
	StatHitouts   Stat = "hitouts"
)

// AFLStats holds the AFL performance statistics used to calculate fantasy scores.
// It is a boundary struct (named fields flow through events and GraphQL); Value
// is the single adapter from the Stat enum onto its fields.
type AFLStats struct {
	Goals     int
	Kicks     int
	Handballs int
	Marks     int
	Tackles   int
	Hitouts   int
}

// Value returns the count for a single stat.
func (s AFLStats) Value(st Stat) int {
	switch st {
	case StatGoals:
		return s.Goals
	case StatKicks:
		return s.Kicks
	case StatHandballs:
		return s.Handballs
	case StatMarks:
		return s.Marks
	case StatTackles:
		return s.Tackles
	case StatHitouts:
		return s.Hitouts
	}
	return 0
}

// AFLAvgStats holds season-average AFL statistics for bye score calculation.
// Values are raw averages (not yet floored); ByeScore floors per stat before
// multiplying.
type AFLAvgStats struct {
	Goals     float64
	Kicks     float64
	Handballs float64
	Marks     float64
	Tackles   float64
	Hitouts   float64
}

// Value returns the average for a single stat.
func (a AFLAvgStats) Value(st Stat) float64 {
	switch st {
	case StatGoals:
		return a.Goals
	case StatKicks:
		return a.Kicks
	case StatHandballs:
		return a.Handballs
	case StatMarks:
		return a.Marks
	case StatTackles:
		return a.Tackles
	case StatHitouts:
		return a.Hitouts
	}
	return 0
}
