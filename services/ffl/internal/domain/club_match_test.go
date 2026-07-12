package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pos(p Position) *Position                     { return &p }
func aflSts(s AFLStatus) *AFLStatus                { return &s }
func pmSts(s PlayerMatchStatus) *PlayerMatchStatus { return &s }
func strPtr(s string) *string                      { return &s }

// ── Score() ──────────────────────────────────────────────────────────────────

func TestClubMatch_Score(t *testing.T) {
	tests := []struct {
		name string
		cm   ClubMatch
		want int
	}{
		{
			"returns 0 with no players",
			ClubMatch{},
			0,
		},
		{
			"counts a single named starter",
			ClubMatch{PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), Status: pmSts(PlayerMatchStatusNamed), Score: 20},
			}},
			20,
		},
		{
			"sums all named starters",
			ClubMatch{PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), Status: pmSts(PlayerMatchStatusNamed), Score: 15},
				{Position: pos(PositionKicks), Status: pmSts(PlayerMatchStatusNamed), Score: 10},
				{Position: pos(PositionMarks), Status: pmSts(PlayerMatchStatusNamed), Score: 25},
			}},
			50,
		},
		{
			"named bench player does not score",
			ClubMatch{PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), Status: pmSts(PlayerMatchStatusNamed), Score: 20},
				{Score: 30, BackupPositions: strPtr("goals"), Status: pmSts(PlayerMatchStatusNamed)},
			}},
			20,
		},
		{
			"nil status starter scores (default named behaviour)",
			ClubMatch{PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), Score: 20},
				{Position: pos(PositionKicks), Score: 10},
			}},
			30,
		},
		{
			"DNP starter with no declaration scores zero",
			ClubMatch{PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), AFLStatus: aflSts(AFLStatusDNP), Score: 0},
				{Position: pos(PositionKicks), Status: pmSts(PlayerMatchStatusNamed), Score: 10},
				{Score: 25, BackupPositions: strPtr("goals"), Status: pmSts(PlayerMatchStatusNamed)},
			}},
			10, // bench stays unused; DNP scores 0
		},
		{
			"subbed_out starter does not score",
			ClubMatch{PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), Status: pmSts(PlayerMatchStatusSubbedOut), Score: 0},
				{Position: pos(PositionKicks), Status: pmSts(PlayerMatchStatusNamed), Score: 10},
			}},
			10,
		},
		{
			"interchanged_out starter does not score",
			ClubMatch{PlayerMatches: []PlayerMatch{
				{Position: pos(PositionKicks), Status: pmSts(PlayerMatchStatusInterchangedOut), Score: 5},
				{Position: pos(PositionGoals), Status: pmSts(PlayerMatchStatusNamed), Score: 10},
			}},
			10,
		},
		{
			"subbed_in bench player scores",
			ClubMatch{PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), Status: pmSts(PlayerMatchStatusSubbedOut), Score: 0},
				{Position: pos(PositionKicks), Status: pmSts(PlayerMatchStatusNamed), Score: 10},
				// Position inherited from subbed-out goals starter.
				{Position: pos(PositionGoals), Score: 25, BackupPositions: strPtr("goals"), Status: pmSts(PlayerMatchStatusSubbedIn)},
			}},
			35, // subbed_in bench (25) + kicks starter (10)
		},
		{
			"interchanged_in bench player scores",
			ClubMatch{PlayerMatches: []PlayerMatch{
				{Position: pos(PositionKicks), Status: pmSts(PlayerMatchStatusInterchangedOut), Score: 5},
				{Position: pos(PositionGoals), Status: pmSts(PlayerMatchStatusNamed), Score: 10},
				// Position set to interchange position (kicks).
				{Position: pos(PositionKicks), Score: 20, BackupPositions: strPtr("kicks"), Status: pmSts(PlayerMatchStatusInterchangedIn)},
			}},
			30, // interchanged_in bench (20) + goals starter (10)
		},
		{
			"named bench player with stale position does not score",
			ClubMatch{PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), Status: pmSts(PlayerMatchStatusNamed), Score: 10},
				// Bench player retains a stale position from a previous team submission where
				// they were a starter — must not be counted.
				{Position: pos(PositionGoals), BackupPositions: strPtr("goals,kicks"), Status: pmSts(PlayerMatchStatusNamed), Score: 5},
			}},
			10,
		},
		{
			"sub and interchange both active",
			ClubMatch{PlayerMatches: []PlayerMatch{
				{Position: pos(PositionGoals), Status: pmSts(PlayerMatchStatusSubbedOut), Score: 0},
				{Position: pos(PositionKicks), Status: pmSts(PlayerMatchStatusInterchangedOut), Score: 5},
				{Position: pos(PositionMarks), Status: pmSts(PlayerMatchStatusNamed), Score: 8},
				{Position: pos(PositionGoals), Score: 12, BackupPositions: strPtr("goals"), Status: pmSts(PlayerMatchStatusSubbedIn)},
				{Position: pos(PositionKicks), Score: 18, BackupPositions: strPtr("kicks"), Status: pmSts(PlayerMatchStatusInterchangedIn)},
			}},
			38, // 12 (sub) + 18 (interchange) + 8 (marks)
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cm.Score())
		})
	}
}

// ── DeclareSubs() validation ──────────────────────────────────────────────────

func TestClubMatch_DeclareSubs_Validation(t *testing.T) {
	t.Run("error when club match is final", func(t *testing.T) {
		cm := ClubMatch{DataStatus: ClubMatchDataFinal}
		_, err := cm.DeclareSubs(nil, nil)
		require.Error(t, err)
		assert.ErrorContains(t, err, "final")
	})

	t.Run("error when replaced ID is a bench player", func(t *testing.T) {
		cm := ClubMatch{PlayerMatches: []PlayerMatch{
			{ID: 1, Score: 5, BackupPositions: strPtr("goals,kicks")},
		}}
		_, err := cm.DeclareSubs([]SubPairing{{ReplacedPMID: 1, ReplacingPMID: 99}}, nil)
		require.Error(t, err)
		assert.ErrorContains(t, err, "bench player")
	})

	t.Run("error when replaced starter is not DNP", func(t *testing.T) {
		cm := ClubMatch{PlayerMatches: []PlayerMatch{
			{ID: 1, Position: pos(PositionGoals), AFLStatus: aflSts(AFLStatusPlayed), Score: 10},
			{ID: 2, Score: 5, BackupPositions: strPtr("goals")},
		}}
		_, err := cm.DeclareSubs([]SubPairing{{ReplacedPMID: 1, ReplacingPMID: 2}}, nil)
		require.Error(t, err)
		assert.ErrorContains(t, err, "not DNP")
	})

	t.Run("error when replaced starter has nil AFL status", func(t *testing.T) {
		cm := ClubMatch{PlayerMatches: []PlayerMatch{
			{ID: 1, Position: pos(PositionGoals), AFLStatus: nil, Score: 0},
			{ID: 2, Score: 5, BackupPositions: strPtr("goals")},
		}}
		_, err := cm.DeclareSubs([]SubPairing{{ReplacedPMID: 1, ReplacingPMID: 2}}, nil)
		require.Error(t, err)
		assert.ErrorContains(t, err, "not DNP")
	})

	t.Run("error when replacing ID is not a bench player", func(t *testing.T) {
		cm := ClubMatch{PlayerMatches: []PlayerMatch{
			{ID: 1, Position: pos(PositionGoals), AFLStatus: aflSts(AFLStatusDNP), Score: 0},
			{ID: 2, Position: pos(PositionKicks), Score: 10},
		}}
		_, err := cm.DeclareSubs([]SubPairing{{ReplacedPMID: 1, ReplacingPMID: 2}}, nil)
		require.Error(t, err)
		assert.ErrorContains(t, err, "not a bench player")
	})

	t.Run("error when interchange replaced is a bench player", func(t *testing.T) {
		cm := ClubMatch{PlayerMatches: []PlayerMatch{
			{ID: 1, Score: 5, BackupPositions: strPtr("kicks")},
			{ID: 2, Score: 10, BackupPositions: strPtr("goals")},
		}}
		_, err := cm.DeclareSubs(nil, &SubPairing{ReplacedPMID: 1, ReplacingPMID: 2})
		require.Error(t, err)
		assert.ErrorContains(t, err, "bench player")
	})
}

// ── DeclareSubs() substitution ────────────────────────────────────────────────

func TestClubMatch_DeclareSubs_SubstitutionOnly(t *testing.T) {
	cm := ClubMatch{
		PlayerMatches: []PlayerMatch{
			{ID: 1, Position: pos(PositionGoals), AFLStatus: aflSts(AFLStatusDNP), Score: 0, Status: pmSts(PlayerMatchStatusNamed)},
			{ID: 2, Position: pos(PositionKicks), AFLStatus: aflSts(AFLStatusPlayed), Score: 10, Status: pmSts(PlayerMatchStatusNamed)},
			{ID: 3, Score: 8, BackupPositions: strPtr("goals,kicks"), Status: pmSts(PlayerMatchStatusNamed)},
		},
	}
	updated, err := cm.DeclareSubs([]SubPairing{{ReplacedPMID: 1, ReplacingPMID: 3}}, nil)
	require.NoError(t, err)

	byID := make(map[int]PlayerMatch)
	for _, pm := range updated {
		byID[pm.ID] = pm
	}
	assert.Equal(t, PlayerMatchStatusSubbedOut, *byID[1].Status)
	assert.Equal(t, PlayerMatchStatusNamed, *byID[2].Status)
	assert.Equal(t, PlayerMatchStatusSubbedIn, *byID[3].Status)
}

// ── DeclareSubs() interchange ─────────────────────────────────────────────────

func TestClubMatch_DeclareSubs_InterchangeOnly(t *testing.T) {
	cm := ClubMatch{
		PlayerMatches: []PlayerMatch{
			{ID: 1, Position: pos(PositionKicks), AFLStatus: aflSts(AFLStatusPlayed), Score: 5, Status: pmSts(PlayerMatchStatusNamed)},
			{ID: 2, Position: pos(PositionGoals), AFLStatus: aflSts(AFLStatusPlayed), Score: 10, Status: pmSts(PlayerMatchStatusNamed)},
			{ID: 3, Score: 20, BackupPositions: strPtr("kicks"), Status: pmSts(PlayerMatchStatusNamed)},
		},
	}
	updated, err := cm.DeclareSubs(nil, &SubPairing{ReplacedPMID: 1, ReplacingPMID: 3})
	require.NoError(t, err)

	byID := make(map[int]PlayerMatch)
	for _, pm := range updated {
		byID[pm.ID] = pm
	}
	assert.Equal(t, PlayerMatchStatusInterchangedOut, *byID[1].Status)
	assert.Equal(t, PlayerMatchStatusNamed, *byID[2].Status)
	assert.Equal(t, PlayerMatchStatusInterchangedIn, *byID[3].Status)
}

func TestClubMatch_DeclareSubs_SubAndInterchange(t *testing.T) {
	cm := ClubMatch{
		PlayerMatches: []PlayerMatch{
			{ID: 1, Position: pos(PositionGoals), AFLStatus: aflSts(AFLStatusDNP), Score: 0, Status: pmSts(PlayerMatchStatusNamed)},
			{ID: 2, Position: pos(PositionKicks), AFLStatus: aflSts(AFLStatusPlayed), Score: 5, Status: pmSts(PlayerMatchStatusNamed)},
			{ID: 3, Position: pos(PositionMarks), AFLStatus: aflSts(AFLStatusPlayed), Score: 8, Status: pmSts(PlayerMatchStatusNamed)},
			{ID: 4, Score: 12, BackupPositions: strPtr("goals"), Status: pmSts(PlayerMatchStatusNamed)},
			{ID: 5, Score: 18, BackupPositions: strPtr("kicks"), Status: pmSts(PlayerMatchStatusNamed)},
		},
	}
	updated, err := cm.DeclareSubs(
		[]SubPairing{{ReplacedPMID: 1, ReplacingPMID: 4}},
		&SubPairing{ReplacedPMID: 2, ReplacingPMID: 5},
	)
	require.NoError(t, err)

	byID := make(map[int]PlayerMatch)
	for _, pm := range updated {
		byID[pm.ID] = pm
	}
	assert.Equal(t, PlayerMatchStatusSubbedOut, *byID[1].Status)
	assert.Equal(t, PlayerMatchStatusInterchangedOut, *byID[2].Status)
	assert.Equal(t, PlayerMatchStatusNamed, *byID[3].Status)
	assert.Equal(t, PlayerMatchStatusSubbedIn, *byID[4].Status)
	assert.Equal(t, PlayerMatchStatusInterchangedIn, *byID[5].Status)
}

// ── DeclareSubs() reset ───────────────────────────────────────────────────────

func TestClubMatch_DeclareSubs_ResetsPreviousDecisions(t *testing.T) {
	cm := ClubMatch{
		PlayerMatches: []PlayerMatch{
			{ID: 1, Position: pos(PositionGoals), AFLStatus: aflSts(AFLStatusDNP), Score: 0,
				Status: pmSts(PlayerMatchStatusSubbedOut)},
			{ID: 2, Position: pos(PositionKicks), AFLStatus: aflSts(AFLStatusPlayed), Score: 10,
				Status: pmSts(PlayerMatchStatusInterchangedOut)},
			{ID: 3, Score: 8, BackupPositions: strPtr("goals"), Status: pmSts(PlayerMatchStatusSubbedIn)},
			{ID: 4, Score: 15, BackupPositions: strPtr("kicks"), Status: pmSts(PlayerMatchStatusInterchangedIn)},
		},
	}
	// Re-declare with empty lists — all statuses reset to named.
	updated, err := cm.DeclareSubs(nil, nil)
	require.NoError(t, err)

	byID := make(map[int]PlayerMatch)
	for _, pm := range updated {
		byID[pm.ID] = pm
	}
	assert.Equal(t, PlayerMatchStatusNamed, *byID[1].Status)
	assert.Equal(t, PlayerMatchStatusNamed, *byID[2].Status)
	assert.Equal(t, PlayerMatchStatusNamed, *byID[3].Status)
	assert.Equal(t, PlayerMatchStatusNamed, *byID[4].Status)
}

// ── SuggestedSubstitutions() ─────────────────────────────────────────────────
// Bench shape mirrors real data: position=nil, backup_positions set, interchange_position set
// only for the interchange slot player.

func TestClubMatch_SuggestedSubstitutions_Interchange(t *testing.T) {
	icBench := func(score int, status PlayerMatchStatus) PlayerMatch {
		return PlayerMatch{ID: 99, BackupPositions: strPtr("star"), InterchangePosition: strPtr("star"), Status: pmSts(status), Score: score}
	}
	star := func(id, score int) PlayerMatch {
		return PlayerMatch{ID: id, Position: pos(PositionStar), Status: pmSts(PlayerMatchStatusNamed), Score: score}
	}

	tests := []struct {
		name         string
		cm           ClubMatch
		wantKind     SubstitutionKind
		wantReplaced int
	}{
		{"no interchange bench → empty", ClubMatch{PlayerMatches: []PlayerMatch{star(1, 41)}}, "", 0},
		{"bench outscores star → interchange", ClubMatch{PlayerMatches: []PlayerMatch{star(1, 41), icBench(63, PlayerMatchStatusNamed)}}, SubstitutionKindInterchange, 1},
		{"bench equals star → no suggestion", ClubMatch{PlayerMatches: []PlayerMatch{star(1, 41), icBench(41, PlayerMatchStatusNamed)}}, "", 0},
		{"bench lower than star → no suggestion", ClubMatch{PlayerMatches: []PlayerMatch{star(1, 41), icBench(30, PlayerMatchStatusNamed)}}, "", 0},
		{"bench already interchanged_in → no suggestion", ClubMatch{PlayerMatches: []PlayerMatch{
			{ID: 1, Position: pos(PositionStar), Status: pmSts(PlayerMatchStatusInterchangedOut), Score: 41},
			icBench(63, PlayerMatchStatusInterchangedIn),
		}}, "", 0},
		{"picks lowest star when multiple", ClubMatch{PlayerMatches: []PlayerMatch{star(1, 72), star(2, 41), icBench(63, PlayerMatchStatusNamed)}}, SubstitutionKindInterchange, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cm.SuggestedSubstitutions()
			interchanges := filterKind(got, SubstitutionKindInterchange)
			if tt.wantKind == "" {
				assert.Empty(t, interchanges)
			} else {
				require.Len(t, interchanges, 1)
				assert.Equal(t, tt.wantReplaced, interchanges[0].ReplacedPMID)
				assert.Equal(t, 99, interchanges[0].ReplacingPMID)
			}
		})
	}
}

func TestClubMatch_SuggestedSubstitutions_Sub(t *testing.T) {
	backupBench := func(id int, pos1, pos2 string, aflStatus AFLStatus) PlayerMatch {
		return PlayerMatch{ID: id, BackupPositions: strPtr(pos1 + "," + pos2), Status: pmSts(PlayerMatchStatusNamed), AFLStatus: aflSts(aflStatus)}
	}
	dnpStarter := func(id int, p Position) PlayerMatch {
		return PlayerMatch{ID: id, Position: &p, Status: pmSts(PlayerMatchStatusNamed), AFLStatus: aflSts(AFLStatusDNP)}
	}
	activeStarter := func(id int, p Position) PlayerMatch {
		return PlayerMatch{ID: id, Position: &p, Status: pmSts(PlayerMatchStatusNamed), AFLStatus: aflSts(AFLStatusPlayed)}
	}

	t.Run("DNP starter with covering bench → sub", func(t *testing.T) {
		cm := ClubMatch{PlayerMatches: []PlayerMatch{
			dnpStarter(1, PositionGoals),
			backupBench(10, "goals", "marks", AFLStatusPlayed),
		}}
		got := filterKind(cm.SuggestedSubstitutions(), SubstitutionKindSub)
		require.Len(t, got, 1)
		assert.Equal(t, 1, got[0].ReplacedPMID)
		assert.Equal(t, 10, got[0].ReplacingPMID)
	})

	t.Run("active starter → no sub", func(t *testing.T) {
		cm := ClubMatch{PlayerMatches: []PlayerMatch{
			activeStarter(1, PositionGoals),
			backupBench(10, "goals", "marks", AFLStatusPlayed),
		}}
		assert.Empty(t, filterKind(cm.SuggestedSubstitutions(), SubstitutionKindSub))
	})

	t.Run("DNP bench player → no sub", func(t *testing.T) {
		cm := ClubMatch{PlayerMatches: []PlayerMatch{
			dnpStarter(1, PositionGoals),
			backupBench(10, "goals", "marks", AFLStatusDNP),
		}}
		assert.Empty(t, filterKind(cm.SuggestedSubstitutions(), SubstitutionKindSub))
	})

	t.Run("bye bench player → sub suggested", func(t *testing.T) {
		cm := ClubMatch{PlayerMatches: []PlayerMatch{
			dnpStarter(1, PositionGoals),
			backupBench(10, "goals", "marks", AFLStatusBye),
		}}
		got := filterKind(cm.SuggestedSubstitutions(), SubstitutionKindSub)
		require.Len(t, got, 1)
		assert.Equal(t, 1, got[0].ReplacedPMID)
	})

	t.Run("already subbed_out → no re-suggestion", func(t *testing.T) {
		cm := ClubMatch{PlayerMatches: []PlayerMatch{
			{ID: 1, Position: pos(PositionGoals), Status: pmSts(PlayerMatchStatusSubbedOut), AFLStatus: aflSts(AFLStatusDNP)},
			backupBench(10, "goals", "marks", AFLStatusPlayed),
		}}
		assert.Empty(t, filterKind(cm.SuggestedSubstitutions(), SubstitutionKindSub))
	})

	t.Run("DNP star with only interchange bench → interchange only, no sub", func(t *testing.T) {
		// Interchange bench player covers star but is not eligible for sub suggestions.
		// Only an interchange suggestion fires (bench outscores DNP star's score of 0).
		cm := ClubMatch{PlayerMatches: []PlayerMatch{
			{ID: 1, Position: pos(PositionStar), Status: pmSts(PlayerMatchStatusNamed), AFLStatus: aflSts(AFLStatusDNP), Score: 0},
			{ID: 99, BackupPositions: strPtr("star"), InterchangePosition: strPtr("star"), Status: pmSts(PlayerMatchStatusNamed), Score: 63},
		}}
		got := cm.SuggestedSubstitutions()
		require.Len(t, got, 1)
		assert.Equal(t, SubstitutionKindInterchange, got[0].Kind)
	})

	t.Run("DNP starter with both interchange bench and backup bench → both suggestions", func(t *testing.T) {
		// A regular backup bench player also covers goals → sub fires.
		// Interchange bench player outscores the goals starter → interchange also fires.
		cm := ClubMatch{PlayerMatches: []PlayerMatch{
			{ID: 1, Position: pos(PositionGoals), Status: pmSts(PlayerMatchStatusNamed), AFLStatus: aflSts(AFLStatusDNP), Score: 0},
			{ID: 10, BackupPositions: strPtr("goals,marks"), Status: pmSts(PlayerMatchStatusNamed), AFLStatus: aflSts(AFLStatusPlayed)},
			{ID: 99, BackupPositions: strPtr("star"), InterchangePosition: strPtr("star"), Status: pmSts(PlayerMatchStatusNamed), Score: 63},
			{ID: 2, Position: pos(PositionStar), Status: pmSts(PlayerMatchStatusNamed), Score: 41},
		}}
		got := cm.SuggestedSubstitutions()
		assert.Len(t, got, 2)
		assert.Equal(t, SubstitutionKindInterchange, got[0].Kind)
		assert.Equal(t, SubstitutionKindSub, got[1].Kind)
		assert.Equal(t, 1, got[1].ReplacedPMID)
		assert.Equal(t, 10, got[1].ReplacingPMID)
	})
}

func filterKind(ss []SuggestedSubstitution, kind SubstitutionKind) []SuggestedSubstitution {
	var out []SuggestedSubstitution
	for _, s := range ss {
		if s.Kind == kind {
			out = append(out, s)
		}
	}
	return out
}

func TestClubMatch_DeclareSubs_RedeclareReplacesPreviousPairing(t *testing.T) {
	// First pairing: starter 1 ↔ bench 3.
	cm := ClubMatch{
		PlayerMatches: []PlayerMatch{
			{ID: 1, Position: pos(PositionGoals), AFLStatus: aflSts(AFLStatusDNP), Score: 0,
				Status: pmSts(PlayerMatchStatusSubbedOut)},
			{ID: 2, Position: pos(PositionKicks), AFLStatus: aflSts(AFLStatusDNP), Score: 0,
				Status: pmSts(PlayerMatchStatusNamed)},
			{ID: 3, Score: 10, BackupPositions: strPtr("goals"), Status: pmSts(PlayerMatchStatusSubbedIn)},
		},
	}
	// Re-declare: now sub out starter 2 instead.
	updated, err := cm.DeclareSubs([]SubPairing{{ReplacedPMID: 2, ReplacingPMID: 3}}, nil)
	require.NoError(t, err)

	byID := make(map[int]PlayerMatch)
	for _, pm := range updated {
		byID[pm.ID] = pm
	}
	assert.Equal(t, PlayerMatchStatusNamed, *byID[1].Status)     // reset from subbed_out
	assert.Equal(t, PlayerMatchStatusSubbedOut, *byID[2].Status) // new sub
	assert.Equal(t, PlayerMatchStatusSubbedIn, *byID[3].Status)  // still subbed in
}
