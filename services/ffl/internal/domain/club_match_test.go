package domain

import "testing"

func pos(p Position) *Position                { return &p }
func stat(s PlayerMatchStatus) *PlayerMatchStatus { return &s }

func TestClubMatch_Score(t *testing.T) {
	tests := []struct {
		name string
		cm   ClubMatch
		want int
	}{
		{"no players", ClubMatch{}, 0},
		{"single starter", ClubMatch{
			PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), Status: stat(PlayerMatchStatusPlayed), Score: 20},
			},
		}, 20},
		{"multiple starters", ClubMatch{
			PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), Status: stat(PlayerMatchStatusPlayed), Score: 15},
				{Position: pos(PositionKicks), Status: stat(PlayerMatchStatusPlayed), Score: 10},
				{Position: pos(PositionMarks), Status: stat(PlayerMatchStatusPlayed), Score: 25},
			},
		}, 50},
		{"bench excluded when starters play", ClubMatch{
			PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), Status: stat(PlayerMatchStatusPlayed), Score: 20},
				{Status: stat(PlayerMatchStatusPlayed), Score: 30, BackupPositions: strPtr("goals")},
			},
		}, 20},
		{"nil position skipped", ClubMatch{
			PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), Status: stat(PlayerMatchStatusPlayed), Score: 20},
				{Status: stat(PlayerMatchStatusPlayed), Score: 10},
			},
		}, 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cm.Score(); got != tt.want {
				t.Errorf("Score() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestClubMatch_Score_BenchSubstitution(t *testing.T) {
	tests := []struct {
		name string
		cm   ClubMatch
		want int
	}{
		{"bench subs for DNP starter", ClubMatch{
			PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), Status: stat(PlayerMatchStatusDNP), Score: 0},
				{Position: pos(PositionKicks), Status: stat(PlayerMatchStatusPlayed), Score: 10},
				{Status: stat(PlayerMatchStatusPlayed), Score: 12, BackupPositions: strPtr("goals")},
			},
		}, 22}, // bench (12) replaces DNP goals starter (0), kicks starter (10) stays
		{"no sub available for DNP starter", ClubMatch{
			PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), Status: stat(PlayerMatchStatusDNP), Score: 0},
				{Position: pos(PositionKicks), Status: stat(PlayerMatchStatusPlayed), Score: 10},
			},
		}, 10}, // no bench player, DNP starter contributes 0
		{"bench does not sub for played starter", ClubMatch{
			PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), Status: stat(PlayerMatchStatusPlayed), Score: 5},
				{Status: stat(PlayerMatchStatusPlayed), Score: 20, BackupPositions: strPtr("goals")},
			},
		}, 5}, // starter played, bench stays out
		{"nil status treated as non-DNP", ClubMatch{
			PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), Score: 5},
				{Status: stat(PlayerMatchStatusPlayed), Score: 20, BackupPositions: strPtr("goals")},
			},
		}, 5}, // nil status != DNP, so no sub
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cm.Score(); got != tt.want {
				t.Errorf("Score() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestClubMatch_Score_InterchangeSwap(t *testing.T) {
	tests := []struct {
		name string
		cm   ClubMatch
		want int
	}{
		{"interchange swaps when bench outscores starter", ClubMatch{
			PlayerMatches: []PlayerMatch{
				{Position: pos(PositionKicks), Status: stat(PlayerMatchStatusPlayed), Score: 8},
				{Status: stat(PlayerMatchStatusPlayed), Score: 15, InterchangePosition: strPtr("kicks")},
			},
		}, 15}, // bench (15) replaces kicks starter (8)
		{"interchange does not swap when starter outscores bench", ClubMatch{
			PlayerMatches: []PlayerMatch{
				{Position: pos(PositionKicks), Status: stat(PlayerMatchStatusPlayed), Score: 20},
				{Status: stat(PlayerMatchStatusPlayed), Score: 10, InterchangePosition: strPtr("kicks")},
			},
		}, 20}, // starter (20) stays
		{"interchange does not swap when scores equal", ClubMatch{
			PlayerMatches: []PlayerMatch{
				{Position: pos(PositionKicks), Status: stat(PlayerMatchStatusPlayed), Score: 10},
				{Status: stat(PlayerMatchStatusPlayed), Score: 10, InterchangePosition: strPtr("kicks")},
			},
		}, 10}, // no swap on tie
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cm.Score(); got != tt.want {
				t.Errorf("Score() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestClubMatch_Score_SubTakesPriorityOverInterchange(t *testing.T) {
	cm := ClubMatch{
		PlayerMatches: []PlayerMatch{
			{Position: pos(PositionGoals), Status: stat(PlayerMatchStatusDNP), Score: 0},
			{Status: stat(PlayerMatchStatusPlayed), Score: 18, BackupPositions: strPtr("goals"), InterchangePosition: strPtr("goals")},
		},
	}
	if got := cm.Score(); got != 18 {
		t.Errorf("Score() = %d, want 18", got)
	}
}

func strPtr(s string) *string { return &s }
