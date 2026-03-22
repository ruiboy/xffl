package domain

import "testing"

func TestClubMatch_Score(t *testing.T) {
	tests := []struct {
		name string
		cm   ClubMatch
		want int
	}{
		{"no players no rushed", ClubMatch{}, 0},
		{"rushed behinds only", ClubMatch{RushedBehinds: 3}, 3},
		{"single player", ClubMatch{
			PlayerMatches: []PlayerMatch{{Goals: 2, Behinds: 1}},
		}, 13},
		{"multiple players with rushed", ClubMatch{
			RushedBehinds: 4,
			PlayerMatches: []PlayerMatch{
				{Goals: 3, Behinds: 2}, // 20
				{Goals: 1, Behinds: 0}, // 6
			},
		}, 30},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cm.Score(); got != tt.want {
				t.Errorf("Score() = %d, want %d", got, tt.want)
			}
		})
	}
}
