package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatch_Winner(t *testing.T) {
	tests := []struct {
		name     string
		match    Match
		wantHome bool
		wantDraw bool
	}{
		{
			"home team wins when they have a higher score",
			Match{
				Home: ClubMatch{PlayerMatches: []PlayerMatch{{Goals: 3}}},
				Away: ClubMatch{PlayerMatches: []PlayerMatch{{Goals: 1}}},
			},
			true, false,
		},
		{
			"away team wins when they have a higher score",
			Match{
				Home: ClubMatch{PlayerMatches: []PlayerMatch{{Goals: 1}}},
				Away: ClubMatch{PlayerMatches: []PlayerMatch{{Goals: 3}}},
			},
			false, false,
		},
		{
			"match is a draw when both teams score equally",
			Match{
				Home: ClubMatch{PlayerMatches: []PlayerMatch{{Goals: 2}}},
				Away: ClubMatch{PlayerMatches: []PlayerMatch{{Goals: 2}}},
			},
			false, true,
		},
		{
			"empty match is a draw",
			Match{},
			false, true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			winner := tt.match.Winner()
			if tt.wantDraw {
				assert.Nil(t, winner)
				return
			}
			require.NotNil(t, winner)
			if tt.wantHome {
				assert.Equal(t, &tt.match.Home, winner)
			} else {
				assert.Equal(t, &tt.match.Away, winner)
			}
		})
	}
}
