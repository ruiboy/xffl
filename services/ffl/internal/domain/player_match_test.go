package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func bpPtr(s string) *string { return &s }

// validFullTeam builds a complete 18-starter team with no bench.
func validFullTeam() []UpsertPlayerMatchParams {
	entries := []UpsertPlayerMatchParams{}
	for pos, count := range PositionSlots {
		p := pos
		for range count {
			entries = append(entries, UpsertPlayerMatchParams{Position: &p})
		}
	}
	return entries
}

func TestValidateTeam_ValidCases(t *testing.T) {
	t.Run("empty team is valid", func(t *testing.T) {
		require.NoError(t, ValidateTeam(nil))
	})

	t.Run("full 18-starter team is valid", func(t *testing.T) {
		require.NoError(t, ValidateTeam(validFullTeam()))
	})

	t.Run("starters with backup star and 3 dual-position bench", func(t *testing.T) {
		entries := validFullTeam()
		star := PositionStar
		entries = append(entries, UpsertPlayerMatchParams{Position: &star, BackupPositions: bpPtr("star")})
		goals := PositionGoals
		entries = append(entries, UpsertPlayerMatchParams{Position: &goals, BackupPositions: bpPtr("goals,kicks")})
		handballs := PositionHandballs
		entries = append(entries, UpsertPlayerMatchParams{Position: &handballs, BackupPositions: bpPtr("handballs,marks")})
		tackles := PositionTackles
		entries = append(entries, UpsertPlayerMatchParams{Position: &tackles, BackupPositions: bpPtr("tackles,hitouts")})
		require.NoError(t, ValidateTeam(entries))
	})

	t.Run("interchange on bench star is valid", func(t *testing.T) {
		entries := validFullTeam()
		star := PositionStar
		ic := "star"
		entries = append(entries, UpsertPlayerMatchParams{Position: &star, BackupPositions: bpPtr("star"), InterchangePosition: &ic})
		require.NoError(t, ValidateTeam(entries))
	})

	t.Run("partial team is valid", func(t *testing.T) {
		goals := PositionGoals
		entries := []UpsertPlayerMatchParams{
			{Position: &goals},
			{Position: &goals},
		}
		require.NoError(t, ValidateTeam(entries))
	})
}

func TestValidateTeam_InvalidCases(t *testing.T) {
	tests := []struct {
		name        string
		entries     []UpsertPlayerMatchParams
		errContains string
	}{
		{
			name: "too many goal kickers",
			entries: func() []UpsertPlayerMatchParams {
				p := PositionGoals
				return []UpsertPlayerMatchParams{{Position: &p}, {Position: &p}, {Position: &p}, {Position: &p}}
			}(),
			errContains: "goals",
		},
		{
			name: "too many star starters",
			entries: func() []UpsertPlayerMatchParams {
				p := PositionStar
				return []UpsertPlayerMatchParams{{Position: &p}, {Position: &p}}
			}(),
			errContains: "star",
		},
		{
			name: "5 bench players",
			entries: func() []UpsertPlayerMatchParams {
				p := PositionGoals
				entries := []UpsertPlayerMatchParams{}
				for range 5 {
					entries = append(entries, UpsertPlayerMatchParams{Position: &p, BackupPositions: bpPtr("goals,kicks")})
				}
				return entries
			}(),
			errContains: "bench has 5",
		},
		{
			name: "two backup stars",
			entries: func() []UpsertPlayerMatchParams {
				p := PositionStar
				return []UpsertPlayerMatchParams{
					{Position: &p, BackupPositions: bpPtr("star")},
					{Position: &p, BackupPositions: bpPtr("star")},
				}
			}(),
			errContains: "backup star",
		},
		{
			name: "non-star bench with only 1 backup position",
			entries: func() []UpsertPlayerMatchParams {
				p := PositionGoals
				return []UpsertPlayerMatchParams{
					{Position: &p, BackupPositions: bpPtr("goals")},
				}
			}(),
			errContains: "exactly 2",
		},
		{
			name: "non-star bench with 3 backup positions",
			entries: func() []UpsertPlayerMatchParams {
				p := PositionGoals
				return []UpsertPlayerMatchParams{
					{Position: &p, BackupPositions: bpPtr("goals,kicks,handballs")},
				}
			}(),
			errContains: "exactly 2",
		},
		{
			name: "non-star bench with star in backup positions",
			entries: func() []UpsertPlayerMatchParams {
				p := PositionGoals
				return []UpsertPlayerMatchParams{
					{Position: &p, BackupPositions: bpPtr("goals,star")},
				}
			}(),
			errContains: "star",
		},
		{
			name: "same position covered by two bench players",
			entries: func() []UpsertPlayerMatchParams {
				p := PositionGoals
				return []UpsertPlayerMatchParams{
					{Position: &p, BackupPositions: bpPtr("goals,kicks")},
					{Position: &p, BackupPositions: bpPtr("goals,marks")},
				}
			}(),
			errContains: "goals",
		},
		{
			name: "two interchange positions",
			entries: func() []UpsertPlayerMatchParams {
				p := PositionGoals
				ic := "goals"
				return []UpsertPlayerMatchParams{
					{Position: &p, BackupPositions: bpPtr("goals,kicks"), InterchangePosition: &ic},
					{Position: &p, BackupPositions: bpPtr("marks,tackles"), InterchangePosition: &ic},
				}
			}(),
			errContains: "interchange",
		},
		{
			name: "unknown interchange position",
			entries: func() []UpsertPlayerMatchParams {
				p := PositionGoals
				ic := "unknown"
				return []UpsertPlayerMatchParams{
					{Position: &p, BackupPositions: bpPtr("goals,kicks"), InterchangePosition: &ic},
				}
			}(),
			errContains: "interchange position",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTeam(tt.entries)
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.errContains)
		})
	}
}
