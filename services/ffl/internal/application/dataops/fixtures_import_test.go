package dataops

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/ffl/internal/application"
	"xffl/services/ffl/internal/domain"
)

func intp(n int) *int { return &n }

func TestPlanFixtureImport(t *testing.T) {
	// Five clubs; one byes each round.
	nameToCS := map[string]int{
		"thc": 1, "ruiboys": 2, "cheetahs": 3, "slashers": 4, "bombers": 5,
	}
	seasonCS := []int{1, 2, 3, 4, 5}

	rounds := []FixtureImportRound{{
		Name:       "Round 1",
		AFLRoundID: 10,
		Type:       domain.RoundTypeMinor,
		Fixtures: []application.ParsedFixture{
			{HomeClub: "THC", HomeScore: intp(452), AwayClub: "Ruiboys", AwayScore: intp(372)},
			{HomeClub: "Cheetahs", HomeScore: intp(370), AwayClub: "Slashers", AwayScore: intp(326)},
		},
	}}

	specs, refs, unresolved := planFixtureImport(rounds, seasonCS, nameToCS)

	require.Empty(t, unresolved)
	require.Len(t, specs, 1)
	spec := specs[0]
	assert.Equal(t, "Round 1", spec.Name)
	assert.Equal(t, 10, spec.AFLRoundID)

	t.Run("two versus matches in club-name order", func(t *testing.T) {
		require.GreaterOrEqual(t, len(spec.Matches), 2)
		assert.Equal(t, domain.MatchStyleVersus, spec.Matches[0].Style)
		assert.Equal(t, []int{1, 2}, spec.Matches[0].ClubSeasonIDs)
		assert.Equal(t, []int{3, 4}, spec.Matches[1].ClubSeasonIDs)
	})

	t.Run("the club not playing gets a bye", func(t *testing.T) {
		var byes []MatchSpec
		for _, m := range spec.Matches {
			if m.Style == domain.MatchStyleBye {
				byes = append(byes, m)
			}
		}
		require.Len(t, byes, 1)
		assert.Equal(t, []int{5}, byes[0].ClubSeasonIDs) // Bombers
	})

	t.Run("reference scores captured per club_season", func(t *testing.T) {
		got := map[int]int{}
		for _, r := range refs {
			assert.Equal(t, "Round 1", r.RoundName)
			got[r.ClubSeasonID] = r.Score
		}
		assert.Equal(t, map[int]int{1: 452, 2: 372, 3: 370, 4: 326}, got)
	})
}

func TestPlanFixtureImport_FinalsGetNoByes(t *testing.T) {
	nameToCS := map[string]int{"thc": 1, "ruiboys": 2, "cheetahs": 3, "slashers": 4}
	seasonCS := []int{1, 2, 3, 4}

	rounds := []FixtureImportRound{{
		Name:       "Grand Final",
		AFLRoundID: 30,
		Type:       domain.RoundTypeGrandFinal,
		Fixtures: []application.ParsedFixture{
			{HomeClub: "THC", AwayClub: "Ruiboys"},
		},
	}}

	specs, _, unresolved := planFixtureImport(rounds, seasonCS, nameToCS)
	require.Empty(t, unresolved)
	require.Len(t, specs, 1)

	for _, m := range specs[0].Matches {
		assert.NotEqual(t, domain.MatchStyleBye, m.Style, "a final must not fabricate byes for the eliminated clubs")
	}
}

func TestPlanFixtureImport_UnresolvedClubBlocks(t *testing.T) {
	nameToCS := map[string]int{"thc": 1, "ruiboys": 2}
	rounds := []FixtureImportRound{{
		Name: "Round 1",
		Fixtures: []application.ParsedFixture{
			{HomeClub: "THC", AwayClub: "Ghosts"}, // Ghosts not registered
		},
	}}

	specs, _, unresolved := planFixtureImport(rounds, []int{1, 2}, nameToCS)

	assert.Equal(t, []string{"Ghosts"}, unresolved)
	// The unresolved fixture is dropped, so no versus match is planned — callers
	// must not import while anything is unresolved (ImportFixtures enforces this).
	var versusCount int
	for _, m := range specs[0].Matches {
		if m.Style == domain.MatchStyleVersus {
			versusCount++
		}
	}
	assert.Zero(t, versusCount)
}
