//go:build integration

package graphql_test

import (
	"context"
	"encoding/json"
	"fmt"
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

func sp(s string) *string { return &s }

func buildDeclareSubs(clubMatchID string, subs [][]string, interchange *[2]string) string {
	subsJSON := "["
	for i, pair := range subs {
		if i > 0 {
			subsJSON += ", "
		}
		subsJSON += fmt.Sprintf(`{replacedPmId: "%s", replacingPmId: "%s"}`, pair[0], pair[1])
	}
	subsJSON += "]"

	interchangeJSON := "null"
	if interchange != nil {
		interchangeJSON = fmt.Sprintf(`{replacedPmId: "%s", replacingPmId: "%s"}`, interchange[0], interchange[1])
	}

	return fmt.Sprintf(`mutation {
		declareFFLSubstitutions(input: {
			clubMatchId: "%s"
			subs: %s
			interchange: %s
		}) {
			id
			status
			aflStatus
			score
		}
	}`, clubMatchID, subsJSON, interchangeJSON)
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

// ── Test 1: explicit sub pairing ─────────────────────────────────────────────

func TestDeclareFFLSubstitutions_ExplicitSubPairing(t *testing.T) {
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

	// Bench player covering the goals position.
	extras := seedExtraPlayers(t, pool, ids, 1)
	benchPS := mustAtoi(extras[0])
	benchPMID := seedPM(t, pool, ids.homeClubMatchID, benchPS, nil, sp("dnp"), sp("goals"), nil, 25)

	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)
	starterID := fmt.Sprintf("%d", ids.playerMatchID)
	benchID := fmt.Sprintf("%d", benchPMID)

	result := execQuery(t, server, buildDeclareSubs(cmID,
		[][]string{{starterID, benchID}},
		nil,
	))
	require.Empty(t, result.Errors)

	pms := pmsByID(t, result.Data)

	t.Run("starter is marked subbed_out", func(t *testing.T) {
		pm, ok := pms[starterID]
		require.True(t, ok)
		require.NotNil(t, pm.Status)
		assert.Equal(t, "subbed_out", *pm.Status)
	})

	t.Run("bench player is marked subbed_in", func(t *testing.T) {
		pm, ok := pms[benchID]
		require.True(t, ok)
		require.NotNil(t, pm.Status)
		assert.Equal(t, "subbed_in", *pm.Status)
	})

	t.Run("club match score uses bench player after recalculation", func(t *testing.T) {
		var score int
		require.NoError(t, pool.QueryRow(ctx,
			"SELECT drv_score FROM ffl.club_match WHERE id = $1", ids.homeClubMatchID).Scan(&score))
		assert.Equal(t, 25, score)
	})
}

// ── Test 2: explicit interchange pairing ──────────────────────────────────────

func TestDeclareFFLSubstitutions_ExplicitInterchangePairing(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()
	ctx := context.Background()

	// Starter1 (reuse seeded): goals, played, score 20.
	_, err := pool.Exec(ctx,
		"UPDATE ffl.player_match SET drv_afl_status = 'played', drv_score = 20 WHERE id = $1",
		ids.playerMatchID)
	require.NoError(t, err)

	extras := seedExtraPlayers(t, pool, ids, 2)

	// Starter2: goals, played, score 5.
	s2PS := mustAtoi(extras[0])
	pos := "goals"
	played := "played"
	s2PMID := seedPM(t, pool, ids.homeClubMatchID, s2PS, &pos, &played, nil, nil, 5)

	// Interchange bench player.
	icPS := mustAtoi(extras[1])
	bp := "goals"
	icPMID := seedPM(t, pool, ids.homeClubMatchID, icPS, nil, &played, &bp, &bp, 30)

	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)
	s1ID := fmt.Sprintf("%d", ids.playerMatchID)
	s2ID := fmt.Sprintf("%d", s2PMID)
	icID := fmt.Sprintf("%d", icPMID)

	// TM explicitly picks starter2 to be displaced.
	result := execQuery(t, server, buildDeclareSubs(cmID,
		nil,
		&[2]string{s2ID, icID},
	))
	require.Empty(t, result.Errors)

	pms := pmsByID(t, result.Data)

	t.Run("chosen starter is marked interchanged_out", func(t *testing.T) {
		pm, ok := pms[s2ID]
		require.True(t, ok)
		require.NotNil(t, pm.Status)
		assert.Equal(t, "interchanged_out", *pm.Status)
	})

	t.Run("other starter stays named", func(t *testing.T) {
		pm, ok := pms[s1ID]
		require.True(t, ok)
		require.NotNil(t, pm.Status)
		assert.Equal(t, "named", *pm.Status)
	})

	t.Run("interchange bench player is marked interchanged_in", func(t *testing.T) {
		pm, ok := pms[icID]
		require.True(t, ok)
		require.NotNil(t, pm.Status)
		assert.Equal(t, "interchanged_in", *pm.Status)
	})

	t.Run("club match score uses interchange bench instead of displaced starter", func(t *testing.T) {
		var score int
		require.NoError(t, pool.QueryRow(ctx,
			"SELECT drv_score FROM ffl.club_match WHERE id = $1", ids.homeClubMatchID).Scan(&score))
		// Starter1(20) + IC Bench(30) = 50; Starter2 displaced.
		assert.Equal(t, 50, score)
	})
}

// ── Test 3: reject non-DNP starter for sub ───────────────────────────────────

func TestDeclareFFLSubstitutions_RejectsNonDNPStarterForSub(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()
	ctx := context.Background()

	_, err := pool.Exec(ctx,
		"UPDATE ffl.player_match SET drv_afl_status = 'played', drv_score = 20 WHERE id = $1",
		ids.playerMatchID)
	require.NoError(t, err)

	extras := seedExtraPlayers(t, pool, ids, 1)
	benchPS := mustAtoi(extras[0])
	benchPMID := seedPM(t, pool, ids.homeClubMatchID, benchPS, nil, sp("played"), sp("goals"), nil, 0)

	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)
	pmID := fmt.Sprintf("%d", ids.playerMatchID)
	benchID := fmt.Sprintf("%d", benchPMID)

	result := execQuery(t, server, buildDeclareSubs(cmID,
		[][]string{{pmID, benchID}},
		nil,
	))

	t.Run("returns an error when subbing out a non-DNP starter", func(t *testing.T) {
		assert.NotEmpty(t, result.Errors)
	})
}

// ── Test 4: reject bench player as replaced ───────────────────────────────────

func TestDeclareFFLSubstitutions_RejectsBenchPlayerAsReplaced(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	extras := seedExtraPlayers(t, pool, ids, 2)
	bench1PS := mustAtoi(extras[0])
	bench2PS := mustAtoi(extras[1])
	bench1ID := fmt.Sprintf("%d", seedPM(t, pool, ids.homeClubMatchID, bench1PS, nil, sp("dnp"), sp("goals"), nil, 0))
	bench2ID := fmt.Sprintf("%d", seedPM(t, pool, ids.homeClubMatchID, bench2PS, nil, sp("dnp"), sp("goals"), nil, 10))

	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)

	result := execQuery(t, server, buildDeclareSubs(cmID,
		[][]string{{bench1ID, bench2ID}}, // both are bench players
		nil,
	))

	t.Run("returns an error when replaced ID is a bench player", func(t *testing.T) {
		assert.NotEmpty(t, result.Errors)
	})
}

// ── Test 5: idempotent reset (re-declare with empty lists) ────────────────────

func TestDeclareFFLSubstitutions_IdempotentReset(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()
	ctx := context.Background()

	_, err := pool.Exec(ctx,
		"UPDATE ffl.player_match SET drv_afl_status = 'dnp', drv_score = 0 WHERE id = $1",
		ids.playerMatchID)
	require.NoError(t, err)

	extras := seedExtraPlayers(t, pool, ids, 1)
	benchPS := mustAtoi(extras[0])
	benchPMID := seedPM(t, pool, ids.homeClubMatchID, benchPS, nil, sp("dnp"), sp("goals"), nil, 15)

	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)
	starterID := fmt.Sprintf("%d", ids.playerMatchID)
	benchID := fmt.Sprintf("%d", benchPMID)

	// First declare: sub out the DNP starter.
	first := execQuery(t, server, buildDeclareSubs(cmID,
		[][]string{{starterID, benchID}},
		nil,
	))
	require.Empty(t, first.Errors)

	// Re-declare with empty lists — resets all to named.
	second := execQuery(t, server, buildDeclareSubs(cmID, nil, nil))
	require.Empty(t, second.Errors)

	pms := pmsByID(t, second.Data)

	t.Run("starter is reset to named", func(t *testing.T) {
		pm, ok := pms[starterID]
		require.True(t, ok)
		require.NotNil(t, pm.Status)
		assert.Equal(t, "named", *pm.Status)
	})

	t.Run("bench player is reset to named", func(t *testing.T) {
		pm, ok := pms[benchID]
		require.True(t, ok)
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

	_, err := pool.Exec(ctx,
		"UPDATE ffl.club_match SET data_status = 'final' WHERE id = $1", ids.homeClubMatchID)
	require.NoError(t, err)

	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)
	result := execQuery(t, server, buildDeclareSubs(cmID, nil, nil))

	t.Run("returns an error when club match is already final", func(t *testing.T) {
		assert.NotEmpty(t, result.Errors)
	})
}

// ── Test 7: DNP starter with no declaration scores zero ───────────────────────

func TestDeclareFFLSubstitutions_DNPNoDeclarationScoresZero(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()
	ctx := context.Background()

	// Starter DNP, bench available but TM declares nothing.
	_, err := pool.Exec(ctx,
		"UPDATE ffl.player_match SET drv_afl_status = 'dnp', drv_score = 0 WHERE id = $1",
		ids.playerMatchID)
	require.NoError(t, err)

	extras := seedExtraPlayers(t, pool, ids, 1)
	benchPS := mustAtoi(extras[0])
	seedPM(t, pool, ids.homeClubMatchID, benchPS, nil, sp("dnp"), sp("goals"), nil, 20)

	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)

	// Declare nothing — bench stays unused.
	result := execQuery(t, server, buildDeclareSubs(cmID, nil, nil))
	require.Empty(t, result.Errors)

	t.Run("club match score is zero when DNP starter undeclared", func(t *testing.T) {
		var score int
		require.NoError(t, pool.QueryRow(ctx,
			"SELECT drv_score FROM ffl.club_match WHERE id = $1", ids.homeClubMatchID).Scan(&score))
		assert.Equal(t, 0, score)
	})
}

func mustAtoi(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}
