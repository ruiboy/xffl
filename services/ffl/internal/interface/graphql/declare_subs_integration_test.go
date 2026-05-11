//go:build integration

package graphql_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedPM inserts a player_match with explicit control over optional fields.
// Pass nil for position (bench players), drvAFLStatus, backupPositions, interchangePosition.
func seedPM(t *testing.T, pool *pgxpool.Pool, clubMatchID, playerSeasonID int, position, drvAFLStatus, backupPositions, interchangePosition *string, score int) int {
	t.Helper()
	ctx := context.Background()
	var id int
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO ffl.player_match
			(club_match_id, player_season_id, position, status, drv_afl_status, backup_positions, interchange_position, drv_score)
		VALUES ($1, $2, $3, 'named', $4, $5, $6, $7)
		RETURNING id`,
		clubMatchID, playerSeasonID, position, drvAFLStatus, backupPositions, interchangePosition, score,
	).Scan(&id))
	return id
}

func buildDeclareSubs(clubMatchID string, subbedOutIDs []string, interchangeApplied bool) string {
	ids := ""
	for i, id := range subbedOutIDs {
		if i > 0 {
			ids += ", "
		}
		ids += `"` + id + `"`
	}
	return fmt.Sprintf(`mutation {
		declareFFLSubstitutions(input: {
			clubMatchId: "%s"
			subbedOutPlayerMatchIds: [%s]
			interchangeApplied: %v
		}) {
			id
			status
			aflStatus
			score
		}
	}`, clubMatchID, ids, interchangeApplied)
}

// pmsByID indexes the declareFFLSubstitutions response by player-match ID.
func pmsByID(t *testing.T, raw json.RawMessage) map[string]struct {
	Status    *string
	AflStatus *string
	Score     int
} {
	t.Helper()
	var data struct {
		DeclareFFLSubstitutions []struct {
			ID        string  `json:"id"`
			Status    *string `json:"status"`
			AflStatus *string `json:"aflStatus"`
			Score     int     `json:"score"`
		} `json:"declareFFLSubstitutions"`
	}
	require.NoError(t, json.Unmarshal(raw, &data))
	out := make(map[string]struct {
		Status    *string
		AflStatus *string
		Score     int
	}, len(data.DeclareFFLSubstitutions))
	for _, pm := range data.DeclareFFLSubstitutions {
		out[pm.ID] = struct {
			Status    *string
			AflStatus *string
			Score     int
		}{pm.Status, pm.AflStatus, pm.Score}
	}
	return out
}

// ── Test 1: sub out a DNP starter ────────────────────────────────────────────

func TestDeclareFFLSubstitutions_SubbedOutDNPStarter(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()
	ctx := context.Background()

	// Make seeded player_match a DNP starter with score 0.
	_, err := pool.Exec(ctx,
		"UPDATE ffl.player_match SET drv_afl_status = 'dnp', drv_score = 0 WHERE id = $1",
		ids.playerMatchID)
	require.NoError(t, err)

	// Bench player covering the goals position, score 25.
	extras := seedExtraPlayers(t, pool, ids, 1)
	benchPS, _ := strconv.Atoi(extras[0])
	dnp := "dnp"
	bp := "goals"
	benchPMID := seedPM(t, pool, ids.homeClubMatchID, benchPS, nil, &dnp, &bp, nil, 25)

	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)
	starterID := fmt.Sprintf("%d", ids.playerMatchID)
	benchID := fmt.Sprintf("%d", benchPMID)

	result := execQuery(t, server, buildDeclareSubs(cmID, []string{starterID}, false))
	require.Empty(t, result.Errors)

	pms := pmsByID(t, result.Data)

	t.Run("starter is marked subbed", func(t *testing.T) {
		pm, ok := pms[starterID]
		require.True(t, ok, "starter not in response")
		require.NotNil(t, pm.Status)
		assert.Equal(t, "subbed", *pm.Status)
	})

	t.Run("bench player stays named", func(t *testing.T) {
		pm, ok := pms[benchID]
		require.True(t, ok, "bench player not in response")
		require.NotNil(t, pm.Status)
		assert.Equal(t, "named", *pm.Status)
	})

	t.Run("club match score uses bench player after recalculation", func(t *testing.T) {
		var score int
		require.NoError(t, pool.QueryRow(ctx,
			"SELECT drv_score FROM ffl.club_match WHERE id = $1", ids.homeClubMatchID).Scan(&score))
		// TM mode: subbed starter replaced by bench (score=25); starter score=0 dropped.
		assert.Equal(t, 25, score)
	})
}

// ── Test 2: interchange applied ───────────────────────────────────────────────

func TestDeclareFFLSubstitutions_InterchangeApplied(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()
	ctx := context.Background()

	// Starter1: goals, played, score 20 (reuse seeded player_match).
	_, err := pool.Exec(ctx,
		"UPDATE ffl.player_match SET drv_afl_status = 'played', drv_score = 20 WHERE id = $1",
		ids.playerMatchID)
	require.NoError(t, err)

	extras := seedExtraPlayers(t, pool, ids, 2)

	// Starter2: goals, played, score 5 (lower — will be marked interchanged).
	s2PS, _ := strconv.Atoi(extras[0])
	pos := "goals"
	played := "played"
	s2PMID := seedPM(t, pool, ids.homeClubMatchID, s2PS, &pos, &played, nil, nil, 5)

	// Interchange bench: backup_positions=goals, interchange_position=goals, score 30.
	icPS, _ := strconv.Atoi(extras[1])
	bp := "goals"
	ip := "goals"
	icPMID := seedPM(t, pool, ids.homeClubMatchID, icPS, nil, &played, &bp, &ip, 30)

	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)
	s1ID := fmt.Sprintf("%d", ids.playerMatchID)
	s2ID := fmt.Sprintf("%d", s2PMID)
	icID := fmt.Sprintf("%d", icPMID)

	result := execQuery(t, server, buildDeclareSubs(cmID, []string{}, true))
	require.Empty(t, result.Errors)

	pms := pmsByID(t, result.Data)

	t.Run("lowest-scoring starter at interchange position is marked interchanged", func(t *testing.T) {
		pm, ok := pms[s2ID]
		require.True(t, ok, "starter2 not in response")
		require.NotNil(t, pm.Status)
		assert.Equal(t, "interchanged", *pm.Status)
	})

	t.Run("higher-scoring starter stays named", func(t *testing.T) {
		pm, ok := pms[s1ID]
		require.True(t, ok, "starter1 not in response")
		require.NotNil(t, pm.Status)
		assert.Equal(t, "named", *pm.Status)
	})

	t.Run("interchange bench player stays named", func(t *testing.T) {
		pm, ok := pms[icID]
		require.True(t, ok, "interchange bench not in response")
		require.NotNil(t, pm.Status)
		assert.Equal(t, "named", *pm.Status)
	})

	t.Run("club match score swaps interchanged starter for bench player", func(t *testing.T) {
		var score int
		require.NoError(t, pool.QueryRow(ctx,
			"SELECT drv_score FROM ffl.club_match WHERE id = $1", ids.homeClubMatchID).Scan(&score))
		// Starter1(20) + IC Bench(30) = 50; Starter2 is swapped out.
		assert.Equal(t, 50, score)
	})
}

// ── Test 3: reject non-DNP starter ───────────────────────────────────────────

func TestDeclareFFLSubstitutions_RejectsNonDNPStarter(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()
	ctx := context.Background()

	// Starter with played status (not DNP).
	_, err := pool.Exec(ctx,
		"UPDATE ffl.player_match SET drv_afl_status = 'played', drv_score = 20 WHERE id = $1",
		ids.playerMatchID)
	require.NoError(t, err)

	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)
	pmID := fmt.Sprintf("%d", ids.playerMatchID)

	result := execQuery(t, server, buildDeclareSubs(cmID, []string{pmID}, false))

	t.Run("returns an error when trying to sub out a non-DNP starter", func(t *testing.T) {
		assert.NotEmpty(t, result.Errors)
	})
}

// ── Test 4: reject bench player passed as subbedOut ──────────────────────────

func TestDeclareFFLSubstitutions_RejectsBenchPlayer(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	extras := seedExtraPlayers(t, pool, ids, 1)
	benchPS, _ := strconv.Atoi(extras[0])
	dnp := "dnp"
	bp := "goals"
	benchPMID := seedPM(t, pool, ids.homeClubMatchID, benchPS, nil, &dnp, &bp, nil, 0)

	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)
	benchID := fmt.Sprintf("%d", benchPMID)

	result := execQuery(t, server, buildDeclareSubs(cmID, []string{benchID}, false))

	t.Run("returns an error when a bench player is passed as subbedOut", func(t *testing.T) {
		assert.NotEmpty(t, result.Errors)
	})
}

// ── Test 5: idempotent reset ─────────────────────────────────────────────────

func TestDeclareFFLSubstitutions_IdempotentReset(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()
	ctx := context.Background()

	// DNP starter.
	_, err := pool.Exec(ctx,
		"UPDATE ffl.player_match SET drv_afl_status = 'dnp', drv_score = 0 WHERE id = $1",
		ids.playerMatchID)
	require.NoError(t, err)

	extras := seedExtraPlayers(t, pool, ids, 1)
	benchPS, _ := strconv.Atoi(extras[0])
	dnp := "dnp"
	bp := "goals"
	seedPM(t, pool, ids.homeClubMatchID, benchPS, nil, &dnp, &bp, nil, 15)

	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)
	starterID := fmt.Sprintf("%d", ids.playerMatchID)

	// First call: sub out the DNP starter.
	first := execQuery(t, server, buildDeclareSubs(cmID, []string{starterID}, false))
	require.Empty(t, first.Errors)

	// Second call: empty subbedOut list — should reset starter back to named.
	second := execQuery(t, server, buildDeclareSubs(cmID, []string{}, false))
	require.Empty(t, second.Errors)

	pms := pmsByID(t, second.Data)

	t.Run("starter is reset to named when called again with empty list", func(t *testing.T) {
		pm, ok := pms[starterID]
		require.True(t, ok, "starter not in response")
		require.NotNil(t, pm.Status)
		assert.Equal(t, "named", *pm.Status)
	})
}

// ── Test 6: reject call on a final club match ─────────────────────────────────

func TestDeclareFFLSubstitutions_RejectsFinalClubMatch(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()
	ctx := context.Background()

	// Mark club match as final.
	_, err := pool.Exec(ctx,
		"UPDATE ffl.club_match SET data_status = 'final' WHERE id = $1", ids.homeClubMatchID)
	require.NoError(t, err)

	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)
	result := execQuery(t, server, buildDeclareSubs(cmID, []string{}, false))

	t.Run("returns an error when club match is already final", func(t *testing.T) {
		assert.NotEmpty(t, result.Errors)
	})
}
