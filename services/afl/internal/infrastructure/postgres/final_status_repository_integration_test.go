//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pg "xffl/services/afl/internal/infrastructure/postgres"
	"xffl/services/afl/internal/infrastructure/postgres/sqlcgen"
)

// finalStatusTestIDs holds the database IDs inserted by seedFinalStatusTestData.
type finalStatusTestIDs struct {
	playedPlayerSeasonID  int // has a player_match row in the final round
	dnpPlayerSeasonID     int // named for the club but no player_match row in the final round
	partialPlayerSeasonID int // belongs to a club whose round match is still partial
	finalRoundID          int
	partialRoundID        int
}

// seedFinalStatusTestData creates two clubs' matches in a final round (one player with
// stats, one without) plus a second round left partial, to exercise all three outcomes
// of FindFinalStatusBySeasonIDsAndRoundID: played, dnp, and omitted (not final yet).
func seedFinalStatusTestData(t *testing.T) finalStatusTestIDs {
	t.Helper()
	ctx := context.Background()
	pool := connectDB(t)

	truncateStats(t)
	t.Cleanup(func() { truncateStats(t) })

	var ids finalStatusTestIDs
	var leagueID, seasonID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.league (name) VALUES ('Final Status Test AFL') RETURNING id").Scan(&leagueID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.season (name, league_id) VALUES ('Final Status Test 2025', $1) RETURNING id",
		leagueID).Scan(&seasonID))

	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.round (name, season_id) VALUES ('Round 1', $1) RETURNING id",
		seasonID).Scan(&ids.finalRoundID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.round (name, season_id) VALUES ('Round 2', $1) RETURNING id",
		seasonID).Scan(&ids.partialRoundID))

	// Final round: one match, home club's player has stats, away club's does not.
	var homeClubID, awayClubID, homeClubSeasonID, awayClubSeasonID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.club (name) VALUES ('Home FC') RETURNING id").Scan(&homeClubID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.club (name) VALUES ('Away FC') RETURNING id").Scan(&awayClubID))
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO afl.club_season (club_id, season_id, drv_played, drv_won, drv_lost, drv_drawn, drv_for, drv_against, drv_premiership_points)
		 VALUES ($1, $2, 0, 0, 0, 0, 0, 0, 0) RETURNING id`, homeClubID, seasonID).Scan(&homeClubSeasonID))
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO afl.club_season (club_id, season_id, drv_played, drv_won, drv_lost, drv_drawn, drv_for, drv_against, drv_premiership_points)
		 VALUES ($1, $2, 0, 0, 0, 0, 0, 0, 0) RETURNING id`, awayClubID, seasonID).Scan(&awayClubSeasonID))

	var finalMatchID, homeClubMatchID, awayClubMatchID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.match (round_id, venue, start_dt, data_status) VALUES ($1, 'Test Ground', '2025-06-01 14:00:00', 'final') RETURNING id",
		ids.finalRoundID).Scan(&finalMatchID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.club_match (match_id, club_season_id, drv_score, rushed_behinds, side) VALUES ($1, $2, 80, 0, 'home') RETURNING id",
		finalMatchID, homeClubSeasonID).Scan(&homeClubMatchID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.club_match (match_id, club_season_id, drv_score, rushed_behinds, side) VALUES ($1, $2, 60, 0, 'away') RETURNING id",
		finalMatchID, awayClubSeasonID).Scan(&awayClubMatchID))

	var playedPlayerID, dnpPlayerID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player (name) VALUES ('Played Player') RETURNING id").Scan(&playedPlayerID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player (name) VALUES ('DNP Player') RETURNING id").Scan(&dnpPlayerID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player_season (player_id, club_season_id) VALUES ($1, $2) RETURNING id",
		playedPlayerID, homeClubSeasonID).Scan(&ids.playedPlayerSeasonID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player_season (player_id, club_season_id) VALUES ($1, $2) RETURNING id",
		dnpPlayerID, awayClubSeasonID).Scan(&ids.dnpPlayerSeasonID))

	_, err := pool.Exec(ctx,
		`INSERT INTO afl.player_match (club_match_id, player_season_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
		 VALUES ($1, $2, 15, 8, 4, 0, 3, 1, 0)`, homeClubMatchID, ids.playedPlayerSeasonID)
	require.NoError(t, err)
	// No player_match row for dnpPlayerSeasonID — that's the point.

	// Partial round: one club with a named player, match not yet final.
	var partialClubID, partialClubSeasonID, partialMatchID, partialClubMatchID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.club (name) VALUES ('Partial FC') RETURNING id").Scan(&partialClubID))
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO afl.club_season (club_id, season_id, drv_played, drv_won, drv_lost, drv_drawn, drv_for, drv_against, drv_premiership_points)
		 VALUES ($1, $2, 0, 0, 0, 0, 0, 0, 0) RETURNING id`, partialClubID, seasonID).Scan(&partialClubSeasonID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.match (round_id, venue, start_dt, data_status) VALUES ($1, 'Test Ground', '2025-06-08 14:00:00', 'partial') RETURNING id",
		ids.partialRoundID).Scan(&partialMatchID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.club_match (match_id, club_season_id, drv_score, rushed_behinds, side) VALUES ($1, $2, 0, 0, 'home') RETURNING id",
		partialMatchID, partialClubSeasonID).Scan(&partialClubMatchID))
	var partialPlayerID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player (name) VALUES ('Partial Player') RETURNING id").Scan(&partialPlayerID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player_season (player_id, club_season_id) VALUES ($1, $2) RETURNING id",
		partialPlayerID, partialClubSeasonID).Scan(&ids.partialPlayerSeasonID))
	_ = partialClubMatchID

	return ids
}

func TestFindFinalStatusBySeasonIDsAndRoundID(t *testing.T) {
	ids := seedFinalStatusTestData(t)
	repo := pg.NewPlayerMatchRepository(sqlcgen.New(connectDB(t)))

	t.Run("player with a player_match row in the final round is played", func(t *testing.T) {
		result, err := repo.FindFinalStatusBySeasonIDsAndRoundID(context.Background(),
			[]int{ids.playedPlayerSeasonID}, ids.finalRoundID)
		require.NoError(t, err)
		assert.Equal(t, "played", result[ids.playedPlayerSeasonID])
	})

	t.Run("player named for the club but with no player_match row in the final round is dnp", func(t *testing.T) {
		result, err := repo.FindFinalStatusBySeasonIDsAndRoundID(context.Background(),
			[]int{ids.dnpPlayerSeasonID}, ids.finalRoundID)
		require.NoError(t, err)
		assert.Equal(t, "dnp", result[ids.dnpPlayerSeasonID])
	})

	t.Run("player whose club's match in the round is not yet final is omitted", func(t *testing.T) {
		result, err := repo.FindFinalStatusBySeasonIDsAndRoundID(context.Background(),
			[]int{ids.partialPlayerSeasonID}, ids.partialRoundID)
		require.NoError(t, err)
		_, ok := result[ids.partialPlayerSeasonID]
		assert.False(t, ok, "partial-round player should not appear in the result at all")
	})

	t.Run("batch request mixing played and dnp players returns both", func(t *testing.T) {
		result, err := repo.FindFinalStatusBySeasonIDsAndRoundID(context.Background(),
			[]int{ids.playedPlayerSeasonID, ids.dnpPlayerSeasonID}, ids.finalRoundID)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "played", result[ids.playedPlayerSeasonID])
		assert.Equal(t, "dnp", result[ids.dnpPlayerSeasonID])
	})
}
