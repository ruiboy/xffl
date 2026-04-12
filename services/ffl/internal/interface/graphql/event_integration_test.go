package graphql_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	contractevents "xffl/contracts/events"
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
		"INSERT INTO afl.club_match (match_id, club_season_id) VALUES ($1, $2) RETURNING id",
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
		"INSERT INTO ffl.club_match (match_id, club_season_id) VALUES ($1, $2) RETURNING id",
		fflMatchID, ids.fflClubSeasonID).Scan(&ids.fflClubMatchID))

	// FFL player linked to AFL player
	var fflPlayerID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.player (drv_name, afl_player_id) VALUES ('Event Test Player', $1) RETURNING id",
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
	commands := application.NewCommands(db, dispatcher, application.EventRepos{
		Rounds:        pg.NewRoundRepository(q),
		PlayerSeasons: pg.NewPlayerSeasonRepository(q),
		PlayerMatches: pg.NewPlayerMatchRepository(q),
	})
	return commands, dispatcher
}

func TestHandlePlayerMatchUpdated_scores_ffl_player_match(t *testing.T) {
	pool := connectDB(t)
	ids := seedEventTestData(t, pool)
	commands, _ := setupCommandsWithDispatcher(t, pool)
	ctx := context.Background()

	// Simulate AFL publishing a PlayerMatchUpdated event.
	payload, err := json.Marshal(contractevents.PlayerMatchUpdatedPayload{
		PlayerMatchID:  999, // AFL player_match ID (arbitrary, used for linking)
		PlayerSeasonID: ids.aflPlayerSeasonID,
		ClubMatchID:    ids.aflClubMatchID,
		RoundID:        ids.aflRoundID,
		Kicks:          20,
		Handballs:      10,
		Marks:          5,
		Hitouts:        0,
		Tackles:        3,
		Goals:          2,
		Behinds:        1,
	})
	require.NoError(t, err)

	// Handle the event.
	err = commands.HandlePlayerMatchUpdated(ctx, payload)
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

	// Event for an AFL player_season not in any FFL squad.
	payload, err := json.Marshal(contractevents.PlayerMatchUpdatedPayload{
		PlayerMatchID:  1000,
		PlayerSeasonID: 99999, // does not exist in FFL
		ClubMatchID:    1,
		RoundID:        1,
		Kicks:          10,
	})
	require.NoError(t, err)

	err = commands.HandlePlayerMatchUpdated(ctx, payload)
	assert.NoError(t, err) // should not error, just skip
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
		"INSERT INTO ffl.club_match (match_id, club_season_id) VALUES ($1, $2) RETURNING id",
		secondMatchID, secondClubSeasonID).Scan(&secondClubMatchID))

	// Second FFL player for the same AFL player
	var secondPlayerID int
	require.NoError(t, pool.QueryRow(ctx,
		"SELECT id FROM ffl.player WHERE drv_name = 'Event Test Player'").Scan(&secondPlayerID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.player_season (player_id, club_season_id, afl_player_season_id) VALUES ($1, $2, $3) RETURNING id",
		secondPlayerID, secondClubSeasonID, ids.aflPlayerSeasonID).Scan(&secondPlayerSeasonID))

	// Assign to goals position in second club.
	_, err := pool.Exec(ctx,
		`INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status) VALUES ($1, $2, 'goals', 'named')`,
		secondClubMatchID, secondPlayerSeasonID)
	require.NoError(t, err)

	// Fire event.
	payload, err := json.Marshal(contractevents.PlayerMatchUpdatedPayload{
		PlayerMatchID:  888,
		PlayerSeasonID: ids.aflPlayerSeasonID,
		ClubMatchID:    ids.aflClubMatchID,
		RoundID:        ids.aflRoundID,
		Kicks:          15,
		Handballs:      8,
		Marks:          4,
		Hitouts:        0,
		Tackles:        6,
		Goals:          3,
		Behinds:        2,
	})
	require.NoError(t, err)

	err = commands.HandlePlayerMatchUpdated(ctx, payload)
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
