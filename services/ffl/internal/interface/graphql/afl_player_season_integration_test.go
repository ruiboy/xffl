//go:build integration

package graphql_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFflPlayerSeasonsByAflPlayerSeason covers the new query and the two new
// fields it exposes: FFLPlayerSeason.club and FFLPlayerSeason.playerMatches.
func TestFflPlayerSeasonsByAflPlayerSeason(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()
	ctx := context.Background()

	// Create a real AFL player season so the FK is valid.
	aflSeasonID := insertAFLSeason(t, pool)
	aflPlayerSeasonID := insertAFLPlayerSeason(t, pool, aflSeasonID)

	// Point the seeded FFL player season at the real AFL player season.
	_, err := pool.Exec(ctx,
		"UPDATE ffl.player_season SET afl_player_season_id = $1 WHERE id = $2",
		aflPlayerSeasonID, ids.playerSeasonID)
	require.NoError(t, err)

	apsIDStr := fmt.Sprintf("%d", aflPlayerSeasonID)

	result := execQuery(t, server, `{
		fflPlayerSeasonsByAflPlayerSeason(aflPlayerSeasonId: "`+apsIDStr+`") {
			id
			clubSeasonId
			club { id name }
			playerMatches {
				id
				position
				status
				aflStatus
				score
			}
		}
	}`)
	require.Empty(t, result.Errors)

	var data struct {
		FflPlayerSeasonsByAflPlayerSeason []struct {
			ID           string `json:"id"`
			ClubSeasonID string `json:"clubSeasonId"`
			Club         struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"club"`
			PlayerMatches []struct {
				ID        string  `json:"id"`
				Position  *string `json:"position"`
				Status    *string `json:"status"`
				AflStatus *string `json:"aflStatus"`
				Score     int     `json:"score"`
			} `json:"playerMatches"`
		} `json:"fflPlayerSeasonsByAflPlayerSeason"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))

	t.Run("returns the player season linked to the AFL player season ID", func(t *testing.T) {
		require.Len(t, data.FflPlayerSeasonsByAflPlayerSeason, 1)
		assert.Equal(t, fmt.Sprintf("%d", ids.playerSeasonID), data.FflPlayerSeasonsByAflPlayerSeason[0].ID)
		assert.Equal(t, fmt.Sprintf("%d", ids.homeClubSeaID), data.FflPlayerSeasonsByAflPlayerSeason[0].ClubSeasonID)
	})

	t.Run("club is resolved from the club season", func(t *testing.T) {
		require.Len(t, data.FflPlayerSeasonsByAflPlayerSeason, 1)
		club := data.FflPlayerSeasonsByAflPlayerSeason[0].Club
		assert.NotEmpty(t, club.ID)
		assert.Equal(t, "Test Eagles", club.Name)
	})

	t.Run("player matches are returned with correct position and score", func(t *testing.T) {
		require.Len(t, data.FflPlayerSeasonsByAflPlayerSeason, 1)
		pms := data.FflPlayerSeasonsByAflPlayerSeason[0].PlayerMatches
		require.Len(t, pms, 1)
		require.NotNil(t, pms[0].Position)
		assert.Equal(t, "goals", *pms[0].Position)
		require.NotNil(t, pms[0].AflStatus)
		assert.Equal(t, "played", *pms[0].AflStatus)
		assert.Equal(t, 15, pms[0].Score)
	})
}

func TestFflPlayerSeasonsByAflPlayerSeason_UnknownID(t *testing.T) {
	pool := connectDB(t)
	seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	result := execQuery(t, server, `{
		fflPlayerSeasonsByAflPlayerSeason(aflPlayerSeasonId: "999999") { id }
	}`)
	require.Empty(t, result.Errors)

	var data struct {
		FflPlayerSeasonsByAflPlayerSeason []struct{ ID string } `json:"fflPlayerSeasonsByAflPlayerSeason"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))

	t.Run("returns empty list for an AFL player season with no FFL stints", func(t *testing.T) {
		assert.Empty(t, data.FflPlayerSeasonsByAflPlayerSeason)
	})
}

func TestFflPlayerSeasonsByAflPlayerSeason_MultipleFflStints(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()
	ctx := context.Background()

	// Create a real AFL player season.
	aflSeasonID := insertAFLSeason(t, pool)
	aflPlayerSeasonID := insertAFLPlayerSeason(t, pool, aflSeasonID)

	// Link the existing FFL player season (home club) to this AFL player season.
	_, err := pool.Exec(ctx,
		"UPDATE ffl.player_season SET afl_player_season_id = $1 WHERE id = $2",
		aflPlayerSeasonID, ids.playerSeasonID)
	require.NoError(t, err)

	// Add a second FFL player season for the same AFL player season (away club — simulates a trade).
	var ps2ID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.player_season (player_id, club_season_id, afl_player_season_id) VALUES ($1, $2, $3) RETURNING id",
		ids.playerID, ids.awayClubSeaID, aflPlayerSeasonID).Scan(&ps2ID))
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM ffl.player_season WHERE id = $1", ps2ID)
	})

	apsIDStr := fmt.Sprintf("%d", aflPlayerSeasonID)
	result := execQuery(t, server, `{
		fflPlayerSeasonsByAflPlayerSeason(aflPlayerSeasonId: "`+apsIDStr+`") {
			id
			club { name }
		}
	}`)
	require.Empty(t, result.Errors)

	var data struct {
		FflPlayerSeasonsByAflPlayerSeason []struct {
			ID   string `json:"id"`
			Club struct {
				Name string `json:"name"`
			} `json:"club"`
		} `json:"fflPlayerSeasonsByAflPlayerSeason"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))

	t.Run("returns one stint per FFL club that held the player", func(t *testing.T) {
		assert.Len(t, data.FflPlayerSeasonsByAflPlayerSeason, 2)
	})

	t.Run("each stint resolves its own FFL club", func(t *testing.T) {
		clubs := make(map[string]bool)
		for _, ps := range data.FflPlayerSeasonsByAflPlayerSeason {
			clubs[ps.Club.Name] = true
		}
		assert.True(t, clubs["Test Eagles"], "expected Test Eagles")
		assert.True(t, clubs["Test Lions"], "expected Test Lions")
	})
}
