//go:build integration

package graphql_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedExtraSeasonForPlayer creates a whole season (season → round → match →
// club_season → player_season) for an existing player, with the match played at
// matchStart. Returns the new player_season id.
//
// Seasons are created in call order, so ids ascend with each call regardless of
// the dates given — which is what lets the ordering test below distinguish
// chronological ordering from id ordering.
func seedExtraSeasonForPlayer(
	t *testing.T, pool *pgxpool.Pool, ids testIDs, seasonName, matchStart string,
) int {
	t.Helper()
	ctx := context.Background()

	var seasonID, roundID, clubSeasonID, matchID, playerSeasonID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.season (name, league_id) VALUES ($1, $2) RETURNING id",
		seasonName, ids.leagueID).Scan(&seasonID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.round (name, season_id) VALUES ('Round 1', $1) RETURNING id",
		seasonID).Scan(&roundID))
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO afl.club_season (club_id, season_id, drv_played, drv_won, drv_lost, drv_drawn, drv_for, drv_against, drv_premiership_points)
		 VALUES ($1, $2, 0, 0, 0, 0, 0, 0, 0) RETURNING id`,
		ids.homeClubID, seasonID).Scan(&clubSeasonID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.match (round_id, venue, start_dt, data_status) VALUES ($1, 'Test Ground', $2, 'final') RETURNING id",
		roundID, matchStart).Scan(&matchID))
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player_season (player_id, club_season_id) VALUES ($1, $2) RETURNING id",
		ids.playerID, clubSeasonID).Scan(&playerSeasonID))

	return playerSeasonID
}

// queryPlayerSeasonIDs asks for the player's season history via the page's own
// entry point — the player_season being viewed — and returns the ids in the
// order the server produced them.
func queryPlayerSeasonIDs(t *testing.T, server *httptest.Server, viewingPlayerSeasonID int) []string {
	t.Helper()
	result := execQuery(t, server, fmt.Sprintf(`{
		aflPlayerSeason(id: "%d") {
			player { playerSeasons { id } }
		}
	}`, viewingPlayerSeasonID))
	require.Empty(t, result.Errors)

	var data struct {
		AflPlayerSeason struct {
			Player struct {
				PlayerSeasons []struct{ ID string } `json:"playerSeasons"`
			} `json:"player"`
		} `json:"aflPlayerSeason"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))

	out := make([]string, len(data.AflPlayerSeason.Player.PlayerSeasons))
	for i, ps := range data.AflPlayerSeason.Player.PlayerSeasons {
		out[i] = ps.ID
	}
	return out
}

// TestPlayerSeasons_OrdersByMatchDateNotSeasonID is the test this feature
// exists for. afl.season.id is not chronological, so ordering by id looks
// correct on tidy data and silently wrong on real data.
//
// The seasons below are created oldest-content-last, so their ids ascend while
// their match dates do not:
//
//	base season  (lowest id)  matches in 2025
//	'Test 2019'  (middle id)  matches in 2019
//	'Test 2030'  (highest id) matches in 2030
//
// Correct (chronological): 2030, 2025, 2019.
// Ordering by season id would give: 2030, 2019, 2025 — so the middle position
// is what distinguishes the two.
func TestPlayerSeasons_OrdersByMatchDateNotSeasonID(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	older := seedExtraSeasonForPlayer(t, pool, ids, "Test 2019", "2019-06-15 14:00:00")
	newer := seedExtraSeasonForPlayer(t, pool, ids, "Test 2030", "2030-06-15 14:00:00")

	require.Greater(t, older, ids.playerSeasonID,
		"precondition: the 2019 season must have a higher id than the 2025 one, or this test proves nothing")

	got := queryPlayerSeasonIDs(t, server, ids.playerSeasonID)

	assert.Equal(t, []string{
		fmt.Sprintf("%d", newer),              // 2030
		fmt.Sprintf("%d", ids.playerSeasonID), // 2025
		fmt.Sprintf("%d", older),              // 2019
	}, got, "player seasons must be ordered by match date, most recent first")
}

// TestPlayerSeasons_IncludesTheSeasonBeingViewed keeps the server honest about
// what it returns: the full history, including the current one. The view drops
// the current season itself, and that filtering belongs on the client.
func TestPlayerSeasons_IncludesTheSeasonBeingViewed(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	got := queryPlayerSeasonIDs(t, server, ids.playerSeasonID)

	assert.Equal(t, []string{fmt.Sprintf("%d", ids.playerSeasonID)}, got)
}

// TestPlayerSeasons_ExcludesSoftDeleted guards the deleted_at filters in the
// query; without them a removed season would still appear as a chip.
func TestPlayerSeasons_ExcludesSoftDeleted(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	deleted := seedExtraSeasonForPlayer(t, pool, ids, "Test 2020", "2020-06-15 14:00:00")
	_, err := pool.Exec(context.Background(),
		"UPDATE afl.player_season SET deleted_at = now() WHERE id = $1", deleted)
	require.NoError(t, err)

	got := queryPlayerSeasonIDs(t, server, ids.playerSeasonID)

	assert.NotContains(t, got, fmt.Sprintf("%d", deleted))
	assert.Equal(t, []string{fmt.Sprintf("%d", ids.playerSeasonID)}, got)
}
