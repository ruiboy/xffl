//go:build integration

package dataops

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/ffl/internal/application"
	"xffl/services/ffl/internal/domain"
	"xffl/services/ffl/internal/infrastructure/postgres"
	"xffl/services/ffl/internal/infrastructure/postgres/sqlcgen"
	"xffl/services/ffl/internal/infrastructure/spreadsheet"
)

// TestImportFixtures_BuildsFixturesAndReferenceScores drives the whole Slice 4
// commit path: resolve club names to club_seasons, build the round (versus + an
// inferred bye for the odd club), and stamp the reference scores onto notes.
func TestImportFixtures_BuildsFixturesAndReferenceScores(t *testing.T) {
	ctx := context.Background()
	builder := NewBuilder(postgres.NewDB(testPool), spreadsheet.NewFixtureParser())
	// buildOddSeason names the clubs "<name> A/B/C"; ImportFixtures resolves by name.
	seasonID, csA, csB, csC := buildOddSeason(ctx, t, "ImpFix", 2)

	res, err := builder.ImportFixtures(ctx, ImportFixturesParams{
		SeasonID: seasonID,
		Rounds: []FixtureImportRound{{
			Name: "Round 1", AFLRoundID: 10, Type: domain.RoundTypeMinor,
			Fixtures: []application.ParsedFixture{
				{HomeClub: "ImpFix A", HomeScore: intp(452), AwayClub: "ImpFix B", AwayScore: intp(372)},
			},
		}},
	})
	require.NoError(t, err)
	require.Empty(t, res.Unresolved)
	assert.Equal(t, 1, res.RoundsCreated)
	assert.Equal(t, 2, res.ScoresWritten)

	roundID := onlyRoundID(ctx, t, seasonID)
	groups := roundClubMatches(ctx, t, roundID)
	require.Len(t, groups, 2, "one versus + one bye")
	assert.Equal(t, map[int]bool{csA: true, csB: true, csC: true}, clubSeasonSet(groups))

	notesByCS := map[int]string{}
	for _, g := range groups {
		for _, cm := range g {
			if cm.Notes != nil {
				notesByCS[cm.ClubSeasonID] = *cm.Notes
			}
		}
	}
	assert.Equal(t, "spreadsheet:452", notesByCS[csA])
	assert.Equal(t, "spreadsheet:372", notesByCS[csB])
	assert.NotContains(t, notesByCS, csC, "the bye club gets no reference score")
}

// TestImportFixtures_UnresolvedClubWritesNothing verifies the safety guard: a club
// name with no club_season blocks the whole import rather than dropping a fixture.
func TestImportFixtures_UnresolvedClubWritesNothing(t *testing.T) {
	ctx := context.Background()
	builder := NewBuilder(postgres.NewDB(testPool), spreadsheet.NewFixtureParser())
	seasonID, _, _, _ := buildOddSeason(ctx, t, "ImpUnres", 3)

	res, err := builder.ImportFixtures(ctx, ImportFixturesParams{
		SeasonID: seasonID,
		Rounds: []FixtureImportRound{{
			Name: "Round 1", AFLRoundID: 10, Type: domain.RoundTypeMinor,
			Fixtures: []application.ParsedFixture{
				{HomeClub: "ImpUnres A", AwayClub: "Nonexistent FC"},
			},
		}},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"Nonexistent FC"}, res.Unresolved)
	assert.Zero(t, res.ScoresWritten)

	rounds, err := postgres.NewRoundRepository(sqlcgen.New(testPool)).FindBySeasonID(ctx, seasonID)
	require.NoError(t, err)
	assert.Empty(t, rounds, "nothing is written when a club is unresolved")
}
