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
	"xffl/services/ffl/internal/infrastructure/spreadsheet"
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
// manager: build a season (explicit rules_id + existing clubs), add a round, add a fixture.
func TestBuilder(t *testing.T) {
	ctx := context.Background()
	builder := NewBuilder(postgres.NewDB(testPool), spreadsheet.NewFixtureParser())

	eagles := createClub(ctx, t, "Builder Eagles")
	lions := createClub(ctx, t, "Builder Lions")

	built, err := builder.BuildSeason(ctx, BuildSeasonParams{
		SeasonName:  "2015",
		RulesID:     "2011",
		AFLSeasonID: 1,
		ClubIDs:     []int{eagles, lions},
	})
	require.NoError(t, err)
	assert.Equal(t, "2011", built.RulesID)
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

	// The same club can play another season (a new club_season row).
	built2, err := builder.BuildSeason(ctx, BuildSeasonParams{
		SeasonName:  "2016",
		RulesID:     "2011",
		AFLSeasonID: 2,
		ClubIDs:     []int{eagles},
	})
	require.NoError(t, err)
	assert.Len(t, built2.ClubSeasons, 1)
}

// createClub inserts a club and returns its id — the season builder now takes
// pre-existing clubs by id rather than creating them by name.
func createClub(ctx context.Context, t *testing.T, name string) int {
	t.Helper()
	club, err := postgres.NewClubRepository(sqlcgen.New(testPool)).Create(ctx, name)
	require.NoError(t, err)
	return club.ID
}

