//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/afl/internal/domain"
	pg "xffl/services/afl/internal/infrastructure/postgres"
)

// statsTestIDs holds the database IDs inserted by seedStatsTestData.
type statsTestIDs struct {
	playerSeasonID int
	round1ID       int
	round2ID       int
	round3ID       int
}

// seedStatsTestData creates a minimal but multi-round fixture:
//   - 3 rounds, each with one final match
//   - 1 player with one player_season across all three rounds
//   - player_match per round: (goals=2,kicks=10), (goals=0,kicks=8), (goals=1,kicks=12)
//     ordered chronologically Round1 < Round2 < Round3
func seedStatsTestData(t *testing.T) statsTestIDs {
	t.Helper()
	ctx := context.Background()
	pool := connectDB(t)

	// Truncate in FK order to isolate each test run.
	truncateStats(t)
	t.Cleanup(func() { truncateStats(t) })

	var ids statsTestIDs
	var leagueID, seasonID int
	var clubID, clubSeasonID int

	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.league (name) VALUES ('Stats Test AFL') RETURNING id").Scan(&leagueID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.season (name, league_id) VALUES ('Stats Test 2025', $1) RETURNING id",
		leagueID).Scan(&seasonID))

	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.round (name, season_id) VALUES ('Round 1', $1) RETURNING id",
		seasonID).Scan(&ids.round1ID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.round (name, season_id) VALUES ('Round 2', $1) RETURNING id",
		seasonID).Scan(&ids.round2ID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.round (name, season_id) VALUES ('Round 3', $1) RETURNING id",
		seasonID).Scan(&ids.round3ID))

	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.club (name) VALUES ('Stats FC') RETURNING id").Scan(&clubID))
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO afl.club_season (club_id, season_id, drv_played, drv_won, drv_lost, drv_drawn, drv_for, drv_against, drv_premiership_points)
		 VALUES ($1, $2, 3, 2, 1, 0, 300, 250, 8) RETURNING id`,
		clubID, seasonID).Scan(&clubSeasonID))

	var playerID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player (name) VALUES ('Stats Player') RETURNING id").Scan(&playerID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player_season (player_id, club_season_id) VALUES ($1, $2) RETURNING id",
		playerID, clubSeasonID).Scan(&ids.playerSeasonID))

	// Three final matches on different dates so chronological ordering is unambiguous.
	type matchRow struct {
		roundID int
		startDt string
		goals   int
		kicks   int
	}
	matches := []matchRow{
		{ids.round1ID, "2025-06-01 14:00:00", 2, 10},
		{ids.round2ID, "2025-06-08 14:00:00", 0, 8},
		{ids.round3ID, "2025-06-15 14:00:00", 1, 12},
	}
	for _, m := range matches {
		var matchID, clubMatchID int
		require.NoError(t, pool.QueryRow(ctx,
			"INSERT INTO afl.match (round_id, venue, start_dt, data_status) VALUES ($1, 'Test Ground', $2, 'final') RETURNING id",
			m.roundID, m.startDt).Scan(&matchID))
		require.NoError(t, pool.QueryRow(ctx,
			"INSERT INTO afl.club_match (match_id, club_season_id, drv_score, rushed_behinds, side) VALUES ($1, $2, 80, 0, 'home') RETURNING id",
			matchID, clubSeasonID).Scan(&clubMatchID))
		_, err := pool.Exec(ctx,
			`INSERT INTO afl.player_match (club_match_id, player_season_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
			 VALUES ($1, $2, $3, 4, 3, 0, 2, $4, 0)`,
			clubMatchID, ids.playerSeasonID, m.kicks, m.goals)
		require.NoError(t, err)
	}

	return ids
}

func truncateStats(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	pool := connectDB(t)
	for _, table := range []string{
		"afl.player_match",
		"afl.player_season",
		"afl.player",
		"afl.club_match",
		"afl.match",
		"afl.club_season",
		"afl.club",
		"afl.round",
		"afl.season",
		"afl.league",
	} {
		_, err := pool.Exec(ctx, fmt.Sprintf("TRUNCATE %s CASCADE", table))
		require.NoError(t, err, "truncate %s", table)
	}
}

func TestGetSeasonStats_NoParams(t *testing.T) {
	ids := seedStatsTestData(t)
	repo := pg.NewPlayerSeasonStatsRepository(connectDB(t))

	rows, err := repo.GetSeasonStats(context.Background(), domain.PlayerSeasonStatsParams{
		PlayerSeasonIDs: []int{ids.playerSeasonID},
	})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	s := rows[0]

	t.Run("games counts all three final matches", func(t *testing.T) {
		assert.Equal(t, 3, s.Games)
	})
	t.Run("goals averaged across all three matches", func(t *testing.T) {
		// (2 + 0 + 1) / 3 = 1.0
		assert.InDelta(t, 1.0, s.Goals, 0.001)
	})
	t.Run("kicks averaged across all three matches", func(t *testing.T) {
		// (10 + 8 + 12) / 3 = 10.0
		assert.InDelta(t, 10.0, s.Kicks, 0.001)
	})
}

func TestGetSeasonStats_UpToRoundId(t *testing.T) {
	ids := seedStatsTestData(t)
	repo := pg.NewPlayerSeasonStatsRepository(connectDB(t))

	t.Run("upToRoundId=round2 returns only round1 match", func(t *testing.T) {
		rows, err := repo.GetSeasonStats(context.Background(), domain.PlayerSeasonStatsParams{
			PlayerSeasonIDs: []int{ids.playerSeasonID},
			UpToRoundID:     &ids.round2ID,
		})
		require.NoError(t, err)
		require.Len(t, rows, 1)
		s := rows[0]
		assert.Equal(t, 1, s.Games)
		assert.InDelta(t, 2.0, s.Goals, 0.001)
		assert.InDelta(t, 10.0, s.Kicks, 0.001)
	})

	t.Run("upToRoundId=round3 returns round1 and round2 matches", func(t *testing.T) {
		rows, err := repo.GetSeasonStats(context.Background(), domain.PlayerSeasonStatsParams{
			PlayerSeasonIDs: []int{ids.playerSeasonID},
			UpToRoundID:     &ids.round3ID,
		})
		require.NoError(t, err)
		require.Len(t, rows, 1)
		s := rows[0]
		assert.Equal(t, 2, s.Games)
		// (2 + 0) / 2 = 1.0
		assert.InDelta(t, 1.0, s.Goals, 0.001)
		// (10 + 8) / 2 = 9.0
		assert.InDelta(t, 9.0, s.Kicks, 0.001)
	})

	t.Run("upToRoundId=round1 returns empty (no matches before round1)", func(t *testing.T) {
		rows, err := repo.GetSeasonStats(context.Background(), domain.PlayerSeasonStatsParams{
			PlayerSeasonIDs: []int{ids.playerSeasonID},
			UpToRoundID:     &ids.round1ID,
		})
		require.NoError(t, err)
		assert.Empty(t, rows)
	})
}

func TestGetSeasonStats_LastN(t *testing.T) {
	ids := seedStatsTestData(t)
	repo := pg.NewPlayerSeasonStatsRepository(connectDB(t))

	lastN1 := 1
	lastN2 := 2

	t.Run("lastN=1 returns the most recent match only", func(t *testing.T) {
		rows, err := repo.GetSeasonStats(context.Background(), domain.PlayerSeasonStatsParams{
			PlayerSeasonIDs: []int{ids.playerSeasonID},
			LastN:           &lastN1,
		})
		require.NoError(t, err)
		require.Len(t, rows, 1)
		s := rows[0]
		assert.Equal(t, 1, s.Games)
		// Round 3 match: goals=1, kicks=12
		assert.InDelta(t, 1.0, s.Goals, 0.001)
		assert.InDelta(t, 12.0, s.Kicks, 0.001)
	})

	t.Run("lastN=2 returns the two most recent matches", func(t *testing.T) {
		rows, err := repo.GetSeasonStats(context.Background(), domain.PlayerSeasonStatsParams{
			PlayerSeasonIDs: []int{ids.playerSeasonID},
			LastN:           &lastN2,
		})
		require.NoError(t, err)
		require.Len(t, rows, 1)
		s := rows[0]
		assert.Equal(t, 2, s.Games)
		// Round 2 + Round 3: (0+1)/2 = 0.5 goals, (8+12)/2 = 10.0 kicks
		assert.InDelta(t, 0.5, s.Goals, 0.001)
		assert.InDelta(t, 10.0, s.Kicks, 0.001)
	})
}

func TestGetSeasonStats_Combined(t *testing.T) {
	ids := seedStatsTestData(t)
	repo := pg.NewPlayerSeasonStatsRepository(connectDB(t))

	lastN1 := 1

	t.Run("upToRoundId=round3 + lastN=1 returns the most recent match before round3", func(t *testing.T) {
		rows, err := repo.GetSeasonStats(context.Background(), domain.PlayerSeasonStatsParams{
			PlayerSeasonIDs: []int{ids.playerSeasonID},
			UpToRoundID:     &ids.round3ID,
			LastN:           &lastN1,
		})
		require.NoError(t, err)
		require.Len(t, rows, 1)
		s := rows[0]
		assert.Equal(t, 1, s.Games)
		// Most recent before Round 3 is Round 2: goals=0, kicks=8
		assert.InDelta(t, 0.0, s.Goals, 0.001)
		assert.InDelta(t, 8.0, s.Kicks, 0.001)
	})
}

func TestGetSeasonStats_UnknownPlayerSeason(t *testing.T) {
	seedStatsTestData(t)
	repo := pg.NewPlayerSeasonStatsRepository(connectDB(t))

	rows, err := repo.GetSeasonStats(context.Background(), domain.PlayerSeasonStatsParams{
		PlayerSeasonIDs: []int{999999},
	})
	require.NoError(t, err)

	t.Run("unknown player season returns empty slice", func(t *testing.T) {
		assert.Empty(t, rows)
	})
}

func TestGetSeasonStats_NonFinalMatchExcluded(t *testing.T) {
	ids := seedStatsTestData(t)
	pool := connectDB(t)
	repo := pg.NewPlayerSeasonStatsRepository(pool)
	ctx := context.Background()

	// Add a 4th match in a new round with data_status='partial' and inflated stats.
	// These should not be included in the averages.
	var round4ID, matchID, clubMatchID int
	var clubSeasonID int
	require.NoError(t, pool.QueryRow(ctx,
		"SELECT id FROM afl.club_season LIMIT 1").Scan(&clubSeasonID))
	require.NoError(t, pool.QueryRow(ctx,
		"SELECT season_id FROM afl.round WHERE id = $1", ids.round1ID).
		Scan(new(int))) // just validate round exists; re-query season below

	var seasonID int
	require.NoError(t, pool.QueryRow(ctx,
		"SELECT season_id FROM afl.round WHERE id = $1", ids.round1ID).Scan(&seasonID))

	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.round (name, season_id) VALUES ('Round 4', $1) RETURNING id",
		seasonID).Scan(&round4ID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.match (round_id, venue, start_dt, data_status) VALUES ($1, 'Test Ground', '2025-06-22 14:00:00', 'partial') RETURNING id",
		round4ID).Scan(&matchID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.club_match (match_id, club_season_id, drv_score, rushed_behinds, side) VALUES ($1, $2, 100, 0, 'home') RETURNING id",
		matchID, clubSeasonID).Scan(&clubMatchID))
	_, err := pool.Exec(ctx,
		`INSERT INTO afl.player_match (club_match_id, player_season_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
		 VALUES ($1, $2, 100, 100, 100, 100, 100, 100, 100)`,
		clubMatchID, ids.playerSeasonID)
	require.NoError(t, err)

	rows, err := repo.GetSeasonStats(ctx, domain.PlayerSeasonStatsParams{
		PlayerSeasonIDs: []int{ids.playerSeasonID},
	})
	require.NoError(t, err)
	require.Len(t, rows, 1)

	t.Run("non-final match is excluded from averages", func(t *testing.T) {
		// Still 3 games (the partial match is excluded), goals avg still 1.0
		assert.Equal(t, 3, rows[0].Games)
		assert.InDelta(t, 1.0, rows[0].Goals, 0.001)
	})
}
