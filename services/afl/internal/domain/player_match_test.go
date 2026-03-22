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
