//go:build integration

package graphql_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	gqlhandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/ffl/internal/application"
	pg "xffl/services/ffl/internal/infrastructure/postgres"
	"xffl/services/ffl/internal/infrastructure/postgres/sqlcgen"
	gql "xffl/services/ffl/internal/interface/graphql"
	memevents "xffl/shared/events/memory"
)

// matchStatLookupStub implements PlayerLookup and returns a fixed set of
// PlayerMatchStats from LookupPlayerMatch. All other methods are no-ops.
// This lets integration tests exercise the RecalculateScore "linked" path
// without standing up the real AFL service.
type matchStatLookupStub struct {
	stats []application.PlayerMatchStats
}

func (s *matchStatLookupStub) LookupPlayers(_ context.Context, _ []int) ([]application.PlayerCandidate, error) {
	return nil, nil
}
func (s *matchStatLookupStub) LookupPlayerSeason(_ context.Context, _ int) (int, error) {
	return 0, nil
}
func (s *matchStatLookupStub) LookupPlayerMatch(_ context.Context, _ []int) ([]application.PlayerMatchStats, error) {
	return s.stats, nil
}
func (s *matchStatLookupStub) LookupPlayerMatchBySeasonRound(_ context.Context, _ []int, _ int) ([]application.PlayerMatchStats, error) {
	return nil, nil
}
func (s *matchStatLookupStub) LookupByeInfo(_ context.Context, _ []int, _ int) ([]application.ByePlayerInfo, error) {
	return nil, nil
}

func (s *matchStatLookupStub) LookupPlayerSeasonsBySeasonID(_ context.Context, _ int) ([]application.PlayerCandidate, error) {
	return nil, nil
}

// setupServerWithMatchStats creates a test HTTP server whose RecalculateScore
// command will return the given AFL stats from LookupPlayerMatch.
func setupServerWithMatchStats(t *testing.T, pool *pgxpool.Pool, stats []application.PlayerMatchStats) *httptest.Server {
	t.Helper()
	q := sqlcgen.New(pool)
	queries := application.NewQueries(
		pg.NewClubRepository(q),
		pg.NewSeasonRepository(q),
		pg.NewRoundRepository(q),
		pg.NewMatchRepository(q),
		pg.NewClubSeasonRepository(q),
		pg.NewClubMatchRepository(q),
		pg.NewPlayerRepository(q),
		pg.NewPlayerMatchRepository(q),
		pg.NewPlayerSeasonRepository(q),
	)
	db := pg.NewDB(pool)
	commands := application.NewCommands(
		db,
		memevents.New(),
		&matchStatLookupStub{stats: stats},
		pg.NewMatchRepository(q),
		pg.NewClubMatchRepository(q),
		pg.NewClubSeasonRepository(q),
		pg.NewRoundRepository(q),
		pg.NewPlayerMatchRepository(q),
		pg.NewPlayerSeasonRepository(q),
	)
	resolver := &gql.Resolver{Queries: queries, Commands: commands}
	srv := gqlhandler.NewDefaultServer(gql.NewExecutableSchema(gql.Config{Resolvers: resolver}))
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := gql.InjectLoaders(r.Context(), gql.NewLoaders(queries))
		srv.ServeHTTP(w, r.WithContext(ctx))
	})
	return httptest.NewServer(h)
}

// suggestionsQuery queries suggestedSubstitutions for a club match.
func suggestionsQuery(clubMatchID string) string {
	return fmt.Sprintf(`{
		fflClubMatch(id: "%s") {
			suggestedSubstitutions { kind replacedPmId replacingPmId }
		}
	}`, clubMatchID)
}

// ── SuggestedSubstitutions: interchange ──────────────────────────────────────

// TestSuggestedSubstitutions_InterchangeWhenBenchOutscores verifies that a
// single interchange suggestion is returned when the bench player's drv_score
// strictly exceeds the lowest starter at the interchange position.
func TestSuggestedSubstitutions_InterchangeWhenBenchOutscores(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	ctx := context.Background()

	// Lower-scoring goals starter (seeded player at score=40).
	_, err := pool.Exec(ctx,
		"UPDATE ffl.player_match SET drv_afl_status = 'played', drv_score = 40 WHERE id = $1",
		ids.playerMatchID)
	require.NoError(t, err)

	extras := seedExtraPlayers(t, pool, ids, 2)

	// Higher-scoring goals starter (score=50).
	seedPM(t, pool, ids.homeClubMatchID, mustAtoi(extras[0]),
		sp("goals"), sp("played"), nil, nil, 50)

	// Interchange bench player covering goals (score=60, outscores both starters).
	benchPMID := seedPM(t, pool, ids.homeClubMatchID, mustAtoi(extras[1]),
		nil, sp("played"), sp("goals"), sp("goals"), 60)

	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)
	result := execQuery(t, server, suggestionsQuery(cmID))
	require.Empty(t, result.Errors)

	var data struct {
		FflClubMatch struct {
			SuggestedSubstitutions []struct {
				Kind          string `json:"kind"`
				ReplacedPmID  string `json:"replacedPmId"`
				ReplacingPmID string `json:"replacingPmId"`
			} `json:"suggestedSubstitutions"`
		} `json:"fflClubMatch"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))

	subs := data.FflClubMatch.SuggestedSubstitutions
	t.Run("exactly one interchange suggestion is returned", func(t *testing.T) {
		require.Len(t, subs, 1)
		assert.Equal(t, "interchange", subs[0].Kind)
	})
	t.Run("bench player is the replacing PM", func(t *testing.T) {
		require.Len(t, subs, 1)
		assert.Equal(t, fmt.Sprintf("%d", benchPMID), subs[0].ReplacingPmID)
	})
	t.Run("lower-scoring starter is the replaced PM", func(t *testing.T) {
		require.Len(t, subs, 1)
		// Lower starter (score=40) is the one displaced, not the higher (score=50).
		assert.Equal(t, fmt.Sprintf("%d", ids.playerMatchID), subs[0].ReplacedPmID)
	})
}

// TestSuggestedSubstitutions_NoInterchangeWhenBenchDoesNotOutscore verifies
// that no suggestion is returned when the bench score is less than the starter.
func TestSuggestedSubstitutions_NoInterchangeWhenBenchDoesNotOutscore(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	ctx := context.Background()

	// Starter with high score (80).
	_, err := pool.Exec(ctx,
		"UPDATE ffl.player_match SET drv_afl_status = 'played', drv_score = 80 WHERE id = $1",
		ids.playerMatchID)
	require.NoError(t, err)

	extras := seedExtraPlayers(t, pool, ids, 1)

	// Interchange bench player with lower score (40 < 80 → no suggestion).
	seedPM(t, pool, ids.homeClubMatchID, mustAtoi(extras[0]),
		nil, sp("played"), sp("goals"), sp("goals"), 40)

	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)
	result := execQuery(t, server, suggestionsQuery(cmID))
	require.Empty(t, result.Errors)

	var data struct {
		FflClubMatch struct {
			SuggestedSubstitutions []struct {
				Kind string `json:"kind"`
			} `json:"suggestedSubstitutions"`
		} `json:"fflClubMatch"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))

	t.Run("no suggestions when bench does not outscore starter", func(t *testing.T) {
		assert.Empty(t, data.FflClubMatch.SuggestedSubstitutions)
	})
}

// ── SuggestedSubstitutions: sub ───────────────────────────────────────────────

// TestSuggestedSubstitutions_SubWhenStarterIsDNP verifies that a sub suggestion
// is returned when a starter is DNP and a bench player covers that position.
func TestSuggestedSubstitutions_SubWhenStarterIsDNP(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	ctx := context.Background()

	// DNP goals starter.
	_, err := pool.Exec(ctx,
		"UPDATE ffl.player_match SET drv_afl_status = 'dnp', drv_score = 0 WHERE id = $1",
		ids.playerMatchID)
	require.NoError(t, err)

	extras := seedExtraPlayers(t, pool, ids, 1)

	// Backup bench player covering goals (no interchange slot).
	benchPMID := seedPM(t, pool, ids.homeClubMatchID, mustAtoi(extras[0]),
		nil, sp("played"), sp("goals"), nil, 25)

	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)
	result := execQuery(t, server, suggestionsQuery(cmID))
	require.Empty(t, result.Errors)

	var data struct {
		FflClubMatch struct {
			SuggestedSubstitutions []struct {
				Kind          string `json:"kind"`
				ReplacedPmID  string `json:"replacedPmId"`
				ReplacingPmID string `json:"replacingPmId"`
			} `json:"suggestedSubstitutions"`
		} `json:"fflClubMatch"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))

	subs := data.FflClubMatch.SuggestedSubstitutions
	t.Run("one sub suggestion is returned", func(t *testing.T) {
		require.Len(t, subs, 1)
		assert.Equal(t, "sub", subs[0].Kind)
	})
	t.Run("DNP starter is the replaced PM", func(t *testing.T) {
		require.Len(t, subs, 1)
		assert.Equal(t, fmt.Sprintf("%d", ids.playerMatchID), subs[0].ReplacedPmID)
	})
	t.Run("bench player is the replacing PM", func(t *testing.T) {
		require.Len(t, subs, 1)
		assert.Equal(t, fmt.Sprintf("%d", benchPMID), subs[0].ReplacingPmID)
	})
}

// TestSuggestedSubstitutions_BothInterchangeAndSubReturnedTogether verifies that
// both an interchange and a sub suggestion are returned when conditions for both
// are present in the same club match.
func TestSuggestedSubstitutions_BothInterchangeAndSubReturnedTogether(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	ctx := context.Background()

	// Seeded player becomes a DNP goals starter.
	_, err := pool.Exec(ctx,
		"UPDATE ffl.player_match SET drv_afl_status = 'dnp', drv_score = 0 WHERE id = $1",
		ids.playerMatchID)
	require.NoError(t, err)

	extras := seedExtraPlayers(t, pool, ids, 3)

	// Star starter (score=30) — will be displaced by the interchange bench.
	seedPM(t, pool, ids.homeClubMatchID, mustAtoi(extras[0]),
		sp("star"), sp("played"), nil, nil, 30)

	// Interchange bench player at 'star' (score=50 > 30 → fires interchange).
	icPMID := seedPM(t, pool, ids.homeClubMatchID, mustAtoi(extras[1]),
		nil, sp("played"), sp("star"), sp("star"), 50)

	// Sub bench player covering goals (score=20, DNP goals starter → fires sub).
	subBenchPMID := seedPM(t, pool, ids.homeClubMatchID, mustAtoi(extras[2]),
		nil, sp("played"), sp("goals"), nil, 20)

	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)
	result := execQuery(t, server, suggestionsQuery(cmID))
	require.Empty(t, result.Errors)

	var data struct {
		FflClubMatch struct {
			SuggestedSubstitutions []struct {
				Kind          string `json:"kind"`
				ReplacedPmID  string `json:"replacedPmId"`
				ReplacingPmID string `json:"replacingPmId"`
			} `json:"suggestedSubstitutions"`
		} `json:"fflClubMatch"`
	}
	require.NoError(t, json.Unmarshal(result.Data, &data))

	subs := data.FflClubMatch.SuggestedSubstitutions
	t.Run("two suggestions are returned", func(t *testing.T) {
		assert.Len(t, subs, 2)
	})

	findKind := func(kind string) (replacedID, replacingID string, found bool) {
		for _, s := range subs {
			if s.Kind == kind {
				return s.ReplacedPmID, s.ReplacingPmID, true
			}
		}
		return "", "", false
	}

	t.Run("interchange suggestion present with correct replacing PM", func(t *testing.T) {
		_, replacingID, ok := findKind("interchange")
		require.True(t, ok, "expected an interchange suggestion")
		assert.Equal(t, fmt.Sprintf("%d", icPMID), replacingID)
	})
	t.Run("sub suggestion present with correct IDs", func(t *testing.T) {
		replacedID, replacingID, ok := findKind("sub")
		require.True(t, ok, "expected a sub suggestion")
		assert.Equal(t, fmt.Sprintf("%d", ids.playerMatchID), replacedID)
		assert.Equal(t, fmt.Sprintf("%d", subBenchPMID), replacingID)
	})
}

// ── RecalculateScore: interchange bench gets potential score ──────────────────

// TestRecalculateFFLScore_InterchangeBenchGetsPotentialScore verifies that
// RecalculateScore scores an interchange bench player at their interchange
// position (not nil), so SuggestedSubstitutions can compare meaningful values.
func TestRecalculateFFLScore_InterchangeBenchGetsPotentialScore(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	ctx := context.Background()

	extras := seedExtraPlayers(t, pool, ids, 1)

	// Interchange bench player: position=nil, backup_positions='goals', interchange_position='goals'.
	benchPMID := seedPM(t, pool, ids.homeClubMatchID, mustAtoi(extras[0]),
		nil, nil, sp("goals"), sp("goals"), 0)

	// Attach a stub AFL player_match_id so RecalculateScore takes the linked path.
	const stubAFLMatchID = 9001
	_, err := pool.Exec(ctx,
		"UPDATE ffl.player_match SET afl_player_match_id = $1 WHERE id = $2",
		stubAFLMatchID, benchPMID)
	require.NoError(t, err)

	// Build server with a stub that returns goals=3 for stubAFLMatchID.
	server := setupServerWithMatchStats(t, pool, []application.PlayerMatchStats{
		{ID: stubAFLMatchID, Status: "played", Goals: 3},
	})
	defer server.Close()

	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)
	recalcResult := execQuery(t, server, fmt.Sprintf(`mutation {
		recalculateFFLClubMatchScore(clubMatchId: "%s")
	}`, cmID))
	require.Empty(t, recalcResult.Errors)

	queryResult := execQuery(t, server, fmt.Sprintf(`{
		fflClubMatch(id: "%s") {
			playerMatches { id score }
		}
	}`, cmID))
	require.Empty(t, queryResult.Errors)

	var queryData struct {
		FflClubMatch struct {
			PlayerMatches []struct {
				ID    string `json:"id"`
				Score int    `json:"score"`
			} `json:"playerMatches"`
		} `json:"fflClubMatch"`
	}
	require.NoError(t, json.Unmarshal(queryResult.Data, &queryData))

	benchID := fmt.Sprintf("%d", benchPMID)
	var benchScore *int
	for _, pm := range queryData.FflClubMatch.PlayerMatches {
		if pm.ID == benchID {
			s := pm.Score
			benchScore = &s
			break
		}
	}

	t.Run("interchange bench drv_score is computed at interchange position", func(t *testing.T) {
		require.NotNil(t, benchScore, "bench player match not found in query result")
		// goals=3 at 'goals' position: 3 * GoalsMultiplier(5) = 15
		assert.Equal(t, 15, *benchScore)
	})
}

// ── RecalculateFFLLadder: premiership points ──────────────────────────────────

// TestRecalculateFFLLadder_PremiershipPointsPersisted verifies that after
// recalculating the ladder, each club_season row has the correct drv_premiership_points
// (4 for the winner, 0 for the loser) and the GraphQL premiershipPoints field reflects it.
func TestRecalculateFFLLadder_PremiershipPointsPersisted(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServerWithScoreCommands(t, pool)
	defer server.Close()

	// Mark both club_matches final so recalculateFFLLadder picks them up.
	// Eagles (home, 85) beat Lions (away, 72).
	_, err := pool.Exec(context.Background(),
		"UPDATE ffl.club_match SET data_status = 'final' WHERE match_id = $1", ids.matchID)
	require.NoError(t, err)

	mutation := fmt.Sprintf(`mutation {
		recalculateFFLLadder(seasonId: "%d")
	}`, ids.seasonID)
	result := execQuery(t, server, mutation)
	require.Empty(t, result.Errors)

	ladderResult := execQuery(t, server, fmt.Sprintf(`{
		fflSeason(id: "%d") {
			ladder { club { name } premiershipPoints }
		}
	}`, ids.seasonID))
	require.Empty(t, ladderResult.Errors)

	var ladderData struct {
		FflSeason struct {
			Ladder []struct {
				Club struct {
					Name string `json:"name"`
				} `json:"club"`
				PremiershipPoints int `json:"premiershipPoints"`
			} `json:"ladder"`
		} `json:"fflSeason"`
	}
	require.NoError(t, json.Unmarshal(ladderResult.Data, &ladderData))
	require.Len(t, ladderData.FflSeason.Ladder, 2)

	t.Run("winner has 4 premiership points", func(t *testing.T) {
		// Ladder ordered by pp DESC: Eagles (winner) first.
		assert.Equal(t, "Test Eagles", ladderData.FflSeason.Ladder[0].Club.Name)
		assert.Equal(t, 4, ladderData.FflSeason.Ladder[0].PremiershipPoints)
	})
	t.Run("loser has 0 premiership points", func(t *testing.T) {
		assert.Equal(t, "Test Lions", ladderData.FflSeason.Ladder[1].Club.Name)
		assert.Equal(t, 0, ladderData.FflSeason.Ladder[1].PremiershipPoints)
	})
	t.Run("stale pre-seeded premiership points are overwritten", func(t *testing.T) {
		// Seed had drv_premiership_points=16 for Eagles — recalculation replaces with 4.
		assert.NotEqual(t, 16, ladderData.FflSeason.Ladder[0].PremiershipPoints)
	})
}
