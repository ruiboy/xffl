package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRules2011_Score(t *testing.T) {
	stats := AFLStats{Goals: 3, Kicks: 10, Handballs: 8, Marks: 4, Tackles: 5, Hitouts: 6}
	tests := []struct {
		pos  Position
		want int
	}{
		{PositionGoals, 3 * 5},
		{PositionKicks, 10 * 1},
		{PositionHandballs, 8 * 1},
		{PositionMarks, 4 * 2},
		{PositionTackles, 5 * 4},
		{PositionHitouts, 6 * 1},
		// star: goals*5 + kicks + handballs + marks*2 + tackles*4 (no hitouts)
		{PositionStar, 3*5 + 10 + 8 + 4*2 + 5*4},
	}
	for _, tt := range tests {
		t.Run(string(tt.pos), func(t *testing.T) {
			assert.Equal(t, tt.want, rules2011.Score(tt.pos, stats))
		})
	}
	assert.Equal(t, 0, rules2011.Score(Position("nope"), stats), "unknown position scores 0")
}

func TestRules2011_ByeScore(t *testing.T) {
	// fractional averages verify per-stat floor-before-multiply behaviour
	avg := AFLAvgStats{Goals: 3.9, Kicks: 10.2, Handballs: 8.8, Marks: 4.5, Tackles: 5.1, Hitouts: 6.7}
	tests := []struct {
		pos  Position
		want int
	}{
		{PositionGoals, 3 * 5},
		{PositionKicks, 10 * 1},
		{PositionHandballs, 8 * 1},
		{PositionMarks, 4 * 2},
		{PositionTackles, 5 * 4},
		{PositionHitouts, 6 * 1},
		// star floors each component, then sums: 3*5 + 10 + 8 + 4*2 + 5*4
		{PositionStar, 3*5 + 10 + 8 + 4*2 + 5*4},
	}
	for _, tt := range tests {
		t.Run(string(tt.pos), func(t *testing.T) {
			assert.Equal(t, tt.want, rules2011.ByeScore(tt.pos, avg))
		})
	}
}

// The 2011 rules' slot counts are the known composition validation relies on.
func TestRules2011_Slots(t *testing.T) {
	want := map[Position]int{
		PositionGoals:     3,
		PositionKicks:     4,
		PositionHandballs: 4,
		PositionMarks:     2,
		PositionTackles:   2,
		PositionHitouts:   2,
		PositionStar:      1,
	}
	for pos, n := range want {
		got, ok := rules2011.Slots(pos)
		require.Truef(t, ok, "2011 rules missing position %q", pos)
		assert.Equalf(t, n, got, "slot mismatch for %q", pos)
	}
	assert.Len(t, rules2011.Composition.Positions, len(want))
}

func TestRulesFor(t *testing.T) {
	t.Run("known id resolves", func(t *testing.T) {
		r, err := RulesFor("2011")
		require.NoError(t, err)
		assert.Equal(t, "2011", r.ID)
	})
	t.Run("empty id errors (no implicit fallback)", func(t *testing.T) {
		_, err := RulesFor("")
		assert.Error(t, err)
	})
	t.Run("unknown id errors", func(t *testing.T) {
		_, err := RulesFor("nope")
		assert.Error(t, err)
	})
}
