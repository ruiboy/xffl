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

// eventTestIDs holds IDs of rows inserted by seedEventTestData.
type eventTestIDs struct {
	aflRoundID        int
	aflPlayerSeasonID int
	aflClubMatchID    int
	fflRoundID        int
	fflClubMatchID    int
	fflPlayerSeasonID int
	fflClubSeasonID   int
}

func seedEventTestData(t *testing.T, pool *pgxpool.Pool) eventTestIDs {
	t.Helper()
	ctx := context.Background()
	var ids eventTestIDs

	cleanupEventTestData(ctx, t, pool)

	// AFL side: league → season → round → club → club_season → match → club_match → player → player_season
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.league (name) VALUES ('Event Test AFL') RETURNING id").Scan(new(int)))

	var aflSeasonID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.season (name, league_id) VALUES ('Event AFL 2026', (SELECT id FROM afl.league WHERE name = 'Event Test AFL')) RETURNING id",
	).Scan(&aflSeasonID))

	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.round (name, season_id) VALUES ('Round 1', $1) RETURNING id",
		aflSeasonID).Scan(&ids.aflRoundID))

	var aflClubID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.club (name) VALUES ('Event Test Club') RETURNING id").Scan(&aflClubID))

	var aflClubSeasonID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.club_season (club_id, season_id) VALUES ($1, $2) RETURNING id",
		aflClubID, aflSeasonID).Scan(&aflClubSeasonID))

	var aflMatchID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.match (round_id, venue) VALUES ($1, 'Test Ground') RETURNING id",
		ids.aflRoundID).Scan(&aflMatchID))

	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.club_match (match_id, club_season_id, side) VALUES ($1, $2, 'home') RETURNING id",
		aflMatchID, aflClubSeasonID).Scan(&ids.aflClubMatchID))

	var aflPlayerID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player (name) VALUES ('Event Test Player') RETURNING id").Scan(&aflPlayerID))

	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player_season (player_id, club_season_id) VALUES ($1, $2) RETURNING id",
		aflPlayerID, aflClubSeasonID).Scan(&ids.aflPlayerSeasonID))

	// FFL side: league → season → round (linked to AFL round) → club → club_season → match → club_match → player → player_season
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.league (name) VALUES ('Event Test FFL') RETURNING id").Scan(new(int)))

	var fflSeasonID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.season (name, league_id) VALUES ('Event FFL 2026', (SELECT id FROM ffl.league WHERE name = 'Event Test FFL')) RETURNING id",
	).Scan(&fflSeasonID))

	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.round (name, season_id, afl_round_id) VALUES ('1', $1, $2) RETURNING id",
		fflSeasonID, ids.aflRoundID).Scan(&ids.fflRoundID))

	var fflClubID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.club (name) VALUES ('Event Test Eagles') RETURNING id").Scan(&fflClubID))

	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.club_season (club_id, season_id) VALUES ($1, $2) RETURNING id",
		fflClubID, fflSeasonID).Scan(&ids.fflClubSeasonID))

	var fflMatchID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.match (round_id, match_style, venue) VALUES ($1, 'versus', 'Test Ground') RETURNING id",
		ids.fflRoundID).Scan(&fflMatchID))

	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.club_match (match_id, club_season_id, side) VALUES ($1, $2, 'home') RETURNING id",
		fflMatchID, ids.fflClubSeasonID).Scan(&ids.fflClubMatchID))

	// FFL player linked to AFL player
	var fflPlayerID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.player (afl_player_id) VALUES ($1) RETURNING id",
		aflPlayerID).Scan(&fflPlayerID))

	// FFL player season linked to AFL player season
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.player_season (player_id, club_season_id, afl_player_season_id) VALUES ($1, $2, $3) RETURNING id",
		fflPlayerID, ids.fflClubSeasonID, ids.aflPlayerSeasonID).Scan(&ids.fflPlayerSeasonID))

	// FFL player match: assigned to kicks position for this round
	_, err := pool.Exec(ctx,
		`INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status)
		 VALUES ($1, $2, 'kicks', 'named')`,
		ids.fflClubMatchID, ids.fflPlayerSeasonID)
	require.NoError(t, err)

	t.Cleanup(func() { cleanupEventTestData(context.Background(), t, pool) })
	return ids
}

func cleanupEventTestData(ctx context.Context, t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	// FFL tables first (depend on AFL player)
	fflTables := []string{
		"ffl.player_match", "ffl.player_season", "ffl.player",
		"ffl.club_match", "ffl.match", "ffl.club_season",
		"ffl.club", "ffl.round", "ffl.season", "ffl.league",
	}
	for _, table := range fflTables {
		pool.Exec(ctx, "TRUNCATE "+table+" CASCADE")
	}
	aflTables := []string{
		"afl.player_match", "afl.player_season", "afl.player",
		"afl.club_match", "afl.match", "afl.club_season",
		"afl.club", "afl.round", "afl.season", "afl.league",
	}
	for _, table := range aflTables {
		pool.Exec(ctx, "TRUNCATE "+table+" CASCADE")
	}
}

func setupCommandsWithDispatcher(t *testing.T, pool *pgxpool.Pool) (*application.Commands, *memevents.Dispatcher) {
	t.Helper()
	q := sqlcgen.New(pool)
	db := pg.NewDB(pool)
	dispatcher := memevents.New()
	commands := application.NewCommands(db, dispatcher, application.CommandsDeps{
		EventRepos: application.EventRepos{
			Rounds:        pg.NewRoundRepository(q),
			PlayerSeasons: pg.NewPlayerSeasonRepository(q),
			PlayerMatches: pg.NewPlayerMatchRepository(q),
			ClubMatches:   pg.NewClubMatchRepository(q),
		},
		PlayerLookup: &stubPlayerLookup{pool: pool},
	})
	return commands, dispatcher
}

func TestHandlePlayerMatchUpdated_scores_ffl_player_match(t *testing.T) {
	pool := connectDB(t)
	ids := seedEventTestData(t, pool)
	commands, _ := setupCommandsWithDispatcher(t, pool)
	ctx := context.Background()

	err := commands.ProcessPlayerMatchUpdated(ctx, application.PlayerMatchUpdate{
		AFLPlayerMatchID:  999,
		AFLPlayerSeasonID: ids.aflPlayerSeasonID,
		ClubMatchID:       ids.aflClubMatchID,
		RoundID:           ids.aflRoundID,
		Kicks:             20,
		Handballs:         10,
		Marks:             5,
		Tackles:           3,
		Goals:             2,
	})
	require.NoError(t, err)

	// Verify the FFL player match was scored.
	// Position is "kicks", so score = kicks * 1 = 20.
	var score int
	err = pool.QueryRow(ctx,
		"SELECT drv_score FROM ffl.player_match WHERE player_season_id = $1 AND club_match_id = $2",
		ids.fflPlayerSeasonID, ids.fflClubMatchID).Scan(&score)
	require.NoError(t, err)
	assert.Equal(t, 20, score)

	// Verify afl_player_match_id was linked.
	var aflPMID *int
	err = pool.QueryRow(ctx,
		"SELECT afl_player_match_id FROM ffl.player_match WHERE player_season_id = $1 AND club_match_id = $2",
		ids.fflPlayerSeasonID, ids.fflClubMatchID).Scan(&aflPMID)
	require.NoError(t, err)
	require.NotNil(t, aflPMID)
	assert.Equal(t, 999, *aflPMID)
}

func TestHandlePlayerMatchUpdated_ignores_unknown_player(t *testing.T) {
	pool := connectDB(t)
	seedEventTestData(t, pool)
	commands, _ := setupCommandsWithDispatcher(t, pool)
	ctx := context.Background()

	err := commands.ProcessPlayerMatchUpdated(ctx, application.PlayerMatchUpdate{
		AFLPlayerMatchID:  1000,
		AFLPlayerSeasonID: 99999, // does not exist in FFL
		ClubMatchID:       1,
		RoundID:           1,
		Kicks:             10,
	})
	assert.NoError(t, err) // should not error, just skip
}

func setupScoreCommandsWithDispatcher(t *testing.T, pool *pgxpool.Pool) (*application.ScoreCommands, *memevents.Dispatcher) {
	t.Helper()
	q := sqlcgen.New(pool)
	dispatcher := memevents.New()
	scoreCommands := application.NewScoreCommands(
		pg.NewMatchRepository(q),
		pg.NewClubMatchRepository(q),
		pg.NewClubSeasonRepository(q),
		pg.NewRoundRepository(q),
		pg.NewPlayerMatchRepository(q),
		dispatcher,
	)
	return scoreCommands, dispatcher
}

func TestHandlePlayerMatchUpdated_syncs_afl_status(t *testing.T) {
	pool := connectDB(t)
	ids := seedEventTestData(t, pool)
	commands, _ := setupCommandsWithDispatcher(t, pool)
	ctx := context.Background()

	err := commands.ProcessPlayerMatchUpdated(ctx, application.PlayerMatchUpdate{
		AFLPlayerMatchID:  999,
		AFLPlayerSeasonID: ids.aflPlayerSeasonID,
		ClubMatchID:       ids.aflClubMatchID,
		RoundID:           ids.aflRoundID,
		Status:            "named",
		Kicks:             5,
	})
	require.NoError(t, err)

	var status *string
	err = pool.QueryRow(ctx,
		"SELECT status FROM ffl.player_match WHERE player_season_id = $1 AND club_match_id = $2",
		ids.fflPlayerSeasonID, ids.fflClubMatchID).Scan(&status)
	require.NoError(t, err)
	require.NotNil(t, status)
	assert.Equal(t, "named", *status)
}

func TestHandleAflMatchFinalized_infers_player_statuses(t *testing.T) {
	pool := connectDB(t)
	ids := seedEventTestData(t, pool)
	ctx := context.Background()

	// Add a second AFL player + FFL player (linked at player level but no afl_player_match_id in player_match → dnp).
	var afl2PlayerID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player (name) VALUES ('Unlinked Test Player') RETURNING id").Scan(&afl2PlayerID))
	var unlinkedPlayerID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.player (afl_player_id) VALUES ($1) RETURNING id", afl2PlayerID).Scan(&unlinkedPlayerID))
	var unlinkedPlayerSeasonID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.player_season (player_id, club_season_id) VALUES ($1, $2) RETURNING id",
		unlinkedPlayerID, ids.fflClubSeasonID).Scan(&unlinkedPlayerSeasonID))

	_, err := pool.Exec(ctx,
		"INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status) VALUES ($1, $2, 'handballs', 'named')",
		ids.fflClubMatchID, unlinkedPlayerSeasonID)
	require.NoError(t, err)

	// The seeded player_match (from seedEventTestData) starts with status='named' and no afl_player_match_id.
	// Set afl_player_match_id = 777 to simulate it being linked (AFL stats already resolved).
	_, err = pool.Exec(ctx,
		"UPDATE ffl.player_match SET afl_player_match_id = 777 WHERE player_season_id = $1 AND club_match_id = $2",
		ids.fflPlayerSeasonID, ids.fflClubMatchID)
	require.NoError(t, err)

	scoreCommands, _ := setupScoreCommandsWithDispatcher(t, pool)

	err = scoreCommands.ProcessAFLRoundFinalized(ctx, ids.aflRoundID)
	require.NoError(t, err)

	t.Run("linked player becomes played", func(t *testing.T) {
		var status *string
		require.NoError(t, pool.QueryRow(ctx,
			"SELECT status FROM ffl.player_match WHERE player_season_id = $1 AND club_match_id = $2",
			ids.fflPlayerSeasonID, ids.fflClubMatchID).Scan(&status))
		require.NotNil(t, status)
		assert.Equal(t, "played", *status)
	})

	t.Run("unlinked named player becomes dnp", func(t *testing.T) {
		var status *string
		require.NoError(t, pool.QueryRow(ctx,
			"SELECT status FROM ffl.player_match WHERE player_season_id = $1 AND club_match_id = $2",
			unlinkedPlayerSeasonID, ids.fflClubMatchID).Scan(&status))
		require.NotNil(t, status)
		assert.Equal(t, "dnp", *status)
	})
}

func TestHandlePlayerMatchUpdated_multiple_ffl_clubs(t *testing.T) {
	pool := connectDB(t)
	ids := seedEventTestData(t, pool)
	commands, _ := setupCommandsWithDispatcher(t, pool)
	ctx := context.Background()

	// Add a second FFL club with the same AFL player.
	var secondClubSeasonID, secondClubMatchID, secondPlayerSeasonID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.club (name) VALUES ('Event Second Club') RETURNING id").Scan(new(int)))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.club_season (club_id, season_id) VALUES ((SELECT id FROM ffl.club WHERE name = 'Event Second Club'), (SELECT season_id FROM ffl.round WHERE id = $1)) RETURNING id",
		ids.fflRoundID).Scan(&secondClubSeasonID))

	var secondMatchID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.match (round_id, match_style, venue) VALUES ($1, 'versus', 'Other Ground') RETURNING id",
		ids.fflRoundID).Scan(&secondMatchID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.club_match (match_id, club_season_id, side) VALUES ($1, $2, 'home') RETURNING id",
		secondMatchID, secondClubSeasonID).Scan(&secondClubMatchID))

	// Reuse the FFL player created by seedEventTestData (linked to the same AFL player).
	var secondPlayerID int
	require.NoError(t, pool.QueryRow(ctx,
		"SELECT id FROM ffl.player WHERE afl_player_id = (SELECT player_id FROM afl.player_season WHERE id = $1)",
		ids.aflPlayerSeasonID).Scan(&secondPlayerID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.player_season (player_id, club_season_id, afl_player_season_id) VALUES ($1, $2, $3) RETURNING id",
		secondPlayerID, secondClubSeasonID, ids.aflPlayerSeasonID).Scan(&secondPlayerSeasonID))

	// Assign to goals position in second club.
	_, err := pool.Exec(ctx,
		`INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status) VALUES ($1, $2, 'goals', 'named')`,
		secondClubMatchID, secondPlayerSeasonID)
	require.NoError(t, err)

	err = commands.ProcessPlayerMatchUpdated(ctx, application.PlayerMatchUpdate{
		AFLPlayerMatchID:  888,
		AFLPlayerSeasonID: ids.aflPlayerSeasonID,
		ClubMatchID:       ids.aflClubMatchID,
		RoundID:           ids.aflRoundID,
		Kicks:             15,
		Handballs:         8,
		Marks:             4,
		Tackles:           6,
		Goals:             3,
	})
	require.NoError(t, err)

	// First club: kicks position → score = kicks * 1 = 15.
	var score1 int
	require.NoError(t, pool.QueryRow(ctx,
		"SELECT drv_score FROM ffl.player_match WHERE player_season_id = $1 AND club_match_id = $2",
		ids.fflPlayerSeasonID, ids.fflClubMatchID).Scan(&score1))
	assert.Equal(t, 15, score1)

	// Second club: goals position → score = goals * 5 = 15.
	var score2 int
	require.NoError(t, pool.QueryRow(ctx,
		"SELECT drv_score FROM ffl.player_match WHERE player_season_id = $1 AND club_match_id = $2",
		secondPlayerSeasonID, secondClubMatchID).Scan(&score2))
	assert.Equal(t, 15, score2)
}
