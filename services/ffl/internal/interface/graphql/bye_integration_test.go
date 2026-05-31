//go:build integration

package graphql_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/ffl/internal/application"
	"xffl/services/ffl/internal/domain"
	pg "xffl/services/ffl/internal/infrastructure/postgres"
	"xffl/services/ffl/internal/infrastructure/postgres/sqlcgen"
	memevents "xffl/shared/events/memory"
)

// setupByeCommands creates a Commands instance with a stub that returns the given bye info
// keyed by AFL player_season_id.
func setupByeCommands(t *testing.T, pool *pgxpool.Pool, byeInfo map[int]application.ByePlayerInfo) *application.Commands {
	t.Helper()
	q := sqlcgen.New(pool)
	return application.NewCommands(
		pg.NewDB(pool),
		memevents.New(),
		&stubPlayerLookup{pool: pool, byeInfo: byeInfo},
		pg.NewMatchRepository(q),
		pg.NewClubMatchRepository(q),
		pg.NewClubSeasonRepository(q),
		pg.NewRoundRepository(q),
		pg.NewPlayerMatchRepository(q),
		pg.NewPlayerSeasonRepository(q),
	)
}

// seedByePlayer inserts an AFL player row, a FFL player, and a FFL player_season linked to
// a stub AFL player_season_id of your choosing (no real afl.player_season row is created —
// the AFL service call is mocked by stubPlayerLookup). Returns the FFL player_season_id.
// Cleanup is handled by cleanupTestData via CASCADE.
func seedByePlayer(t *testing.T, pool *pgxpool.Pool, ids testIDs, name string, aflPSIDStub int) int {
	t.Helper()
	ctx := context.Background()

	var aflPlayerID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player (name) VALUES ($1) RETURNING id", name).Scan(&aflPlayerID))

	var fflPlayerID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.player (afl_player_id) VALUES ($1) RETURNING id", aflPlayerID).Scan(&fflPlayerID))

	var fflPSID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.player_season (player_id, club_season_id, afl_player_season_id) VALUES ($1, $2, $3) RETURNING id",
		fflPlayerID, ids.homeClubSeaID, aflPSIDStub).Scan(&fflPSID))

	return fflPSID
}

// ── Test 1: eligible bye starter at kicks → drv_score = floor(avg)*1 ─────────

func TestSetTeam_ByeStarter_EligibleScoresViaAverage(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	ctx := context.Background()

	const stubAFLPSID = 9001
	fflPSID := seedByePlayer(t, pool, ids, "Bye Kicks Player", stubAFLPSID)

	// avg_kicks = 14.8 → floor(14.8)*1 = 14
	commands := setupByeCommands(t, pool, map[int]application.ByePlayerInfo{
		stubAFLPSID: {PlayerSeasonID: stubAFLPSID, HasBye: true, PlayedLast: true, AvgKicks: 14.8},
	})

	_, err := commands.SetTeam(ctx, application.SetTeamParams{
		ClubMatchID: ids.homeClubMatchID,
		Entries:     []application.SetTeamEntry{{PlayerSeasonID: fflPSID, Position: "kicks"}},
	})
	require.NoError(t, err)

	var score int
	var aflStatus string
	require.NoError(t, pool.QueryRow(ctx,
		"SELECT drv_score, drv_afl_status FROM ffl.player_match WHERE player_season_id = $1 AND club_match_id = $2",
		fflPSID, ids.homeClubMatchID).Scan(&score, &aflStatus))

	t.Run("drv_afl_status is bye", func(t *testing.T) {
		assert.Equal(t, "bye", aflStatus)
	})
	t.Run("drv_score is floor(avg_kicks)*1 = 14", func(t *testing.T) {
		assert.Equal(t, 14, score)
	})
}

// ── Test 2: star position — floor applied per stat ────────────────────────────

func TestSetTeam_ByeStarter_StarFloorPerStat(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	ctx := context.Background()

	const stubAFLPSID = 9002
	fflPSID := seedByePlayer(t, pool, ids, "Bye Star Player", stubAFLPSID)

	// floor(2.9)=2 ×5=10, floor(14.8)=14 ×1=14, floor(9.7)=9 ×1=9,
	// floor(5.6)=5 ×2=10, floor(3.9)=3 ×4=12 → total 55
	commands := setupByeCommands(t, pool, map[int]application.ByePlayerInfo{
		stubAFLPSID: {
			PlayerSeasonID: stubAFLPSID, HasBye: true, PlayedLast: true,
			AvgGoals: 2.9, AvgKicks: 14.8, AvgHandballs: 9.7, AvgMarks: 5.6, AvgTackles: 3.9,
		},
	})

	_, err := commands.SetTeam(ctx, application.SetTeamParams{
		ClubMatchID: ids.homeClubMatchID,
		Entries:     []application.SetTeamEntry{{PlayerSeasonID: fflPSID, Position: "star"}},
	})
	require.NoError(t, err)

	var score int
	require.NoError(t, pool.QueryRow(ctx,
		"SELECT drv_score FROM ffl.player_match WHERE player_season_id = $1 AND club_match_id = $2",
		fflPSID, ids.homeClubMatchID).Scan(&score))

	t.Run("star bye score floors each stat independently", func(t *testing.T) {
		assert.Equal(t, 55, score)
	})
}

// ── Test 3: ineligible bye player → SetTeam rejected ─────────────────────────

func TestSetTeam_ByeStarter_IneligibleRejected(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	ctx := context.Background()

	const stubAFLPSID = 9003
	fflPSID := seedByePlayer(t, pool, ids, "Ineligible Bye Player", stubAFLPSID)

	commands := setupByeCommands(t, pool, map[int]application.ByePlayerInfo{
		stubAFLPSID: {PlayerSeasonID: stubAFLPSID, HasBye: true, PlayedLast: false},
	})

	_, err := commands.SetTeam(ctx, application.SetTeamParams{
		ClubMatchID: ids.homeClubMatchID,
		Entries:     []application.SetTeamEntry{{PlayerSeasonID: fflPSID, Position: "kicks"}},
	})

	t.Run("returns error for ineligible bye player", func(t *testing.T) {
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ineligible")
	})
}

// ── Test 4: non-bye player unaffected — no drv_afl_status set by SetTeam ─────

func TestSetTeam_NonByePlayer_Unaffected(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	ctx := context.Background()

	const stubAFLPSID = 9004
	fflPSID := seedByePlayer(t, pool, ids, "Non-Bye Player", stubAFLPSID)

	// Stub returns has_bye=false for this player.
	commands := setupByeCommands(t, pool, map[int]application.ByePlayerInfo{
		stubAFLPSID: {PlayerSeasonID: stubAFLPSID, HasBye: false},
	})

	_, err := commands.SetTeam(ctx, application.SetTeamParams{
		ClubMatchID: ids.homeClubMatchID,
		Entries:     []application.SetTeamEntry{{PlayerSeasonID: fflPSID, Position: "kicks"}},
	})
	require.NoError(t, err)

	var aflStatus *string
	require.NoError(t, pool.QueryRow(ctx,
		"SELECT drv_afl_status FROM ffl.player_match WHERE player_season_id = $1 AND club_match_id = $2",
		fflPSID, ids.homeClubMatchID).Scan(&aflStatus))

	t.Run("drv_afl_status is null for non-bye player", func(t *testing.T) {
		assert.Nil(t, aflStatus)
	})
}

// ── Test 5: bye bench player activated → drv_score computed at activation ─────

func TestDeclareSubs_ByeBenchActivated_ScoreComputedFromAverage(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	ctx := context.Background()

	const stubAFLPSID = 9005
	fflBenchPSID := seedByePlayer(t, pool, ids, "Bye Bench Goals Player", stubAFLPSID)

	// Starter: mark DNP so the TM can sub them out.
	_, err := pool.Exec(ctx,
		"UPDATE ffl.player_match SET drv_afl_status = 'dnp', drv_score = 0 WHERE id = $1",
		ids.playerMatchID)
	require.NoError(t, err)

	// Bench bye player: backup='goals', drv_afl_status='bye', score=0.
	benchPMID := seedPM(t, pool, ids.homeClubMatchID, fflBenchPSID,
		nil, sp("bye"), sp("goals"), nil, 0)

	// floor(2.9)*5 = 2*5 = 10
	commands := setupByeCommands(t, pool, map[int]application.ByePlayerInfo{
		stubAFLPSID: {PlayerSeasonID: stubAFLPSID, HasBye: true, PlayedLast: true, AvgGoals: 2.9},
	})

	_, err = commands.DeclareSubs(ctx, ids.homeClubMatchID,
		[]domain.SubPairing{{ReplacedPMID: ids.playerMatchID, ReplacingPMID: benchPMID}},
		nil,
	)
	require.NoError(t, err)

	var benchScore int
	var benchStatus string
	require.NoError(t, pool.QueryRow(ctx,
		"SELECT drv_score, status FROM ffl.player_match WHERE id = $1", benchPMID).Scan(&benchScore, &benchStatus))

	t.Run("bench player status is subbed_in", func(t *testing.T) {
		assert.Equal(t, "subbed_in", benchStatus)
	})
	t.Run("bench bye score is floor(avgGoals)*5 = 10", func(t *testing.T) {
		assert.Equal(t, 10, benchScore)
	})

	var clubScore int
	require.NoError(t, pool.QueryRow(ctx,
		"SELECT drv_score FROM ffl.club_match WHERE id = $1", ids.homeClubMatchID).Scan(&clubScore))

	t.Run("club match score reflects activated bench bye score", func(t *testing.T) {
		assert.Equal(t, 10, clubScore)
	})
}
