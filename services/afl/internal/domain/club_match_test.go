package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClubMatch_Score(t *testing.T) {
	tests := []struct {
		name string
		cm   ClubMatch
		want int
	}{
		{"empty match scores zero", ClubMatch{}, 0},
		{"rushed behinds contribute to score without players", ClubMatch{RushedBehinds: 3}, 3},
		{"single player goals and behinds are summed correctly", ClubMatch{
			PlayerMatches: []PlayerMatch{{Goals: 2, Behinds: 1}},
		}, 13},
		{"multiple player scores and rushed behinds are all included", ClubMatch{
			RushedBehinds: 4,
			PlayerMatches: []PlayerMatch{
				{Goals: 3, Behinds: 2},
				{Goals: 1, Behinds: 0},
			},
		}, 30},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cm.Score())
		})
	}
}
