package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatch_DeriveResult(t *testing.T) {
	tests := []struct {
		name string
		home int
		away int
		want MatchResult
	}{
		{"home win", 1200, 1000, MatchResultHomeWin},
		{"away win", 900, 1100, MatchResultAwayWin},
		{"draw", 1000, 1000, MatchResultDraw},
		{"zero scores draw", 0, 0, MatchResultDraw},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Match{
				Home: ClubMatch{StoredScore: tt.home},
				Away: ClubMatch{StoredScore: tt.away},
			}
			assert.Equal(t, tt.want, m.DeriveResult())
		})
	}
}
