package domain

import "testing"

func TestCalculateScore(t *testing.T) {
	stats := AFLStats{
		Goals:    3,
		Kicks:    15,
		Handballs: 10,
		Marks:    6,
		Tackles:  4,
		Hitouts:  2,
	}

	tests := []struct {
		name     string
		position Position
		want     int
	}{
		{"goals position", PositionGoals, 15},      // 3 * 5
		{"kicks position", PositionKicks, 15},       // 15 * 1
		{"handballs position", PositionHandballs, 10}, // 10 * 1
		{"marks position", PositionMarks, 12},       // 6 * 2
		{"tackles position", PositionTackles, 16},   // 4 * 4
		{"hitouts position", PositionHitouts, 2},    // 2 * 1
		{"star position", PositionStar, 68},         // 3*5 + 15*1 + 10*1 + 6*2 + 4*4
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := PlayerMatch{Position: tt.position}
			if got := pm.CalculateScore(stats); got != tt.want {
				t.Errorf("CalculateScore() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCalculateScore_ZeroStats(t *testing.T) {
	stats := AFLStats{}
	positions := []Position{
		PositionGoals, PositionKicks, PositionHandballs,
		PositionMarks, PositionTackles, PositionHitouts, PositionStar,
	}
	for _, pos := range positions {
		t.Run(string(pos), func(t *testing.T) {
			pm := PlayerMatch{Position: pos}
			if got := pm.CalculateScore(stats); got != 0 {
				t.Errorf("CalculateScore() with zero stats = %d, want 0", got)
			}
		})
	}
}
