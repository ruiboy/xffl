package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// versus builds the two scored club_matches of a home-vs-away match.
func versus(matchID, csHome, scoreHome, csAway, scoreAway int, rt RoundType) []ScoredClubMatch {
	return []ScoredClubMatch{
		{MatchID: matchID, ClubMatchID: matchID*10 + 1, ClubSeasonID: csHome, Score: scoreHome, Style: MatchStyleVersus, RoundType: rt},
		{MatchID: matchID, ClubMatchID: matchID*10 + 2, ClubSeasonID: csAway, Score: scoreAway, Style: MatchStyleVersus, RoundType: rt},
	}
}

func TestCalculateLadder(t *testing.T) {
	tests := []struct {
		name    string
		results []ScoredClubMatch
		want    map[int]ClubSeason
	}{
		{
			name:    "home win",
			results: versus(1, 1, 1200, 2, 1000, RoundTypeMinor),
			want: map[int]ClubSeason{
				1: {ID: 1, Played: 1, Won: 1, For: 1200, Against: 1000, PremiershipPoints: 4},
				2: {ID: 2, Played: 1, Lost: 1, For: 1000, Against: 1200},
			},
		},
		{
			name:    "away win",
			results: versus(1, 1, 800, 2, 1100, RoundTypeMinor),
			want: map[int]ClubSeason{
				1: {ID: 1, Played: 1, Lost: 1, For: 800, Against: 1100},
				2: {ID: 2, Played: 1, Won: 1, For: 1100, Against: 800, PremiershipPoints: 4},
			},
		},
		{
			name:    "draw",
			results: versus(1, 1, 1000, 2, 1000, RoundTypeMinor),
			want: map[int]ClubSeason{
				1: {ID: 1, Played: 1, Drawn: 1, For: 1000, Against: 1000, PremiershipPoints: 2},
				2: {ID: 2, Played: 1, Drawn: 1, For: 1000, Against: 1000, PremiershipPoints: 2},
			},
		},
		{
			name:    "multiple matches accumulate",
			results: append(versus(1, 1, 1200, 2, 1000, RoundTypeMinor), versus(2, 1, 900, 3, 950, RoundTypeMinor)...),
			want: map[int]ClubSeason{
				1: {ID: 1, Played: 2, Won: 1, Lost: 1, For: 2100, Against: 1950, PremiershipPoints: 4},
				2: {ID: 2, Played: 1, Lost: 1, For: 1000, Against: 1200},
				3: {ID: 3, Played: 1, Won: 1, For: 950, Against: 900, PremiershipPoints: 4},
			},
		},
		{
			name: "versus match with only one final side is not counted",
			results: []ScoredClubMatch{
				{MatchID: 1, ClubMatchID: 11, ClubSeasonID: 1, Score: 1200, Style: MatchStyleVersus, RoundType: RoundTypeMinor},
			},
			want: map[int]ClubSeason{},
		},
		{
			name:    "grand final is excluded from the ladder",
			results: append(versus(1, 1, 1200, 2, 1000, RoundTypeMinor), versus(2, 1, 1200, 2, 1000, RoundTypeGrandFinal)...),
			want: map[int]ClubSeason{
				1: {ID: 1, Played: 1, Won: 1, For: 1200, Against: 1000, PremiershipPoints: 4},
				2: {ID: 2, Played: 1, Lost: 1, For: 1000, Against: 1200},
			},
		},
		{
			name:    "empty",
			results: nil,
			want:    map[int]ClubSeason{},
		},
		{
			name: "scoring bye adds to For only — not played, no premiership points",
			results: append(versus(1, 1, 1200, 2, 1000, RoundTypeMinor),
				ScoredClubMatch{MatchID: 2, ClubMatchID: 21, ClubSeasonID: 3, Score: 850, Style: MatchStyleBye, RoundType: RoundTypeMinor}),
			want: map[int]ClubSeason{
				1: {ID: 1, Played: 1, Won: 1, For: 1200, Against: 1000, PremiershipPoints: 4},
				2: {ID: 2, Played: 1, Lost: 1, For: 1000, Against: 1200},
				3: {ID: 3, For: 850},
			},
		},
		{
			name: "bye adds only For onto the club's head-to-head standings",
			results: append(versus(1, 1, 1000, 2, 900, RoundTypeMinor),
				ScoredClubMatch{MatchID: 2, ClubMatchID: 21, ClubSeasonID: 1, Score: 800, Style: MatchStyleBye, RoundType: RoundTypeMinor}),
			want: map[int]ClubSeason{
				1: {ID: 1, Played: 1, Won: 1, For: 1800, Against: 900, PremiershipPoints: 4},
				2: {ID: 2, Played: 1, Lost: 1, For: 900, Against: 1000},
			},
		},
		{
			name: "grand-final bye is excluded from the ladder",
			results: []ScoredClubMatch{
				{MatchID: 1, ClubMatchID: 11, ClubSeasonID: 3, Score: 850, Style: MatchStyleBye, RoundType: RoundTypeGrandFinal},
			},
			want: map[int]ClubSeason{},
		},
		{
			name: "superbye: every score counts to For, the top scorer earns one extra point",
			results: []ScoredClubMatch{
				{MatchID: 1, ClubMatchID: 11, ClubSeasonID: 1, Score: 900, Style: MatchStyleSuperbye, RoundType: RoundTypeMinor},
				{MatchID: 1, ClubMatchID: 12, ClubSeasonID: 2, Score: 1100, Style: MatchStyleSuperbye, RoundType: RoundTypeMinor},
				{MatchID: 1, ClubMatchID: 13, ClubSeasonID: 3, Score: 800, Style: MatchStyleSuperbye, RoundType: RoundTypeMinor},
			},
			want: map[int]ClubSeason{
				1: {ID: 1, For: 900},
				2: {ID: 2, For: 1100, ExtraPoints: 1, PremiershipPoints: 1},
				3: {ID: 3, For: 800},
			},
		},
		{
			name: "superbye: a tie for top shares the extra point",
			results: []ScoredClubMatch{
				{MatchID: 1, ClubMatchID: 11, ClubSeasonID: 1, Score: 1000, Style: MatchStyleSuperbye, RoundType: RoundTypeMinor},
				{MatchID: 1, ClubMatchID: 12, ClubSeasonID: 2, Score: 1000, Style: MatchStyleSuperbye, RoundType: RoundTypeMinor},
				{MatchID: 1, ClubMatchID: 13, ClubSeasonID: 3, Score: 500, Style: MatchStyleSuperbye, RoundType: RoundTypeMinor},
			},
			want: map[int]ClubSeason{
				1: {ID: 1, For: 1000, ExtraPoints: 1, PremiershipPoints: 1},
				2: {ID: 2, For: 1000, ExtraPoints: 1, PremiershipPoints: 1},
				3: {ID: 3, For: 500},
			},
		},
		{
			name: "superbye: no team scored, so no extra point is awarded",
			results: []ScoredClubMatch{
				{MatchID: 1, ClubMatchID: 11, ClubSeasonID: 1, Score: 0, Style: MatchStyleSuperbye, RoundType: RoundTypeMinor},
				{MatchID: 1, ClubMatchID: 12, ClubSeasonID: 2, Score: 0, Style: MatchStyleSuperbye, RoundType: RoundTypeMinor},
			},
			want: map[int]ClubSeason{
				1: {ID: 1},
				2: {ID: 2},
			},
		},
		{
			name: "superbye adds its extra point on top of head-to-head points",
			results: append(versus(1, 1, 1000, 2, 900, RoundTypeMinor),
				ScoredClubMatch{MatchID: 2, ClubMatchID: 21, ClubSeasonID: 1, Score: 1200, Style: MatchStyleSuperbye, RoundType: RoundTypeMinor},
				ScoredClubMatch{MatchID: 2, ClubMatchID: 22, ClubSeasonID: 2, Score: 800, Style: MatchStyleSuperbye, RoundType: RoundTypeMinor}),
			want: map[int]ClubSeason{
				1: {ID: 1, Played: 1, Won: 1, For: 2200, Against: 900, ExtraPoints: 1, PremiershipPoints: 5},
				2: {ID: 2, Played: 1, Lost: 1, For: 1700, Against: 1000},
			},
		},
		{
			name: "grand-final superbye is excluded from the ladder",
			results: []ScoredClubMatch{
				{MatchID: 1, ClubMatchID: 11, ClubSeasonID: 1, Score: 1200, Style: MatchStyleSuperbye, RoundType: RoundTypeGrandFinal},
				{MatchID: 1, ClubMatchID: 12, ClubSeasonID: 2, Score: 800, Style: MatchStyleSuperbye, RoundType: RoundTypeGrandFinal},
			},
			want: map[int]ClubSeason{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, CalculateLadder(tt.results))
		})
	}
}

// ClubMatchPremiershipPoints awards the per-club_match points the ladder writes back.
func TestClubMatchPremiershipPoints(t *testing.T) {
	t.Run("versus winner gets 4, loser 0", func(t *testing.T) {
		got := ClubMatchPremiershipPoints(versus(1, 1, 1200, 2, 1000, RoundTypeMinor))
		assert.Equal(t, map[int]int{11: 4, 12: 0}, got)
	})
	t.Run("versus draw gives both 2", func(t *testing.T) {
		got := ClubMatchPremiershipPoints(versus(1, 1, 1000, 2, 1000, RoundTypeMinor))
		assert.Equal(t, map[int]int{11: 2, 12: 2}, got)
	})
	t.Run("superbye top scorer gets 1, rest 0", func(t *testing.T) {
		got := ClubMatchPremiershipPoints([]ScoredClubMatch{
			{MatchID: 1, ClubMatchID: 11, ClubSeasonID: 1, Score: 900, Style: MatchStyleSuperbye, RoundType: RoundTypeMinor},
			{MatchID: 1, ClubMatchID: 12, ClubSeasonID: 2, Score: 1100, Style: MatchStyleSuperbye, RoundType: RoundTypeMinor},
		})
		assert.Equal(t, map[int]int{11: 0, 12: 1}, got)
	})
	t.Run("bye earns nothing", func(t *testing.T) {
		got := ClubMatchPremiershipPoints([]ScoredClubMatch{
			{MatchID: 1, ClubMatchID: 11, ClubSeasonID: 1, Score: 900, Style: MatchStyleBye, RoundType: RoundTypeMinor},
		})
		assert.Equal(t, map[int]int{11: 0}, got)
	})
}
