//go:build integration

package graphql_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/ffl/internal/application"
	pg "xffl/services/ffl/internal/infrastructure/postgres"
	"xffl/services/ffl/internal/infrastructure/postgres/sqlcgen"
	memevents "xffl/shared/events/memory"
)

// finalStatusStub implements application.PlayerLookup. LookupPlayerMatchBySeasonRound
// always returns no stats (simulating a player with no AFL player_match row);
// LookupFinalAFLStatus returns the canned finalStatus map, simulating the AFL round
// already having finalised.
type finalStatusStub struct {
	finalStatus map[int]string // keyed by AFL player_season_id
}

func (s *finalStatusStub) LookupPlayers(_ context.Context, _ []int) ([]application.PlayerCandidate, error) {
	return nil, nil
}
func (s *finalStatusStub) LookupPlayerSeason(_ context.Context, _ int) (int, error) { return 0, nil }
func (s *finalStatusStub) LookupPlayerMatch(_ context.Context, _ []int) ([]application.PlayerMatchStats, error) {
	return nil, nil
}
func (s *finalStatusStub) LookupPlayerMatchBySeasonRound(_ context.Context, _ []int, _ int) ([]application.PlayerMatchStats, error) {
	return nil, nil
}
func (s *finalStatusStub) LookupByeInfo(_ context.Context, _ []int, _ int) ([]application.ByePlayerInfo, error) {
	return nil, nil
}
func (s *finalStatusStub) LookupPlayerSeasonsBySeasonID(_ context.Context, _ int) ([]application.PlayerCandidate, error) {
	return nil, nil
}
func (s *finalStatusStub) LookupFinalAFLStatus(_ context.Context, _ []int, _ int) (map[int]string, error) {
	return s.finalStatus, nil
}

func commandsWithFinalStatusStub(pool *pgxpool.Pool, finalStatus map[int]string) *application.Commands {
	q := sqlcgen.New(pool)
	db := pg.NewDB(pool)
	return application.NewCommands(
		db,
		memevents.New(),
		&finalStatusStub{finalStatus: finalStatus},
		pg.NewMatchRepository(q),
		pg.NewClubMatchRepository(q),
		pg.NewClubSeasonRepository(q),
		pg.NewRoundRepository(q),
		pg.NewPlayerMatchRepository(q),
		pg.NewPlayerSeasonRepository(q),
	)
}

// seedUnlinkedPlayerMatch inserts a second player_season (afl_player_season_id=2, distinct
// from seedTestData's afl_player_season_id=1) with a player_match on the away club_match
// that has no afl_player_match_id link and no drv_afl_status — as if their FFL lineup row
// was created after their real AFL match had already finalised.
func seedUnlinkedPlayerMatch(t *testing.T, pool *pgxpool.Pool, ids testIDs) (playerSeasonID, playerMatchID int) {
	t.Helper()
	ctx := context.Background()

	var aflPlayerID, playerID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player (name) VALUES ('Unlinked AFL Player') RETURNING id").Scan(&aflPlayerID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.player (afl_player_id) VALUES ($1) RETURNING id", aflPlayerID).Scan(&playerID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.player_season (player_id, club_season_id, afl_player_season_id) VALUES ($1, $2, 2) RETURNING id",
		playerID, ids.awayClubSeaID).Scan(&playerSeasonID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status) VALUES ($1, $2, 'kicks', 'named') RETURNING id",
		ids.awayClubMatchID, playerSeasonID).Scan(&playerMatchID))
	return playerSeasonID, playerMatchID
}

func TestRecalculateScore_InfersDNPForConfirmedFinalMatch(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	_, playerMatchID := seedUnlinkedPlayerMatch(t, pool, ids)

	// afl_player_season_id=2 is confirmed dnp for AFL round 1 (seedTestData's round has
	// afl_round_id=1).
	commands := commandsWithFinalStatusStub(pool, map[int]string{2: "dnp"})

	_, err := commands.RecalculateScore(context.Background(), ids.awayClubMatchID)
	require.NoError(t, err)

	var drvAflStatus *string
	require.NoError(t, pool.QueryRow(context.Background(),
		"SELECT drv_afl_status FROM ffl.player_match WHERE id = $1", playerMatchID).Scan(&drvAflStatus))
	require.NotNil(t, drvAflStatus, "drv_afl_status should be set once the AFL match is confirmed final")
	assert.Equal(t, "dnp", *drvAflStatus)
}

func TestRecalculateScore_LeavesStatusUnsetWhenMatchNotYetFinal(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	_, playerMatchID := seedUnlinkedPlayerMatch(t, pool, ids)

	// No entry for afl_player_season_id=2 — the AFL match for their club hasn't
	// finalised yet, so no confident status should be applied.
	commands := commandsWithFinalStatusStub(pool, map[int]string{})

	_, err := commands.RecalculateScore(context.Background(), ids.awayClubMatchID)
	require.NoError(t, err)

	var drvAflStatus *string
	require.NoError(t, pool.QueryRow(context.Background(),
		"SELECT drv_afl_status FROM ffl.player_match WHERE id = $1", playerMatchID).Scan(&drvAflStatus))
	assert.Nil(t, drvAflStatus, "status must stay unset until confirmed, not guessed")
}

func TestRecalculateScore_DNPInferenceDoesNotWipePosition(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	_, playerMatchID := seedUnlinkedPlayerMatch(t, pool, ids)

	commands := commandsWithFinalStatusStub(pool, map[int]string{2: "dnp"})
	_, err := commands.RecalculateScore(context.Background(), ids.awayClubMatchID)
	require.NoError(t, err)

	var position *string
	require.NoError(t, pool.QueryRow(context.Background(),
		"SELECT position FROM ffl.player_match WHERE id = $1", playerMatchID).Scan(&position))
	require.NotNil(t, position)
	assert.Equal(t, "kicks", *position, "setting dnp status must not clobber the player's position")
}
