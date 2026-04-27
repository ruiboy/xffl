package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func pos(p Position) *Position                    { return &p }
func stat(s PlayerMatchStatus) *PlayerMatchStatus { return &s }

func TestClubMatch_Score(t *testing.T) {
	tests := []struct {
		name string
		cm   ClubMatch
		want int
	}{
		{"returns 0 with no players", ClubMatch{}, 0},
		{"counts a single starter score", ClubMatch{
			PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), Status: stat(PlayerMatchStatusPlayed), Score: 20},
			},
		}, 20},
		{"sums all starter scores", ClubMatch{
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
			assert.Equal(t, tt.want, tt.cm.Score())
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
			assert.Equal(t, tt.want, tt.cm.Score())
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
				{Status: stat(PlayerMatchStatusPlayed), Score: 15, BackupPositions: strPtr("kicks,handballs"), InterchangePosition: strPtr("kicks")},
			},
		}, 15}, // bench (15) replaces kicks starter (8)
		{"interchange does not swap when starter outscores bench", ClubMatch{
			PlayerMatches: []PlayerMatch{
				{Position: pos(PositionKicks), Status: stat(PlayerMatchStatusPlayed), Score: 20},
				{Status: stat(PlayerMatchStatusPlayed), Score: 10, BackupPositions: strPtr("kicks,handballs"), InterchangePosition: strPtr("kicks")},
			},
		}, 20}, // starter (20) stays
		{"interchange does not swap when scores equal", ClubMatch{
			PlayerMatches: []PlayerMatch{
				{Position: pos(PositionKicks), Status: stat(PlayerMatchStatusPlayed), Score: 10},
				{Status: stat(PlayerMatchStatusPlayed), Score: 10, BackupPositions: strPtr("kicks,handballs"), InterchangePosition: strPtr("kicks")},
			},
		}, 10}, // no swap on tie
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cm.Score())
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
	assert.Equal(t, 18, cm.Score())
}

func TestClubMatch_Score_MultipleStartersPerPosition(t *testing.T) {
	tests := []struct {
		name string
		cm   ClubMatch
		want int
	}{
		{
			name: "3 goal kickers each score independently",
			cm: ClubMatch{
				PlayerMatches: []PlayerMatch{
					{Position: pos(PositionGoals), Status: stat(PlayerMatchStatusPlayed), Score: 10},
					{Position: pos(PositionGoals), Status: stat(PlayerMatchStatusPlayed), Score: 15},
					{Position: pos(PositionGoals), Status: stat(PlayerMatchStatusPlayed), Score: 20},
				},
			},
			want: 45,
		},
		{
			name: "bench subs for one DNP slot, other goal kicker keeps scoring",
			cm: ClubMatch{
				PlayerMatches: []PlayerMatch{
					{Position: pos(PositionGoals), Status: stat(PlayerMatchStatusPlayed), Score: 20},
					{Position: pos(PositionGoals), Status: stat(PlayerMatchStatusDNP), Score: 0},
					{Status: stat(PlayerMatchStatusPlayed), Score: 12, BackupPositions: strPtr("goals")},
				},
			},
			want: 32, // 20 (starter) + 12 (sub for DNP)
		},
		{
			name: "bench only subs for one slot even if multiple DNP",
			cm: ClubMatch{
				PlayerMatches: []PlayerMatch{
					{Position: pos(PositionGoals), Status: stat(PlayerMatchStatusDNP), Score: 0},
					{Position: pos(PositionGoals), Status: stat(PlayerMatchStatusDNP), Score: 0},
					{Status: stat(PlayerMatchStatusPlayed), Score: 12, BackupPositions: strPtr("goals")},
				},
			},
			want: 12, // only one bench player, fills first DNP slot; second DNP remains 0
		},
		{
			name: "interchange picks best gain across multiple slots",
			cm: ClubMatch{
				PlayerMatches: []PlayerMatch{
					{Position: pos(PositionKicks), Status: stat(PlayerMatchStatusPlayed), Score: 5},
					{Position: pos(PositionKicks), Status: stat(PlayerMatchStatusPlayed), Score: 20},
					{Status: stat(PlayerMatchStatusPlayed), Score: 15, BackupPositions: strPtr("kicks,handballs"), InterchangePosition: strPtr("kicks")},
				},
			},
			want: 35, // bench (15) replaces the 5-score slot; 20-score slot unaffected
		},
		{
			name: "interchange does not apply when bench does not outscore any slot",
			cm: ClubMatch{
				PlayerMatches: []PlayerMatch{
					{Position: pos(PositionKicks), Status: stat(PlayerMatchStatusPlayed), Score: 20},
					{Position: pos(PositionKicks), Status: stat(PlayerMatchStatusPlayed), Score: 25},
					{Status: stat(PlayerMatchStatusPlayed), Score: 15, BackupPositions: strPtr("kicks,handballs"), InterchangePosition: strPtr("kicks")},
				},
			},
			want: 45, // bench (15) beats neither starter
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cm.Score())
		})
	}
}

func strPtr(s string) *string { return &s }
