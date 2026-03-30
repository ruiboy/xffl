package domain

import "testing"

func TestClubSeason_Percentage(t *testing.T) {
	tests := []struct {
		name    string
		cs      ClubSeason
		want    float64
	}{
		{"normal", ClubSeason{For: 1200, Against: 1000}, 120.0},
		{"equal", ClubSeason{For: 500, Against: 500}, 100.0},
		{"losing", ClubSeason{For: 800, Against: 1600}, 50.0},
		{"zero against", ClubSeason{For: 100, Against: 0}, 0},
		{"zero for and against", ClubSeason{For: 0, Against: 0}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cs.Percentage(); got != tt.want {
				t.Errorf("Percentage() = %f, want %f", got, tt.want)
			}
		})
	}
}
