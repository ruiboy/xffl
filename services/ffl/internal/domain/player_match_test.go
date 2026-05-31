package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateScore(t *testing.T) {
	stats := AFLStats{
		Goals:     3,
		Kicks:     15,
		Handballs: 10,
		Marks:     6,
		Tackles:   4,
		Hitouts:   2,
	}

	tests := []struct {
		name     string
		position Position
		want     int
	}{
		{"goals position", PositionGoals, 15},         // 3 * 5
		{"kicks position", PositionKicks, 15},         // 15 * 1
		{"handballs position", PositionHandballs, 10}, // 10 * 1
		{"marks position", PositionMarks, 12},         // 6 * 2
		{"tackles position", PositionTackles, 16},     // 4 * 4
		{"hitouts position", PositionHitouts, 2},      // 2 * 1
		{"star position", PositionStar, 68},           // 3*5 + 15*1 + 10*1 + 6*2 + 4*4
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := PlayerMatch{Position: PositionPtr(tt.position)}
			assert.Equal(t, tt.want, pm.CalculateScore(stats))
		})
	}
}

func TestCalculateScore_NilPosition(t *testing.T) {
	pm := PlayerMatch{}
	assert.Equal(t, 0, pm.CalculateScore(AFLStats{Goals: 5}))
}

func TestCalculateScore_ZeroStats(t *testing.T) {
	stats := AFLStats{}
	positions := []Position{
		PositionGoals, PositionKicks, PositionHandballs,
		PositionMarks, PositionTackles, PositionHitouts, PositionStar,
	}
	for _, pos := range positions {
		t.Run(string(pos), func(t *testing.T) {
			pm := PlayerMatch{Position: PositionPtr(pos)}
			assert.Equal(t, 0, pm.CalculateScore(stats))
		})
	}
}

func TestCalculateByeScore(t *testing.T) {
	// averages with fractional parts to verify floor-per-stat behaviour
	avg := AFLAvgStats{
		Goals:     2.9, // floor → 2, ×5 = 10
		Kicks:     14.8, // floor → 14
		Handballs: 9.7, // floor → 9
		Marks:     5.6, // floor → 5, ×2 = 10
		Tackles:   3.9, // floor → 3, ×4 = 12
		Hitouts:   1.5, // floor → 1
	}

	tests := []struct {
		name     string
		position Position
		want     int
	}{
		{"goals", PositionGoals, 10},     // floor(2.9)*5
		{"kicks", PositionKicks, 14},     // floor(14.8)*1
		{"handballs", PositionHandballs, 9},  // floor(9.7)*1
		{"marks", PositionMarks, 10},     // floor(5.6)*2
		{"tackles", PositionTackles, 12}, // floor(3.9)*4
		{"hitouts", PositionHitouts, 1},  // floor(1.5)*1
		// star: floor each independently: 10+14+9+10+12 = 55
		{"star", PositionStar, 55},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := PlayerMatch{Position: PositionPtr(tt.position)}
			assert.Equal(t, tt.want, pm.CalculateByeScore(avg))
		})
	}
}

func TestCalculateByeScore_NilPosition(t *testing.T) {
	pm := PlayerMatch{}
	assert.Equal(t, 0, pm.CalculateByeScore(AFLAvgStats{Goals: 3.0}))
}
