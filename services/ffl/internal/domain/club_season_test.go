package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClubSeason_Percentage(t *testing.T) {
	tests := []struct {
		name string
		cs   ClubSeason
		want float64
	}{
		{"returns correct percentage when for exceeds against", ClubSeason{For: 1200, Against: 1000}, 120.0},
		{"returns 100 when for equals against", ClubSeason{For: 500, Against: 500}, 100.0},
		{"returns 50 when against is double for", ClubSeason{For: 800, Against: 1600}, 50.0},
		{"returns 0 when against is zero", ClubSeason{For: 100, Against: 0}, 0},
		{"returns 0 when both for and against are zero", ClubSeason{For: 0, Against: 0}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cs.Percentage())
		})
	}
}
