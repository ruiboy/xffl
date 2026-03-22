package domain

import "testing"

func TestPlayerMatch_Disposals(t *testing.T) {
	tests := []struct {
		name      string
		kicks     int
		handballs int
		want      int
	}{
		{"zero stats", 0, 0, 0},
		{"kicks only", 10, 0, 10},
		{"handballs only", 0, 7, 7},
		{"mixed", 12, 8, 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := PlayerMatch{Kicks: tt.kicks, Handballs: tt.handballs}
			if got := pm.Disposals(); got != tt.want {
				t.Errorf("Disposals() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPlayerMatch_Score(t *testing.T) {
	tests := []struct {
		name    string
		goals   int
		behinds int
		want    int
	}{
		{"zero", 0, 0, 0},
		{"goals only", 3, 0, 18},
		{"behinds only", 0, 5, 5},
		{"mixed", 2, 3, 15},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := PlayerMatch{Goals: tt.goals, Behinds: tt.behinds}
			if got := pm.Score(); got != tt.want {
				t.Errorf("Score() = %d, want %d", got, tt.want)
			}
		})
	}
}

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

func TestMatch_Winner(t *testing.T) {
	tests := []struct {
		name     string
		match    Match
		wantHome bool
		wantDraw bool
	}{
		{
			"home wins",
			Match{
				Home: ClubMatch{PlayerMatches: []PlayerMatch{{Goals: 3}}},
				Away: ClubMatch{PlayerMatches: []PlayerMatch{{Goals: 1}}},
			},
			true, false,
		},
		{
			"away wins",
			Match{
				Home: ClubMatch{PlayerMatches: []PlayerMatch{{Goals: 1}}},
				Away: ClubMatch{PlayerMatches: []PlayerMatch{{Goals: 3}}},
			},
			false, false,
		},
		{
			"draw",
			Match{
				Home: ClubMatch{PlayerMatches: []PlayerMatch{{Goals: 2}}},
				Away: ClubMatch{PlayerMatches: []PlayerMatch{{Goals: 2}}},
			},
			false, true,
		},
		{
			"no players is a draw",
			Match{},
			false, true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			winner := tt.match.Winner()
			if tt.wantDraw {
				if winner != nil {
					t.Error("expected draw (nil winner)")
				}
				return
			}
			if winner == nil {
				t.Fatal("expected a winner, got nil")
			}
			if tt.wantHome && winner != &tt.match.Home {
				t.Error("expected home to win")
			}
			if !tt.wantHome && winner != &tt.match.Away {
				t.Error("expected away to win")
			}
		})
	}
}
