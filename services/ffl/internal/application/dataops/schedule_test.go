package dataops

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoundRobinPairings_FourClubs(t *testing.T) {
	// 4 clubs, 6 rounds = two full cycles.
	sched := RoundRobinPairings(4, 6)
	require.Len(t, sched, 6)

	for _, round := range sched {
		assert.Len(t, round, 2, "4 clubs => 2 matches per round")
		// every club appears exactly once per round
		seen := map[int]bool{}
		for _, p := range round {
			assert.False(t, seen[p.HomeIdx], "club plays once per round")
			assert.False(t, seen[p.AwayIdx], "club plays once per round")
			seen[p.HomeIdx] = true
			seen[p.AwayIdx] = true
			assert.NotEqual(t, p.HomeIdx, p.AwayIdx)
		}
		assert.Len(t, seen, 4)
	}

	// Over the six rounds every unordered pair meets, and (two cycles) each
	// pair plays exactly once at home and once away.
	homeAway := map[[2]int]int{} // ordered pair -> count
	for _, round := range sched {
		for _, p := range round {
			homeAway[[2]int{p.HomeIdx, p.AwayIdx}]++
		}
	}
	pairs := map[[2]int]bool{{0, 1}: true, {0, 2}: true, {0, 3}: true, {1, 2}: true, {1, 3}: true, {2, 3}: true}
	for a := 0; a < 4; a++ {
		for b := a + 1; b < 4; b++ {
			if !pairs[[2]int{a, b}] {
				continue
			}
			total := homeAway[[2]int{a, b}] + homeAway[[2]int{b, a}]
			assert.Equalf(t, 2, total, "pair %d-%d meets twice", a, b)
			assert.Equalf(t, 1, homeAway[[2]int{a, b}], "pair %d-%d: one home", a, b)
			assert.Equalf(t, 1, homeAway[[2]int{b, a}], "pair %d-%d: one away", a, b)
		}
	}
}

func TestRoundRobinPairings_Invalid(t *testing.T) {
	assert.Nil(t, RoundRobinPairings(3, 5), "odd club count")
	assert.Nil(t, RoundRobinPairings(0, 5))
	assert.Nil(t, RoundRobinPairings(4, 0))
}
