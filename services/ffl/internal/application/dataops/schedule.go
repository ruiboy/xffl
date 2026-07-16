package dataops

// Pairing is one home-vs-away match by club index (into the ordered clubs slice).
type Pairing struct {
	HomeIdx int
	AwayIdx int
}

// RoundRobinPairings generates home-and-away fixtures for an even number of
// clubs across the requested number of rounds, using the circle method. Each
// full cycle is clubCount-1 rounds (a complete single round-robin); home/away
// swap every cycle so, over an even number of cycles, each pair plays home and
// away equally. Returns nil for an odd or <2 club count.
func RoundRobinPairings(clubCount, rounds int) [][]Pairing {
	if clubCount < 2 || clubCount%2 != 0 || rounds < 1 {
		return nil
	}
	perCycle := clubCount - 1
	half := clubCount / 2

	// circle[0] stays fixed; the rest rotate one position each round.
	circle := make([]int, clubCount)
	for i := range circle {
		circle[i] = i
	}

	out := make([][]Pairing, 0, rounds)
	for r := 0; r < rounds; r++ {
		swap := (r/perCycle)%2 == 1
		round := make([]Pairing, 0, half)
		for i := 0; i < half; i++ {
			home, away := circle[i], circle[clubCount-1-i]
			if swap {
				home, away = away, home
			}
			round = append(round, Pairing{HomeIdx: home, AwayIdx: away})
		}
		out = append(out, round)

		// rotate positions 1..clubCount-1 clockwise by one.
		last := circle[clubCount-1]
		copy(circle[2:], circle[1:clubCount-1])
		circle[1] = last
	}
	return out
}
