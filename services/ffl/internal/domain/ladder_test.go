package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateLadder(t *testing.T) {
	tests := []struct {
		name    string
		matches []Match
		want    map[int]ClubSeason
	}{
		{
			name: "home win",
			matches: []Match{
				{Home: ClubMatch{ClubSeasonID: 1, StoredScore: 1200}, Away: ClubMatch{ClubSeasonID: 2, StoredScore: 1000}},
			},
			want: map[int]ClubSeason{
				1: {ID: 1, Played: 1, Won: 1, For: 1200, Against: 1000, PremiershipPoints: 4},
				2: {ID: 2, Played: 1, Lost: 1, For: 1000, Against: 1200},
			},
		},
		{
			name: "away win",
			matches: []Match{
				{Home: ClubMatch{ClubSeasonID: 1, StoredScore: 800}, Away: ClubMatch{ClubSeasonID: 2, StoredScore: 1100}},
			},
			want: map[int]ClubSeason{
				1: {ID: 1, Played: 1, Lost: 1, For: 800, Against: 1100},
				2: {ID: 2, Played: 1, Won: 1, For: 1100, Against: 800, PremiershipPoints: 4},
			},
		},
		{
			name: "draw",
			matches: []Match{
				{Home: ClubMatch{ClubSeasonID: 1, StoredScore: 1000}, Away: ClubMatch{ClubSeasonID: 2, StoredScore: 1000}},
			},
			want: map[int]ClubSeason{
				1: {ID: 1, Played: 1, Drawn: 1, For: 1000, Against: 1000, PremiershipPoints: 2},
				2: {ID: 2, Played: 1, Drawn: 1, For: 1000, Against: 1000, PremiershipPoints: 2},
			},
		},
		{
			name: "multiple matches accumulate",
			matches: []Match{
				{Home: ClubMatch{ClubSeasonID: 1, StoredScore: 1200}, Away: ClubMatch{ClubSeasonID: 2, StoredScore: 1000}},
				{Home: ClubMatch{ClubSeasonID: 1, StoredScore: 900}, Away: ClubMatch{ClubSeasonID: 3, StoredScore: 950}},
			},
			want: map[int]ClubSeason{
				1: {ID: 1, Played: 2, Won: 1, Lost: 1, For: 2100, Against: 1950, PremiershipPoints: 4},
				2: {ID: 2, Played: 1, Lost: 1, For: 1000, Against: 1200},
				3: {ID: 3, Played: 1, Won: 1, For: 950, Against: 900, PremiershipPoints: 4},
			},
		},
		{
			name: "skips match with missing ClubSeasonID",
			matches: []Match{
				{Home: ClubMatch{ClubSeasonID: 0, StoredScore: 1200}, Away: ClubMatch{ClubSeasonID: 2, StoredScore: 1000}},
				{Home: ClubMatch{ClubSeasonID: 1, StoredScore: 900}, Away: ClubMatch{ClubSeasonID: 0, StoredScore: 950}},
			},
			want: map[int]ClubSeason{},
		},
		{
			name:    "empty matches",
			matches: []Match{},
			want:    map[int]ClubSeason{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateLadder(tt.matches)
			assert.Equal(t, tt.want, got)
		})
	}
}
