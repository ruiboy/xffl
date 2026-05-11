package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pos(p Position) *Position      { return &p }
func aflSts(s AFLStatus) *AFLStatus { return &s }

func TestClubMatch_Score(t *testing.T) {
	tests := []struct {
		name string
		cm   ClubMatch
		want int
	}{
		{"returns 0 with no players", ClubMatch{}, 0},
		{"counts a single starter score", ClubMatch{
			PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), Score: 20},
			},
		}, 20},
		{"sums all starter scores", ClubMatch{
			PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), Score: 15},
				{Position: pos(PositionKicks), Score: 10},
				{Position: pos(PositionMarks), Score: 25},
			},
		}, 50},
		{"bench excluded when starters play", ClubMatch{
			PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), Score: 20},
				{Score: 30, BackupPositions: strPtr("goals")},
			},
		}, 20},
		{"nil position skipped", ClubMatch{
			PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), Score: 20},
				{Score: 10},
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
				{Position: pos(PositionGoals), AFLStatus: aflSts(AFLStatusDNP), Score: 0},
				{Position: pos(PositionKicks), Score: 10},
				{Score: 12, BackupPositions: strPtr("goals")},
			},
		}, 22}, // bench (12) replaces DNP goals starter (0), kicks starter (10) stays
		{"no sub available for DNP starter", ClubMatch{
			PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), AFLStatus: aflSts(AFLStatusDNP), Score: 0},
				{Position: pos(PositionKicks), Score: 10},
			},
		}, 10}, // no bench player, DNP starter contributes 0
		{"bench does not sub for played starter", ClubMatch{
			PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), AFLStatus: aflSts(AFLStatusPlayed), Score: 5},
				{Score: 20, BackupPositions: strPtr("goals")},
			},
		}, 5}, // starter played, bench stays out
		{"nil drv_afl_status treated as non-DNP", ClubMatch{
			PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), Score: 5},
				{Score: 20, BackupPositions: strPtr("goals")},
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
				{Position: pos(PositionKicks), Score: 8},
				{Score: 15, BackupPositions: strPtr("kicks,handballs"), InterchangePosition: strPtr("kicks")},
			},
		}, 15}, // bench (15) replaces kicks starter (8)
		{"interchange does not swap when starter outscores bench", ClubMatch{
			PlayerMatches: []PlayerMatch{
				{Position: pos(PositionKicks), Score: 20},
				{Score: 10, BackupPositions: strPtr("kicks,handballs"), InterchangePosition: strPtr("kicks")},
			},
		}, 20}, // starter (20) stays
		{"interchange does not swap when scores equal", ClubMatch{
			PlayerMatches: []PlayerMatch{
				{Position: pos(PositionKicks), Score: 10},
				{Score: 10, BackupPositions: strPtr("kicks,handballs"), InterchangePosition: strPtr("kicks")},
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
			{Position: pos(PositionGoals), AFLStatus: aflSts(AFLStatusDNP), Score: 0},
			{Score: 18, BackupPositions: strPtr("goals"), InterchangePosition: strPtr("goals")},
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
					{Position: pos(PositionGoals), Score: 10},
					{Position: pos(PositionGoals), Score: 15},
					{Position: pos(PositionGoals), Score: 20},
				},
			},
			want: 45,
		},
		{
			name: "bench subs for one DNP slot, other goal kicker keeps scoring",
			cm: ClubMatch{
				PlayerMatches: []PlayerMatch{
					{Position: pos(PositionGoals), Score: 20},
					{Position: pos(PositionGoals), AFLStatus: aflSts(AFLStatusDNP), Score: 0},
					{Score: 12, BackupPositions: strPtr("goals")},
				},
			},
			want: 32, // 20 (starter) + 12 (sub for DNP)
		},
		{
			name: "bench only subs for one slot even if multiple DNP",
			cm: ClubMatch{
				PlayerMatches: []PlayerMatch{
					{Position: pos(PositionGoals), AFLStatus: aflSts(AFLStatusDNP), Score: 0},
					{Position: pos(PositionGoals), AFLStatus: aflSts(AFLStatusDNP), Score: 0},
					{Score: 12, BackupPositions: strPtr("goals")},
				},
			},
			want: 12, // only one bench player, fills first DNP slot; second DNP remains 0
		},
		{
			name: "interchange picks best gain across multiple slots",
			cm: ClubMatch{
				PlayerMatches: []PlayerMatch{
					{Position: pos(PositionKicks), Score: 5},
					{Position: pos(PositionKicks), Score: 20},
					{Score: 15, BackupPositions: strPtr("kicks,handballs"), InterchangePosition: strPtr("kicks")},
				},
			},
			want: 35, // bench (15) replaces the 5-score slot; 20-score slot unaffected
		},
		{
			name: "interchange does not apply when bench does not outscore any slot",
			cm: ClubMatch{
				PlayerMatches: []PlayerMatch{
					{Position: pos(PositionKicks), Score: 20},
					{Position: pos(PositionKicks), Score: 25},
					{Score: 15, BackupPositions: strPtr("kicks,handballs"), InterchangePosition: strPtr("kicks")},
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
func bpPtr(s string) *string  { return &s }

// validFullTeam builds a complete 18-starter team with no bench.
func validFullTeam() []PlayerMatch {
	entries := []PlayerMatch{}
	for position, count := range PositionSlots {
		p := position
		for range count {
			entries = append(entries, PlayerMatch{Position: &p})
		}
	}
	return entries
}

func TestValidateTeam_ValidCases(t *testing.T) {
	t.Run("empty team is valid", func(t *testing.T) {
		require.NoError(t, validateTeam(nil))
	})

	t.Run("full 18-starter team is valid", func(t *testing.T) {
		require.NoError(t, validateTeam(validFullTeam()))
	})

	t.Run("starters with backup star and 3 dual-position bench", func(t *testing.T) {
		entries := validFullTeam()
		star := PositionStar
		entries = append(entries, PlayerMatch{Position: &star, BackupPositions: bpPtr("star")})
		goals := PositionGoals
		entries = append(entries, PlayerMatch{Position: &goals, BackupPositions: bpPtr("goals,kicks")})
		handballs := PositionHandballs
		entries = append(entries, PlayerMatch{Position: &handballs, BackupPositions: bpPtr("handballs,marks")})
		tackles := PositionTackles
		entries = append(entries, PlayerMatch{Position: &tackles, BackupPositions: bpPtr("tackles,hitouts")})
		require.NoError(t, validateTeam(entries))
	})

	t.Run("interchange on bench star is valid", func(t *testing.T) {
		entries := validFullTeam()
		star := PositionStar
		ic := "star"
		entries = append(entries, PlayerMatch{Position: &star, BackupPositions: bpPtr("star"), InterchangePosition: &ic})
		require.NoError(t, validateTeam(entries))
	})

	t.Run("partial team is valid", func(t *testing.T) {
		goals := PositionGoals
		entries := []PlayerMatch{
			{Position: &goals},
			{Position: &goals},
		}
		require.NoError(t, validateTeam(entries))
	})
}

func TestValidateTeam_InvalidCases(t *testing.T) {
	tests := []struct {
		name        string
		entries     []PlayerMatch
		errContains string
	}{
		{
			name: "too many goal kickers",
			entries: func() []PlayerMatch {
				p := PositionGoals
				return []PlayerMatch{{Position: &p}, {Position: &p}, {Position: &p}, {Position: &p}}
			}(),
			errContains: "goals",
		},
		{
			name: "too many star starters",
			entries: func() []PlayerMatch {
				p := PositionStar
				return []PlayerMatch{{Position: &p}, {Position: &p}}
			}(),
			errContains: "star",
		},
		{
			name: "5 bench players",
			entries: func() []PlayerMatch {
				p := PositionGoals
				entries := []PlayerMatch{}
				for range 5 {
					entries = append(entries, PlayerMatch{Position: &p, BackupPositions: bpPtr("goals,kicks")})
				}
				return entries
			}(),
			errContains: "bench has 5",
		},
		{
			name: "two backup stars",
			entries: func() []PlayerMatch {
				p := PositionStar
				return []PlayerMatch{
					{Position: &p, BackupPositions: bpPtr("star")},
					{Position: &p, BackupPositions: bpPtr("star")},
				}
			}(),
			errContains: "backup star",
		},
		{
			name: "non-star bench with only 1 backup position",
			entries: func() []PlayerMatch {
				p := PositionGoals
				return []PlayerMatch{
					{Position: &p, BackupPositions: bpPtr("goals")},
				}
			}(),
			errContains: "exactly 2",
		},
		{
			name: "non-star bench with 3 backup positions",
			entries: func() []PlayerMatch {
				p := PositionGoals
				return []PlayerMatch{
					{Position: &p, BackupPositions: bpPtr("goals,kicks,handballs")},
				}
			}(),
			errContains: "exactly 2",
		},
		{
			name: "non-star bench with star in backup positions",
			entries: func() []PlayerMatch {
				p := PositionGoals
				return []PlayerMatch{
					{Position: &p, BackupPositions: bpPtr("goals,star")},
				}
			}(),
			errContains: "star",
		},
		{
			name: "same position covered by two bench players",
			entries: func() []PlayerMatch {
				p := PositionGoals
				return []PlayerMatch{
					{Position: &p, BackupPositions: bpPtr("goals,kicks")},
					{Position: &p, BackupPositions: bpPtr("goals,marks")},
				}
			}(),
			errContains: "goals",
		},
		{
			name: "two interchange positions",
			entries: func() []PlayerMatch {
				p := PositionGoals
				ic := "goals"
				return []PlayerMatch{
					{Position: &p, BackupPositions: bpPtr("goals,kicks"), InterchangePosition: &ic},
					{Position: &p, BackupPositions: bpPtr("marks,tackles"), InterchangePosition: &ic},
				}
			}(),
			errContains: "interchange",
		},
		{
			name: "unknown interchange position",
			entries: func() []PlayerMatch {
				p := PositionGoals
				ic := "unknown"
				return []PlayerMatch{
					{Position: &p, BackupPositions: bpPtr("goals,kicks"), InterchangePosition: &ic},
				}
			}(),
			errContains: "interchange position",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTeam(tt.entries)
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.errContains)
		})
	}
}
