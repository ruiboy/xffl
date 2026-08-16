//go:build integration

package twirp_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	aflv1 "xffl/contracts/gen/afl/v1"
	pg "xffl/services/afl/internal/infrastructure/postgres"
	"xffl/services/afl/internal/infrastructure/postgres/sqlcgen"
	rpcsrv "xffl/services/afl/internal/interface/twirp"
)

func newServer(pool *pgxpool.Pool) aflv1.PlayerLookup {
	q := sqlcgen.New(pool)
	return rpcsrv.NewPlayerLookupServer(
		pg.NewPlayerRepository(q),
		pg.NewPlayerSeasonRepository(q),
		pg.NewPlayerMatchRepository(q),
		pg.NewByeRepository(q),
	)
}

func cleanupTestData(ctx context.Context, t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	tables := []string{
		"afl.bye",
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
	}
	for _, tbl := range tables {
		_, err := pool.Exec(ctx, fmt.Sprintf("TRUNCATE %s CASCADE", tbl))
		require.NoError(t, err, "truncate %s", tbl)
	}
}

// seedBase creates the minimal league/season/clubs/club_seasons shared across LookupByeInfo tests.
type baseIDs struct {
	seasonID     int
	clubSeasonID int // "Bye Club" — gets a bye in round 2
	otherCSID    int // "No Bye Club" — no bye in round 2
}

func seedBase(t *testing.T, ctx context.Context, pool *pgxpool.Pool) baseIDs {
	t.Helper()
	var leagueID, seasonID, clubID, otherClubID, clubSeasonID, otherCSID int

	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.league (name) VALUES ('Test League') RETURNING id").Scan(&leagueID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.season (name, league_id) VALUES ('T2025', $1) RETURNING id",
		leagueID).Scan(&seasonID))

	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.club (name) VALUES ('Bye Club') RETURNING id").Scan(&clubID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.club (name) VALUES ('No Bye Club') RETURNING id").Scan(&otherClubID))

	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO afl.club_season (club_id, season_id, drv_played, drv_won, drv_lost, drv_drawn, drv_for, drv_against, drv_premiership_points)
		 VALUES ($1, $2, 1, 1, 0, 0, 80, 70, 4) RETURNING id`,
		clubID, seasonID).Scan(&clubSeasonID))
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO afl.club_season (club_id, season_id, drv_played, drv_won, drv_lost, drv_drawn, drv_for, drv_against, drv_premiership_points)
		 VALUES ($1, $2, 1, 0, 1, 0, 70, 80, 0) RETURNING id`,
		otherClubID, seasonID).Scan(&otherCSID))

	return baseIDs{seasonID: seasonID, clubSeasonID: clubSeasonID, otherCSID: otherCSID}
}

// seedRounds creates round 1 (prev final) and round 2 (bye round). Round 2 gets a
// dateless fixture (start_dt after round 1's) so GetPlayerSeasonAveragesBatch has a
// cutoff to compute, even though the bye club itself has no match that round.
// Returns prevRoundID, byeRoundID.
func seedRounds(t *testing.T, ctx context.Context, pool *pgxpool.Pool, seasonID int) (int, int) {
	t.Helper()
	var prevRoundID, byeRoundID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.round (name, season_id) VALUES ('R1', $1) RETURNING id",
		seasonID).Scan(&prevRoundID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.round (name, season_id) VALUES ('R2', $1) RETURNING id",
		seasonID).Scan(&byeRoundID))
	_, err := pool.Exec(ctx,
		"INSERT INTO afl.match (round_id, venue, start_dt, data_status) VALUES ($1, 'MCG', '2025-05-08 14:00:00', 'no_data')",
		byeRoundID)
	require.NoError(t, err)
	return prevRoundID, byeRoundID
}

// seedFinalMatch creates a final afl.match + afl.club_match for a given round and club_season.
// Returns the club_match_id.
func seedFinalMatch(t *testing.T, ctx context.Context, pool *pgxpool.Pool, roundID, clubSeasonID int) int {
	t.Helper()
	var matchID, clubMatchID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.match (round_id, venue, start_dt, data_status) VALUES ($1, 'MCG', '2025-05-01 14:00:00', 'final') RETURNING id",
		roundID).Scan(&matchID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.club_match (match_id, club_season_id, drv_score, rushed_behinds, side) VALUES ($1, $2, 80, 0, 'home') RETURNING id",
		matchID, clubSeasonID).Scan(&clubMatchID))
	return clubMatchID
}

// seedPlayer creates an afl.player + afl.player_season and returns the player_season_id.
func seedPlayer(t *testing.T, ctx context.Context, pool *pgxpool.Pool, name string, clubSeasonID int) int {
	t.Helper()
	var playerID, psID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player (name) VALUES ($1) RETURNING id", name).Scan(&playerID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player_season (player_id, club_season_id) VALUES ($1, $2) RETURNING id",
		playerID, clubSeasonID).Scan(&psID))
	return psID
}

// ── Test 1: bye player who played in the previous round ───────────────────────

func TestLookupByeInfo_ByePlayerWhoPlayedLast(t *testing.T) {
	pool := testPool
	ctx := context.Background()
	cleanupTestData(ctx, t, pool)
	t.Cleanup(func() { cleanupTestData(context.Background(), t, pool) })

	base := seedBase(t, ctx, pool)
	prevRoundID, byeRoundID := seedRounds(t, ctx, pool, base.seasonID)

	// Previous round: final match with 10 kicks, 5 handballs, 3 marks, 0 hitouts, 2 tackles, 2 goals, 1 behind
	clubMatchID := seedFinalMatch(t, ctx, pool, prevRoundID, base.clubSeasonID)
	psID := seedPlayer(t, ctx, pool, "Eligible Bye Player", base.clubSeasonID)
	_, err := pool.Exec(ctx,
		"INSERT INTO afl.player_match (club_match_id, player_season_id, kicks, handballs, marks, hitouts, tackles, goals, behinds) VALUES ($1, $2, 10, 5, 3, 0, 2, 2, 1)",
		clubMatchID, psID)
	require.NoError(t, err)

	// Bye in round 2 for the club
	_, err = pool.Exec(ctx, "INSERT INTO afl.bye (round_id, club_season_id) VALUES ($1, $2)", byeRoundID, base.clubSeasonID)
	require.NoError(t, err)

	srv := newServer(pool)
	resp, err := srv.LookupByeInfo(ctx, &aflv1.LookupByeInfoRequest{
		PlayerSeasonIds: []int32{int32(psID)},
		RoundId:         int32(byeRoundID),
	})
	require.NoError(t, err)
	require.Len(t, resp.Players, 1)
	info := resp.Players[0]

	t.Run("has_bye is true", func(t *testing.T) {
		assert.True(t, info.HasBye)
	})
	t.Run("played_last is true — player had a match in the previous final round", func(t *testing.T) {
		assert.True(t, info.PlayedLast)
	})
	t.Run("season averages are populated from the single previous match", func(t *testing.T) {
		assert.Equal(t, 10.0, info.AvgKicks)
		assert.Equal(t, 2.0, info.AvgGoals)
		assert.Equal(t, 5.0, info.AvgHandballs)
		assert.Equal(t, 3.0, info.AvgMarks)
		assert.Equal(t, 2.0, info.AvgTackles)
		assert.Equal(t, 0.0, info.AvgHitouts)
	})
}

// ── Test 2: bye player who did NOT play in the previous round ─────────────────

func TestLookupByeInfo_ByePlayerWhoDidNotPlayLast(t *testing.T) {
	pool := testPool
	ctx := context.Background()
	cleanupTestData(ctx, t, pool)
	t.Cleanup(func() { cleanupTestData(context.Background(), t, pool) })

	base := seedBase(t, ctx, pool)
	prevRoundID, byeRoundID := seedRounds(t, ctx, pool, base.seasonID)

	// Previous round: final match exists for the club but the player has no player_match row.
	seedFinalMatch(t, ctx, pool, prevRoundID, base.clubSeasonID)
	psID := seedPlayer(t, ctx, pool, "Absent Bye Player", base.clubSeasonID)

	_, err := pool.Exec(ctx, "INSERT INTO afl.bye (round_id, club_season_id) VALUES ($1, $2)", byeRoundID, base.clubSeasonID)
	require.NoError(t, err)

	srv := newServer(pool)
	resp, err := srv.LookupByeInfo(ctx, &aflv1.LookupByeInfoRequest{
		PlayerSeasonIds: []int32{int32(psID)},
		RoundId:         int32(byeRoundID),
	})
	require.NoError(t, err)
	require.Len(t, resp.Players, 1)
	info := resp.Players[0]

	t.Run("has_bye is true", func(t *testing.T) {
		assert.True(t, info.HasBye)
	})
	t.Run("played_last is false — no player_match in previous final round", func(t *testing.T) {
		assert.False(t, info.PlayedLast)
	})
	t.Run("no averages returned for ineligible player", func(t *testing.T) {
		// GetSeasonAveragesBatch is not called for non-played-last bye players —
		// the handler only fetches averages for players where has_bye=true.
		// Since the player has no matches at all, averages come back as zero values.
		assert.Equal(t, 0.0, info.AvgKicks)
		assert.Equal(t, 0.0, info.AvgGoals)
	})
}

// ── Test 3: non-bye player ────────────────────────────────────────────────────

func TestLookupByeInfo_NonByePlayer(t *testing.T) {
	pool := testPool
	ctx := context.Background()
	cleanupTestData(ctx, t, pool)
	t.Cleanup(func() { cleanupTestData(context.Background(), t, pool) })

	base := seedBase(t, ctx, pool)
	prevRoundID, byeRoundID := seedRounds(t, ctx, pool, base.seasonID)

	// Player in "No Bye Club" — that club has no bye in round 2.
	clubMatchID := seedFinalMatch(t, ctx, pool, prevRoundID, base.otherCSID)
	psID := seedPlayer(t, ctx, pool, "Non Bye Player", base.otherCSID)
	_, err := pool.Exec(ctx,
		"INSERT INTO afl.player_match (club_match_id, player_season_id, kicks, handballs, marks, hitouts, tackles, goals, behinds) VALUES ($1, $2, 8, 4, 2, 0, 1, 1, 0)",
		clubMatchID, psID)
	require.NoError(t, err)

	// Only "Bye Club" gets a bye; "No Bye Club" does not.
	_, err = pool.Exec(ctx, "INSERT INTO afl.bye (round_id, club_season_id) VALUES ($1, $2)", byeRoundID, base.clubSeasonID)
	require.NoError(t, err)

	srv := newServer(pool)
	resp, err := srv.LookupByeInfo(ctx, &aflv1.LookupByeInfoRequest{
		PlayerSeasonIds: []int32{int32(psID)},
		RoundId:         int32(byeRoundID),
	})
	require.NoError(t, err)
	require.Len(t, resp.Players, 1)
	info := resp.Players[0]

	t.Run("has_bye is false for player whose club is not on bye", func(t *testing.T) {
		assert.False(t, info.HasBye)
	})
	t.Run("no averages returned for non-bye player", func(t *testing.T) {
		assert.Equal(t, 0.0, info.AvgKicks)
	})
}

// ── Test 3b: averages only consider matches before the bye round's start_dt ───

func TestLookupByeInfo_AveragesExcludeMatchesAfterByeRound(t *testing.T) {
	pool := testPool
	ctx := context.Background()
	cleanupTestData(ctx, t, pool)
	t.Cleanup(func() { cleanupTestData(context.Background(), t, pool) })

	base := seedBase(t, ctx, pool)
	prevRoundID, byeRoundID := seedRounds(t, ctx, pool, base.seasonID)

	var futureRoundID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.round (name, season_id) VALUES ('R3', $1) RETURNING id",
		base.seasonID).Scan(&futureRoundID))

	// Previous round (R1, start_dt 2025-05-01): final match, player kicked 10.
	clubMatchID := seedFinalMatch(t, ctx, pool, prevRoundID, base.clubSeasonID)
	psID := seedPlayer(t, ctx, pool, "Bye Player", base.clubSeasonID)
	_, err := pool.Exec(ctx,
		"INSERT INTO afl.player_match (club_match_id, player_season_id, kicks, handballs, marks, hitouts, tackles, goals, behinds) VALUES ($1, $2, 10, 0, 0, 0, 0, 0, 0)",
		clubMatchID, psID)
	require.NoError(t, err)

	// Future round (R3, start_dt 2025-06-01): final match, player kicked 20.
	var futureMatchID, futureClubMatchID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.match (round_id, venue, start_dt, data_status) VALUES ($1, 'MCG', '2025-06-01 14:00:00', 'final') RETURNING id",
		futureRoundID).Scan(&futureMatchID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.club_match (match_id, club_season_id, drv_score, rushed_behinds, side) VALUES ($1, $2, 80, 0, 'home') RETURNING id",
		futureMatchID, base.clubSeasonID).Scan(&futureClubMatchID))
	_, err = pool.Exec(ctx,
		"INSERT INTO afl.player_match (club_match_id, player_season_id, kicks, handballs, marks, hitouts, tackles, goals, behinds) VALUES ($1, $2, 20, 0, 0, 0, 0, 0, 0)",
		futureClubMatchID, psID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "INSERT INTO afl.bye (round_id, club_season_id) VALUES ($1, $2)", byeRoundID, base.clubSeasonID)
	require.NoError(t, err)

	srv := newServer(pool)
	resp, err := srv.LookupByeInfo(ctx, &aflv1.LookupByeInfoRequest{
		PlayerSeasonIds: []int32{int32(psID)},
		RoundId:         int32(byeRoundID),
	})
	require.NoError(t, err)
	require.Len(t, resp.Players, 1)
	info := resp.Players[0]

	require.True(t, info.PlayedLast)
	t.Run("averages only include the match before the bye round, not the later one", func(t *testing.T) {
		assert.Equal(t, 10.0, info.AvgKicks)
	})
}

// ── LookupPlayerSeasonsBySeasonID ─────────────────────────────────────────────

func TestLookupPlayerSeasonsBySeasonID_ReturnsPlayerSeasonsWithNameAndClub(t *testing.T) {
	pool := testPool
	ctx := context.Background()
	cleanupTestData(ctx, t, pool)
	t.Cleanup(func() { cleanupTestData(context.Background(), t, pool) })

	base := seedBase(t, ctx, pool)
	darcyPS := seedPlayer(t, ctx, pool, "Darcy Fogarty", base.clubSeasonID)  // Bye Club
	reillyPS := seedPlayer(t, ctx, pool, "Reilly O'Brien", base.clubSeasonID) // Bye Club
	otherPS := seedPlayer(t, ctx, pool, "Lachie Neale", base.otherCSID)      // No Bye Club

	// A player in a different season must not leak into the result.
	var otherSeasonID, otherLeagueID, farClubID, farCSID int
	require.NoError(t, pool.QueryRow(ctx, "INSERT INTO afl.league (name) VALUES ('Other League') RETURNING id").Scan(&otherLeagueID))
	require.NoError(t, pool.QueryRow(ctx, "INSERT INTO afl.season (name, league_id) VALUES ('T2099', $1) RETURNING id", otherLeagueID).Scan(&otherSeasonID))
	require.NoError(t, pool.QueryRow(ctx, "INSERT INTO afl.club (name) VALUES ('Far Club') RETURNING id").Scan(&farClubID))
	require.NoError(t, pool.QueryRow(ctx, "INSERT INTO afl.club_season (club_id, season_id) VALUES ($1, $2) RETURNING id", farClubID, otherSeasonID).Scan(&farCSID))
	seedPlayer(t, ctx, pool, "Wrong Season Player", farCSID)

	srv := newServer(pool)
	resp, err := srv.LookupPlayerSeasonsBySeasonID(ctx, &aflv1.LookupPlayerSeasonsBySeasonIDRequest{AflSeasonId: int32(base.seasonID)})
	require.NoError(t, err)
	require.Len(t, resp.Players, 3, "only the three players in the season")

	byPS := make(map[int32]*aflv1.PlayerSeasonWithClub, 3)
	for _, p := range resp.Players {
		byPS[p.PlayerSeasonId] = p
	}

	t.Run("carries player_season_id, name and club", func(t *testing.T) {
		darcy := byPS[int32(darcyPS)]
		require.NotNil(t, darcy)
		assert.Equal(t, "Darcy Fogarty", darcy.Name)
		assert.Equal(t, "Bye Club", darcy.ClubName)
		assert.NotZero(t, darcy.PlayerId)
	})
	t.Run("includes players across clubs in the season", func(t *testing.T) {
		assert.Equal(t, "No Bye Club", byPS[int32(otherPS)].ClubName)
		assert.Equal(t, "Bye Club", byPS[int32(reillyPS)].ClubName)
	})
}

// ── Test 4: mixed batch — one bye-eligible player and one non-bye player ──────

func TestLookupByeInfo_MixedBatch(t *testing.T) {
	pool := testPool
	ctx := context.Background()
	cleanupTestData(ctx, t, pool)
	t.Cleanup(func() { cleanupTestData(context.Background(), t, pool) })

	base := seedBase(t, ctx, pool)
	prevRoundID, byeRoundID := seedRounds(t, ctx, pool, base.seasonID)

	// Bye Club player: played last round with 12 kicks.
	byeClubMatchID := seedFinalMatch(t, ctx, pool, prevRoundID, base.clubSeasonID)
	byePSID := seedPlayer(t, ctx, pool, "Bye Eligible", base.clubSeasonID)
	_, err := pool.Exec(ctx,
		"INSERT INTO afl.player_match (club_match_id, player_season_id, kicks, handballs, marks, hitouts, tackles, goals, behinds) VALUES ($1, $2, 12, 4, 2, 0, 1, 1, 0)",
		byeClubMatchID, byePSID)
	require.NoError(t, err)

	// No Bye Club player: played last round.
	otherClubMatchID := seedFinalMatch(t, ctx, pool, prevRoundID, base.otherCSID)
	noBYePSID := seedPlayer(t, ctx, pool, "No Bye Player", base.otherCSID)
	_, err = pool.Exec(ctx,
		"INSERT INTO afl.player_match (club_match_id, player_season_id, kicks, handballs, marks, hitouts, tackles, goals, behinds) VALUES ($1, $2, 7, 3, 1, 0, 2, 0, 1)",
		otherClubMatchID, noBYePSID)
	require.NoError(t, err)

	// Only Bye Club gets a bye.
	_, err = pool.Exec(ctx, "INSERT INTO afl.bye (round_id, club_season_id) VALUES ($1, $2)", byeRoundID, base.clubSeasonID)
	require.NoError(t, err)

	srv := newServer(pool)
	resp, err := srv.LookupByeInfo(ctx, &aflv1.LookupByeInfoRequest{
		PlayerSeasonIds: []int32{int32(byePSID), int32(noBYePSID)},
		RoundId:         int32(byeRoundID),
	})
	require.NoError(t, err)
	require.Len(t, resp.Players, 2)

	// Result order is not guaranteed — index by player_season_id.
	byInfo := make(map[int32]*aflv1.ByePlayerInfo, 2)
	for _, p := range resp.Players {
		byInfo[p.PlayerSeasonId] = p
	}

	t.Run("bye club player is flagged and has averages", func(t *testing.T) {
		info, ok := byInfo[int32(byePSID)]
		require.True(t, ok)
		assert.True(t, info.HasBye)
		assert.True(t, info.PlayedLast)
		assert.Equal(t, 12.0, info.AvgKicks)
	})
	t.Run("non-bye club player is not flagged", func(t *testing.T) {
		info, ok := byInfo[int32(noBYePSID)]
		require.True(t, ok)
		assert.False(t, info.HasBye)
		assert.Equal(t, 0.0, info.AvgKicks)
	})
}
