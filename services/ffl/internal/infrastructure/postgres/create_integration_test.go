//go:build integration

package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/ffl/internal/domain"
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

// TestCreateSeasonStructure exercises the whole write-side chain: league → season
// → clubs → club_seasons → round → match → home/away club_matches.
func TestCreateSeasonStructure(t *testing.T) {
	ctx := context.Background()
	q := sqlcgen.New(testPool)

	leagues := NewLeagueRepository(q)
	seasons := NewSeasonRepository(q)
	clubs := NewClubRepository(q)
	clubSeasons := NewClubSeasonRepository(q)
	rounds := NewRoundRepository(q)
	matches := NewMatchRepository(q)
	clubMatches := NewClubMatchRepository(q)

	league, err := leagues.Create(ctx, "Create Test FFL")
	require.NoError(t, err)

	season, err := seasons.Create(ctx, league.ID, "2015", 1, "2011")
	require.NoError(t, err)
	assert.Equal(t, "2011", season.RulesID)
	assert.Equal(t, league.ID, season.LeagueID)

	// find-or-create clubs
	home, err := clubs.Create(ctx, "Create Test Eagles")
	require.NoError(t, err)
	away, err := clubs.Create(ctx, "Create Test Lions")
	require.NoError(t, err)

	// FindByName finds the just-created club; missing name is ErrNotFound.
	found, err := clubs.FindByName(ctx, "Create Test Eagles")
	require.NoError(t, err)
	assert.Equal(t, home.ID, found.ID)
	_, err = clubs.FindByName(ctx, "No Such Club")
	assert.ErrorIs(t, err, domain.ErrNotFound)

	homeCS, err := clubSeasons.Create(ctx, home.ID, season.ID)
	require.NoError(t, err)
	awayCS, err := clubSeasons.Create(ctx, away.ID, season.ID)
	require.NoError(t, err)

	round, err := rounds.Create(ctx, season.ID, "Round 1", 1, domain.RoundTypeMinor)
	require.NoError(t, err)
	assert.Equal(t, domain.RoundTypeMinor, round.Type)

	match, err := matches.Create(ctx, round.ID, nil)
	require.NoError(t, err)
	assert.Equal(t, round.ID, match.RoundID)

	homeCM, err := clubMatches.Create(ctx, match.ID, homeCS.ID, "home")
	require.NoError(t, err)
	awayCM, err := clubMatches.Create(ctx, match.ID, awayCS.ID, "away")
	require.NoError(t, err)

	// The match now reads back with both club_matches wired to it.
	got, err := matches.FindByID(ctx, match.ID)
	require.NoError(t, err)
	assert.Equal(t, homeCM.ID, got.Home.ID)
	assert.Equal(t, awayCM.ID, got.Away.ID)

	// club_seasons appear on the season ladder.
	ladder, err := clubSeasons.FindBySeasonID(ctx, season.ID)
	require.NoError(t, err)
	assert.Len(t, ladder, 2)
}
