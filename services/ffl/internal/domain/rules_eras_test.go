package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A fixed stat line scored under each era, checking the positions the era's
// deltas touch (goals, tackles, star) plus the interchange flag.
func TestRulesEras(t *testing.T) {
	stats := AFLStats{Goals: 3, Kicks: 10, Handballs: 8, Marks: 4, Tackles: 5, Hitouts: 6}

	tests := []struct {
		id          string
		goals       int // goals position
		tackles     int // tackles position
		star        int // star position
		interchange bool
	}{
		// 1998: goal 4, tackle 3, star includes hitouts, no interchange.
		{"1998", 3 * 4, 5 * 3, 3*4 + 10 + 8 + 4*2 + 5*3 + 6, false},
		// 1999: hitouts removed from the star.
		{"1999", 3 * 4, 5 * 3, 3*4 + 10 + 8 + 4*2 + 5*3, false},
		// 2005: tackle 4.
		{"2005", 3 * 4, 5 * 4, 3*4 + 10 + 8 + 4*2 + 5*4, false},
		// 2010: goal 5.
		{"2010", 3 * 5, 5 * 4, 3*5 + 10 + 8 + 4*2 + 5*4, false},
		// 2015: interchange added; scoring same as 2010.
		{"2015", 3 * 5, 5 * 4, 3*5 + 10 + 8 + 4*2 + 5*4, true},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			r, err := RulesFor(tt.id)
			require.NoError(t, err)
			assert.Equal(t, tt.goals, r.Score(PositionGoals, stats), "goals")
			assert.Equal(t, tt.tackles, r.Score(PositionTackles, stats), "tackles")
			assert.Equal(t, tt.star, r.Score(PositionStar, stats), "star")
			assert.Equal(t, tt.interchange, r.Composition.Interchange, "interchange")
			// hitouts position is unchanged across eras
			assert.Equal(t, 6, r.Score(PositionHitouts, stats), "hitouts")
		})
	}
}

// Eras are independent standalone values: the 1998 star including hitouts must
// not imply the 2015 star does.
func TestRulesEras_Isolated(t *testing.T) {
	assert.Equal(t, 5, rules2015.Scoring.Points[StatGoals], "2015 goal points")
	assert.Equal(t, 4, rules2015.Scoring.Points[StatTackles], "2015 tackle points")
	star2015, _ := rules2015.position(PositionStar)
	star1998, _ := rules1998.position(PositionStar)
	assert.NotContains(t, star2015.Stats, StatHitouts)
	assert.Contains(t, star1998.Stats, StatHitouts)
}
