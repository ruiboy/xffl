//go:build integration

package dataops

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/ffl/internal/domain"
	"xffl/services/ffl/internal/infrastructure/postgres"
	"xffl/services/ffl/internal/infrastructure/postgres/sqlcgen"
	"xffl/services/ffl/internal/testutil"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	pool, cleanup, err := testutil.StartPostgres(context.Background())
	if err != nil {
		panic(err)
	}
	testPool = pool
	code := m.Run()
	cleanup()
	os.Exit(code)
}

// TestBuilder exercises the season/fixture builder through the real transaction
// manager: build a season (year → rules_id), add a round, add a fixture.
func TestBuilder(t *testing.T) {
	ctx := context.Background()
	builder := NewBuilder(postgres.NewDB(testPool))

	built, err := builder.BuildSeason(ctx, BuildSeasonParams{
		Year:        2015,
		AFLSeasonID: 1,
		ClubNames:   []string{"Builder Eagles", "Builder Lions"},
	})
	require.NoError(t, err)
	assert.Equal(t, "2011", built.RulesID, "2015 maps to the 2011 era")
	require.Len(t, built.ClubSeasons, 2)

	round, err := builder.AddRound(ctx, built.SeasonID, "Round 1", 1, domain.RoundTypeMinor)
	require.NoError(t, err)

	fixture, err := builder.AddFixture(ctx, round.ID, built.ClubSeasons[0].ClubSeasonID, built.ClubSeasons[1].ClubSeasonID)
	require.NoError(t, err)
	assert.NotZero(t, fixture.MatchID)
	assert.NotZero(t, fixture.HomeClubMatchID)
	assert.NotZero(t, fixture.AwayClubMatchID)

	// The match reads back with both club_matches wired in.
	got, err := postgres.NewMatchRepository(sqlcgen.New(testPool)).FindByID(ctx, fixture.MatchID)
	require.NoError(t, err)
	assert.Equal(t, fixture.HomeClubMatchID, got.Home.ID)
	assert.Equal(t, fixture.AwayClubMatchID, got.Away.ID)

	// Re-building the same clubs in another season reuses the club rows.
	built2, err := builder.BuildSeason(ctx, BuildSeasonParams{
		Year:        2016,
		AFLSeasonID: 2,
		ClubNames:   []string{"Builder Eagles"},
	})
	require.NoError(t, err)
	assert.Len(t, built2.ClubSeasons, 1)
}

// TestBuilder_GenerateHomeAndAway builds a full 6-round round-robin in one call.
func TestBuilder_GenerateHomeAndAway(t *testing.T) {
	ctx := context.Background()
	builder := NewBuilder(postgres.NewDB(testPool))

	built, err := builder.BuildSeason(ctx, BuildSeasonParams{
		Year:        2017,
		AFLSeasonID: 3,
		ClubNames:   []string{"Gen A", "Gen B", "Gen C", "Gen D"},
	})
	require.NoError(t, err)

	csIDs := make([]int, len(built.ClubSeasons))
	for i, cs := range built.ClubSeasons {
		csIDs[i] = cs.ClubSeasonID
	}

	rounds, err := builder.GenerateHomeAndAway(ctx, built.SeasonID, csIDs, 6, 100, "Round ")
	require.NoError(t, err)
	require.Len(t, rounds, 6)
	for _, r := range rounds {
		assert.Len(t, r.Fixtures, 2, "4 clubs => 2 fixtures per round")
	}
}
