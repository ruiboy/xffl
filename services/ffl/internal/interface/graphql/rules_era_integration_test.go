//go:build integration

package graphql_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/ffl/internal/application"
)

// TestRecalculateFFLScore_UsesSeasonEraRules proves the per-season rules wiring
// end-to-end: a club match whose season is tagged with a non-default era scores
// with THAT era's formula, not the current one. Under the 1998 era a goal is
// worth 4 points (vs 5 today), so an identical stat line scores differently
// purely because of the season's rules_id.
func TestRecalculateFFLScore_UsesSeasonEraRules(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	ctx := context.Background()

	// Tag the season with the 1998 era (seedTestData leaves it at the default '2011').
	_, err := pool.Exec(ctx, "UPDATE ffl.season SET rules_id = '1998' WHERE id = $1", ids.seasonID)
	require.NoError(t, err)

	// A single starter at the 'goals' position.
	extras := seedExtraPlayers(t, pool, ids, 1)
	pmID := seedPM(t, pool, ids.homeClubMatchID, mustAtoi(extras[0]),
		sp("goals"), nil, nil, nil, 0)

	// Link a stub AFL player_match so RecalculateScore takes the linked path.
	const stubAFLMatchID = 9101
	_, err = pool.Exec(ctx,
		"UPDATE ffl.player_match SET afl_player_match_id = $1 WHERE id = $2",
		stubAFLMatchID, pmID)
	require.NoError(t, err)

	// goals=3: the current era would score 3*5=15; the 1998 era scores 3*4=12.
	server := setupServerWithMatchStats(t, pool, []application.PlayerMatchStats{
		{ID: stubAFLMatchID, Status: "played", Goals: 3},
	})
	defer server.Close()

	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)
	recalc := execQuery(t, server, fmt.Sprintf(`mutation {
		recalculateFFLClubMatchScore(clubMatchId: "%s")
	}`, cmID))
	require.Empty(t, recalc.Errors)

	queryResult := execQuery(t, server, fmt.Sprintf(`{
		fflClubMatch(id: "%s") { playerMatches { id score } }
	}`, cmID))
	require.Empty(t, queryResult.Errors)

	var data struct {
		FflClubMatch struct {
			PlayerMatches []struct {
				ID    string `json:"id"`
				Score int    `json:"score"`
			} `json:"playerMatches"`
		} `json:"fflClubMatch"`
	}
	require.NoError(t, json.Unmarshal(queryResult.Data, &data))

	var score *int
	target := fmt.Sprintf("%d", pmID)
	for _, pm := range data.FflClubMatch.PlayerMatches {
		if pm.ID == target {
			s := pm.Score
			score = &s
			break
		}
	}
	require.NotNil(t, score, "seeded player match not found in query result")
	assert.Equal(t, 12, *score, "goals=3 scored under 1998 rules (goal worth 4), not 15 under the current era")
}
