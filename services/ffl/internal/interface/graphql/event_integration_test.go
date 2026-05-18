//go:build integration

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
	"xffl/services/ffl/internal/domain"
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
		"INSERT INTO ffl.season (name, league_id, afl_season_id) VALUES ('Event FFL 2026', (SELECT id FROM ffl.league WHERE name = 'Event Test FFL'), $1) RETURNING id",
		aflSeasonID).Scan(&fflSeasonID))

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
	commands := application.NewCommands(
		db,
		dispatcher,
		&stubPlayerLookup{pool: pool},
		pg.NewMatchRepository(q),
		pg.NewClubMatchRepository(q),
		pg.NewClubSeasonRepository(q),
		pg.NewRoundRepository(q),
		pg.NewPlayerMatchRepository(q),
		pg.NewPlayerSeasonRepository(q),
	)
	return commands, dispatcher
}

func TestHandleAflPlayerMatchUpdated_scores_ffl_player_match(t *testing.T) {
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

func TestHandleAflPlayerMatchUpdated_ignores_unknown_player(t *testing.T) {
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

func TestHandleAflPlayerMatchUpdated_multiple_ffl_clubs(t *testing.T) {
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

// addExtraPlayerToClubMatch inserts a new AFL player + player_season + FFL player + player_season +
// player_match into the given ffl.club_match. The new player's afl_player_season_id is returned.
func addExtraPlayerToClubMatch(t *testing.T, pool *pgxpool.Pool, ids eventTestIDs, position string, drvAFLStatus *string) (aflPlayerSeasonID, fflPlayerSeasonID int) {
	t.Helper()
	ctx := context.Background()

	// Derive the AFL club_season_id from the existing afl.player_season.
	var aflClubSeasonID int
	require.NoError(t, pool.QueryRow(ctx,
		"SELECT club_season_id FROM afl.player_season WHERE id = $1",
		ids.aflPlayerSeasonID).Scan(&aflClubSeasonID))

	var aflPlayerID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player (name) VALUES ('Extra Event Player') RETURNING id").Scan(&aflPlayerID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player_season (player_id, club_season_id) VALUES ($1, $2) RETURNING id",
		aflPlayerID, aflClubSeasonID).Scan(&aflPlayerSeasonID))

	var fflPlayerID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.player (afl_player_id) VALUES ($1) RETURNING id", aflPlayerID).Scan(&fflPlayerID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.player_season (player_id, club_season_id, afl_player_season_id) VALUES ($1, $2, $3) RETURNING id",
		fflPlayerID, ids.fflClubSeasonID, aflPlayerSeasonID).Scan(&fflPlayerSeasonID))

	if drvAFLStatus == nil {
		_, err := pool.Exec(ctx,
			"INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status) VALUES ($1, $2, $3, 'named')",
			ids.fflClubMatchID, fflPlayerSeasonID, position)
		require.NoError(t, err)
	} else {
		_, err := pool.Exec(ctx,
			"INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status, drv_afl_status) VALUES ($1, $2, $3, 'named', $4)",
			ids.fflClubMatchID, fflPlayerSeasonID, position, *drvAFLStatus)
		require.NoError(t, err)
	}

	return aflPlayerSeasonID, fflPlayerSeasonID
}

func strptr(s string) *string { return &s }

// ────────────────────────────────────────────────────────────────────────────
// AllAFLStatusesFinal unit-style tests (backed by real DB)
// ────────────────────────────────────────────────────────────────────────────

func TestAllAFLStatusesFinal(t *testing.T) {
	cases := []struct {
		name      string
		status1   *string // nil = NULL drv_afl_status for the seeded player
		status2   *string // non-nil → add a second player with this status (nil = NULL)
		wantFinal bool
	}{
		{"all played (one player)",
			strptr("played"), nil, true},
		{"all dnp (one player)",
			strptr("dnp"), nil, true},
		{"mixed played and dnp",
			strptr("played"), strptr("dnp"), true},
		{"null present",
			nil, nil, false},
		{"playing present",
			strptr("playing"), nil, false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			pool := connectDB(t)
			ids := seedEventTestData(t, pool)
			commands, _ := setupCommandsWithDispatcher(t, pool)
			ctx := context.Background()

			if tc.status1 == nil {
				_, err := pool.Exec(ctx,
					"UPDATE ffl.player_match SET drv_afl_status = NULL WHERE club_match_id = $1",
					ids.fflClubMatchID)
				require.NoError(t, err)
			} else {
				_, err := pool.Exec(ctx,
					"UPDATE ffl.player_match SET drv_afl_status = $1 WHERE club_match_id = $2",
					*tc.status1, ids.fflClubMatchID)
				require.NoError(t, err)
			}

			if tc.status2 != nil {
				addExtraPlayerToClubMatch(t, pool, ids, "handballs", tc.status2)
			}

			got, err := commands.AllAFLStatusesFinal(ctx, ids.fflClubMatchID)
			require.NoError(t, err)
			assert.Equal(t, tc.wantFinal, got, "AllAFLStatusesFinal mismatch")
		})
	}
}

// ────────────────────────────────────────────────────────────────────────────
// ProcessAFLMatchUpdated integration tests
// ────────────────────────────────────────────────────────────────────────────

func TestProcessAFLMatchUpdated_partial_sets_playing_for_correct_players(t *testing.T) {
	pool := connectDB(t)
	ids := seedEventTestData(t, pool)
	commands, _ := setupCommandsWithDispatcher(t, pool)
	ctx := context.Background()

	// Add a second player linked to a different AFL player_season (not in the status map).
	otherAFLPSID, otherFflPSID := addExtraPlayerToClubMatch(t, pool, ids, "handballs", nil)
	_ = otherFflPSID

	err := commands.ProcessAFLMatchUpdated(ctx, contractevents.AflMatchUpdatedPayload{
		RoundID:     ids.aflRoundID,
		MatchStatus: "partial",
		PlayerSeasonIDStatusMap: map[int]string{
			ids.aflPlayerSeasonID: "playing",
			// otherAFLPSID is intentionally absent — player should remain unaffected.
		},
	})
	require.NoError(t, err)

	t.Run("player in status map gets drv_afl_status=playing", func(t *testing.T) {
		var got *string
		require.NoError(t, pool.QueryRow(ctx,
			"SELECT drv_afl_status FROM ffl.player_match WHERE player_season_id = $1 AND club_match_id = $2",
			ids.fflPlayerSeasonID, ids.fflClubMatchID).Scan(&got))
		require.NotNil(t, got)
		assert.Equal(t, "playing", *got)
	})

	t.Run("player absent from status map remains unaffected (null)", func(t *testing.T) {
		var got *string
		// Find the player_match for the extra player in the FFL club_match.
		require.NoError(t, pool.QueryRow(ctx,
			`SELECT pm.drv_afl_status
			 FROM ffl.player_match pm
			 JOIN ffl.player_season ps ON ps.id = pm.player_season_id
			 WHERE ps.afl_player_season_id = $1 AND pm.club_match_id = $2`,
			otherAFLPSID, ids.fflClubMatchID).Scan(&got))
		assert.Nil(t, got, "player not in the status map should still have drv_afl_status=NULL")
	})
}

func TestProcessAFLMatchUpdated_final_sets_played_dnp_and_emits_finalized(t *testing.T) {
	pool := connectDB(t)
	ids := seedEventTestData(t, pool)
	ctx := context.Background()

	// Add a second player who will receive "dnp" (not in AFL match stats).
	dnpAFLPSID, _ := addExtraPlayerToClubMatch(t, pool, ids, "handballs", nil)

	// Mark the FFL club_match as final so both-axes check fires.
	_, err := pool.Exec(ctx,
		"UPDATE ffl.club_match SET data_status = 'final' WHERE id = $1", ids.fflClubMatchID)
	require.NoError(t, err)

	commands, dispatcher := setupCommandsWithDispatcher(t, pool)

	var finalizedClubMatchID int
	dispatcher.Subscribe(contractevents.FflClubMatchScoreFinalized, func(_ context.Context, payload []byte) error {
		var p contractevents.FflClubMatchScoreFinalizedPayload
		if err := unmarshalJSON(payload, &p); err != nil {
			return err
		}
		finalizedClubMatchID = p.ClubMatchID
		return nil
	})

	err = commands.ProcessAFLMatchUpdated(ctx, contractevents.AflMatchUpdatedPayload{
		RoundID:     ids.aflRoundID,
		MatchStatus: "final",
		PlayerSeasonIDStatusMap: map[int]string{
			ids.aflPlayerSeasonID: "played",
			dnpAFLPSID:            "dnp",
		},
	})
	require.NoError(t, err)

	t.Run("played player gets drv_afl_status=played", func(t *testing.T) {
		var got *string
		require.NoError(t, pool.QueryRow(ctx,
			"SELECT drv_afl_status FROM ffl.player_match WHERE player_season_id = $1 AND club_match_id = $2",
			ids.fflPlayerSeasonID, ids.fflClubMatchID).Scan(&got))
		require.NotNil(t, got)
		assert.Equal(t, "played", *got)
	})

	t.Run("unplayed player gets drv_afl_status=dnp", func(t *testing.T) {
		var got *string
		require.NoError(t, pool.QueryRow(ctx,
			`SELECT pm.drv_afl_status
			 FROM ffl.player_match pm
			 JOIN ffl.player_season ps ON ps.id = pm.player_season_id
			 WHERE ps.afl_player_season_id = $1 AND pm.club_match_id = $2`,
			dnpAFLPSID, ids.fflClubMatchID).Scan(&got))
		require.NotNil(t, got)
		assert.Equal(t, "dnp", *got)
	})

	t.Run("FFL.ClubMatchScoreFinalized emitted because FFL team is final", func(t *testing.T) {
		assert.Equal(t, ids.fflClubMatchID, finalizedClubMatchID)
	})
}

// ────────────────────────────────────────────────────────────────────────────
// ProcessFflClubMatchUpdated integration tests
// ────────────────────────────────────────────────────────────────────────────

func TestProcessFflClubMatchUpdated_finalization_axes(t *testing.T) {
	t.Run("FFL final but AFL not final: no ClubMatchScoreFinalized", func(t *testing.T) {
		pool := connectDB(t)
		ids := seedEventTestData(t, pool)
		ctx := context.Background()

		// Leave drv_afl_status NULL (AFL not yet final).
		commands, dispatcher := setupCommandsWithDispatcher(t, pool)

		var emitted bool
		dispatcher.Subscribe(contractevents.FflClubMatchScoreFinalized, func(_ context.Context, _ []byte) error {
			emitted = true
			return nil
		})

		// FFL finalizes their team.
		err := commands.ProcessFflClubMatchUpdated(ctx, ids.fflClubMatchID, 0, domain.ClubMatchDataFinal)
		require.NoError(t, err)
		assert.False(t, emitted, "FFL.ClubMatchScoreFinalized must not fire when AFL status is not yet final")
	})

	t.Run("both axes final: ClubMatchScoreFinalized emitted", func(t *testing.T) {
		pool := connectDB(t)
		ids := seedEventTestData(t, pool)
		ctx := context.Background()

		// Set AFL status to final for the single seeded player.
		_, err := pool.Exec(ctx,
			"UPDATE ffl.player_match SET drv_afl_status = 'played' WHERE club_match_id = $1",
			ids.fflClubMatchID)
		require.NoError(t, err)

		// Look up the FFL match ID so emitClubMatchScoreFinalized can store it.
		var fflMatchID int
		require.NoError(t, pool.QueryRow(ctx,
			"SELECT match_id FROM ffl.club_match WHERE id = $1", ids.fflClubMatchID).Scan(&fflMatchID))

		commands, dispatcher := setupCommandsWithDispatcher(t, pool)

		var finalizedClubMatchID int
		dispatcher.Subscribe(contractevents.FflClubMatchScoreFinalized, func(_ context.Context, payload []byte) error {
			var p contractevents.FflClubMatchScoreFinalizedPayload
			if err := unmarshalJSON(payload, &p); err != nil {
				return err
			}
			finalizedClubMatchID = p.ClubMatchID
			return nil
		})

		err = commands.ProcessFflClubMatchUpdated(ctx, ids.fflClubMatchID, fflMatchID, domain.ClubMatchDataFinal)
		require.NoError(t, err)
		assert.Equal(t, ids.fflClubMatchID, finalizedClubMatchID, "FFL.ClubMatchScoreFinalized must fire when both axes are final")
	})
}

// ────────────────────────────────────────────────────────────────────────────
// Full event chain regression test
// ────────────────────────────────────────────────────────────────────────────

// TestEventChain_fullRoundTrip simulates the complete round lifecycle:
//  1. AFL stats imported (partial) → player score calculated, drv_afl_status=playing
//  2. AFL match finalised → played set; FFL.ClubMatchScoreFinalized NOT yet (FFL team not final)
//  3. FFL home team finalised → both axes final → FFL.ClubMatchScoreFinalized emitted (home)
//  4. FFL away team finalised → FFL.ClubMatchScoreFinalized emitted (away)
//  5. Both club_matches final → FFL.MatchScoreFinalized emitted
//  6. ProcessFflMatchScoreFinalized → match result derived (home_win)
func TestEventChain_fullRoundTrip(t *testing.T) {
	pool := connectDB(t)
	ids := seedEventTestData(t, pool)
	ctx := context.Background()

	// seedEventTestData only creates the home club_match. Create an away club_match so
	// ProcessFflClubMatchScoreFinalized can count two final club_matches and fire
	// FflMatchScoreFinalized.
	var fflMatchID int
	require.NoError(t, pool.QueryRow(ctx,
		"SELECT match_id FROM ffl.club_match WHERE id = $1", ids.fflClubMatchID).Scan(&fflMatchID))

	var awayClubSeasonID, awayClubMatchID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.club (name) VALUES ('Regression Away Club') RETURNING id").Scan(new(int)))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.club_season (club_id, season_id) VALUES ((SELECT id FROM ffl.club WHERE name = 'Regression Away Club'), (SELECT season_id FROM ffl.round WHERE id = $1)) RETURNING id",
		ids.fflRoundID).Scan(&awayClubSeasonID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.club_match (match_id, club_season_id, side) VALUES ($1, $2, 'away') RETURNING id",
		fflMatchID, awayClubSeasonID).Scan(&awayClubMatchID))

	var scoreFinalizedCount, matchFinalizedCount int
	commands, dispatcher := setupCommandsWithDispatcher(t, pool)
	dispatcher.Subscribe(contractevents.FflClubMatchScoreFinalized, func(_ context.Context, _ []byte) error {
		scoreFinalizedCount++
		return nil
	})
	dispatcher.Subscribe(contractevents.FflMatchScoreFinalized, func(_ context.Context, _ []byte) error {
		matchFinalizedCount++
		return nil
	})

	// Step 1: partial AFL stats → player score set, drv_afl_status=playing.
	require.NoError(t, commands.ProcessPlayerMatchUpdated(ctx, application.PlayerMatchUpdate{
		AFLPlayerMatchID: 777, AFLPlayerSeasonID: ids.aflPlayerSeasonID,
		ClubMatchID: ids.aflClubMatchID, RoundID: ids.aflRoundID,
		Kicks: 20, Goals: 2,
	}))
	require.NoError(t, commands.ProcessAFLMatchUpdated(ctx, contractevents.AflMatchUpdatedPayload{
		RoundID: ids.aflRoundID, MatchStatus: "partial",
		PlayerSeasonIDStatusMap: map[int]string{ids.aflPlayerSeasonID: "playing"},
	}))

	var drvStatus *string
	require.NoError(t, pool.QueryRow(ctx,
		"SELECT drv_afl_status FROM ffl.player_match WHERE player_season_id = $1 AND club_match_id = $2",
		ids.fflPlayerSeasonID, ids.fflClubMatchID).Scan(&drvStatus))
	require.NotNil(t, drvStatus)
	assert.Equal(t, "playing", *drvStatus, "[step 1] drv_afl_status=playing after partial import")
	assert.Equal(t, 0, scoreFinalizedCount, "[step 1] no finalization after partial")

	// Step 2: AFL match finalised → played set; FFL team still submitted → no finalization.
	require.NoError(t, commands.ProcessAFLMatchUpdated(ctx, contractevents.AflMatchUpdatedPayload{
		RoundID: ids.aflRoundID, MatchStatus: "final",
		PlayerSeasonIDStatusMap: map[int]string{ids.aflPlayerSeasonID: "played"},
	}))

	require.NoError(t, pool.QueryRow(ctx,
		"SELECT drv_afl_status FROM ffl.player_match WHERE player_season_id = $1 AND club_match_id = $2",
		ids.fflPlayerSeasonID, ids.fflClubMatchID).Scan(&drvStatus))
	require.NotNil(t, drvStatus)
	assert.Equal(t, "played", *drvStatus, "[step 2] drv_afl_status=played after AFL final")
	assert.Equal(t, 0, scoreFinalizedCount, "[step 2] no finalization — FFL team not yet final")

	// Step 3: home FFL team finalised → both axes final → FFL.ClubMatchScoreFinalized (home).
	_, err := pool.Exec(ctx, "UPDATE ffl.club_match SET data_status = 'final' WHERE id = $1", ids.fflClubMatchID)
	require.NoError(t, err)
	require.NoError(t, commands.ProcessFflClubMatchUpdated(ctx, ids.fflClubMatchID, fflMatchID, domain.ClubMatchDataFinal))
	assert.Equal(t, 1, scoreFinalizedCount, "[step 3] FFL.ClubMatchScoreFinalized for home club_match")

	// Step 4: away FFL team finalised (no players → AllAFLStatusesFinal=true immediately).
	_, err = pool.Exec(ctx, "UPDATE ffl.club_match SET data_status = 'final' WHERE id = $1", awayClubMatchID)
	require.NoError(t, err)
	require.NoError(t, commands.ProcessFflClubMatchUpdated(ctx, awayClubMatchID, fflMatchID, domain.ClubMatchDataFinal))
	assert.Equal(t, 2, scoreFinalizedCount, "[step 4] FFL.ClubMatchScoreFinalized for away club_match")

	// Step 5: both club_matches final → ProcessFflClubMatchScoreFinalized emits FflMatchScoreFinalized.
	require.NoError(t, commands.ProcessFflClubMatchScoreFinalized(ctx, ids.fflClubMatchID, fflMatchID))
	assert.Equal(t, 1, matchFinalizedCount, "[step 5] FFL.MatchScoreFinalized fired")

	// Step 6: match result derived from stored scores.
	require.NoError(t, commands.ProcessFflMatchScoreFinalized(ctx, fflMatchID, ids.fflRoundID))

	var drvResult *string
	require.NoError(t, pool.QueryRow(ctx,
		"SELECT drv_result FROM ffl.match WHERE id = $1", fflMatchID).Scan(&drvResult))

	t.Run("match result is home_win (home scored 20, away scored 0)", func(t *testing.T) {
		require.NotNil(t, drvResult)
		assert.Equal(t, "home_win", *drvResult)
	})
	t.Run("FFL.ClubMatchScoreFinalized fired for both club_matches", func(t *testing.T) {
		assert.Equal(t, 2, scoreFinalizedCount)
	})
	t.Run("FFL.MatchScoreFinalized fired once both club_matches finalized", func(t *testing.T) {
		assert.Equal(t, 1, matchFinalizedCount)
	})
}

func unmarshalJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
