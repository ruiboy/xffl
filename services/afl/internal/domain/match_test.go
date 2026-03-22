package domain

import "testing"

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
