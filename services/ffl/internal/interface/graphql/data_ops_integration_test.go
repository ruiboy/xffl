//go:build integration

package graphql_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	gqlhandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/ffl/internal/application"
	"xffl/services/ffl/internal/infrastructure/forum"
	pg "xffl/services/ffl/internal/infrastructure/postgres"
	"xffl/services/ffl/internal/infrastructure/postgres/sqlcgen"
	gql "xffl/services/ffl/internal/interface/graphql"
	memevents "xffl/shared/events/memory"
)

// stubPlayerLookup returns a fixed candidate list regardless of which AFL IDs are requested.
// LookupPlayerSeason reads directly from afl.player_season in the test DB so the
// addFFLPlayerToSeason flow works against real seeded data without standing up
// the AFL service over Twirp.
type stubPlayerLookup struct {
	pool       *pgxpool.Pool
	candidates []application.PlayerCandidate
}

func (s *stubPlayerLookup) LookupPlayers(_ context.Context, _ []int) ([]application.PlayerCandidate, error) {
	return s.candidates, nil
}

func (s *stubPlayerLookup) LookupPlayerSeason(ctx context.Context, aflPlayerSeasonID int) (int, error) {
	if s.pool == nil {
		return 0, fmt.Errorf("stubPlayerLookup: pool not set; cannot resolve afl.player_season %d", aflPlayerSeasonID)
	}
	var aflPlayerID int
	err := s.pool.QueryRow(ctx,
		"SELECT player_id FROM afl.player_season WHERE id = $1", aflPlayerSeasonID).Scan(&aflPlayerID)
	if err != nil {
		return 0, fmt.Errorf("lookup afl.player_season %d: %w", aflPlayerSeasonID, err)
	}
	return aflPlayerID, nil
}

func setupDataOpsServer(t *testing.T, pool *pgxpool.Pool, dataOps *application.DataOpsCommands) *httptest.Server {
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
	commands := application.NewCommands(db, memevents.New(), application.CommandsDeps{
		EventRepos: application.EventRepos{
			Rounds:        pg.NewRoundRepository(q),
			PlayerSeasons: pg.NewPlayerSeasonRepository(q),
			PlayerMatches: pg.NewPlayerMatchRepository(q),
		},
		PlayerLookup: &stubPlayerLookup{pool: pool},
	})

	resolver := &gql.Resolver{Queries: queries, Commands: commands, DataOps: dataOps}
	srv := gqlhandler.NewDefaultServer(gql.NewExecutableSchema(gql.Config{Resolvers: resolver}))

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := gql.InjectLoaders(r.Context(), gql.NewLoaders(queries))
		srv.ServeHTTP(w, r.WithContext(ctx))
	})
	return httptest.NewServer(h)
}

func TestParseAndConfirmFFLTeamSubmission(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	ctx := context.Background()

	// Seed Jeremy Cameron as an AFL player + FFL player + player_season.
	var jeremyAFLID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO afl.player (name) VALUES ('Jeremy Cameron') RETURNING id").Scan(&jeremyAFLID))

	var jeremyFflID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.player (drv_name, afl_player_id) VALUES ('Jeremy Cameron', $1) RETURNING id",
		jeremyAFLID).Scan(&jeremyFflID))

	var jeremyPSID int
	require.NoError(t, pool.QueryRow(ctx,
		"INSERT INTO ffl.player_season (player_id, club_season_id) VALUES ($1, $2) RETURNING id",
		jeremyFflID, ids.homeClubSeaID).Scan(&jeremyPSID))

	// Stub player lookup: returns Jeremy Cameron only.
	stub := &stubPlayerLookup{
		candidates: []application.PlayerCandidate{
			{AFLPlayerID: jeremyAFLID, Name: "Jeremy Cameron", Club: "Geel"},
		},
	}
	dataOps := application.NewDataOpsCommands(
		pg.NewDB(pool),
		stub,
		forum.NewLevenshteinResolver(),
		forum.NewParser(),
	)

	server := setupDataOpsServer(t, pool, dataOps)
	defer server.Close()

	post, err := os.ReadFile("../../infrastructure/forum/testdata/ruiboys.txt")
	require.NoError(t, err)

	clubSeaID := toIDStr(ids.homeClubSeaID)
	clubMatchID := toIDStr(ids.homeClubMatchID)

	// ── Step 1: Parse ────────────────────────────────────────────────────────

	parseResult := execQuery(t, server, `mutation {
		parseFFLTeamSubmission(input: {
			clubSeasonId: "`+clubSeaID+`"
			clubMatchId: "`+clubMatchID+`"
			teamName: "Ruiboys"
			post: `+jsonString(string(post))+`
		}) {
			resolvedPlayers {
				parsedName
				clubHint
				resolvedName
				position
				backupPositions
				interchangePosition
				score
				playerSeasonId
				confidence
			}
			needsReview
		}
	}`)

	require.Empty(t, parseResult.Errors)

	var parseData struct {
		ParseFFLTeamSubmission struct {
			ResolvedPlayers []struct {
				ParsedName          string  `json:"parsedName"`
				ClubHint            string  `json:"clubHint"`
				ResolvedName        *string `json:"resolvedName"`
				Position            string  `json:"position"`
				BackupPositions     string  `json:"backupPositions"`
				InterchangePosition string  `json:"interchangePosition"`
				Score               *int    `json:"score"`
				PlayerSeasonID      *string `json:"playerSeasonId"`
				Confidence          float64 `json:"confidence"`
			} `json:"resolvedPlayers"`
			NeedsReview []int `json:"needsReview"`
		} `json:"parseFFLTeamSubmission"`
	}
	require.NoError(t, json.Unmarshal(parseResult.Data, &parseData))

	rps := parseData.ParseFFLTeamSubmission.ResolvedPlayers

	t.Run("parses all 22 player rows", func(t *testing.T) {
		assert.Len(t, rps, 22)
	})

	// Find Jeremy Cameron in results.
	var jeremy *struct {
		ParsedName          string  `json:"parsedName"`
		ClubHint            string  `json:"clubHint"`
		ResolvedName        *string `json:"resolvedName"`
		Position            string  `json:"position"`
		BackupPositions     string  `json:"backupPositions"`
		InterchangePosition string  `json:"interchangePosition"`
		Score               *int    `json:"score"`
		PlayerSeasonID      *string `json:"playerSeasonId"`
		Confidence          float64 `json:"confidence"`
	}
	for i := range rps {
		if rps[i].ParsedName == "Jeremy Cameron" {
			jeremy = &rps[i]
			break
		}
	}

	t.Run("Jeremy Cameron is resolved with high confidence", func(t *testing.T) {
		require.NotNil(t, jeremy, "Jeremy Cameron not found in resolved players")
		assert.Equal(t, "goals", jeremy.Position)
		assert.Equal(t, "Geel", jeremy.ClubHint)
		assert.NotNil(t, jeremy.PlayerSeasonID)
		assert.InDelta(t, 1.0, jeremy.Confidence, 0.01)
		require.NotNil(t, jeremy.Score)
		assert.Equal(t, 15, *jeremy.Score)
	})

	// ── Step 2: Confirm (Jeremy Cameron only) ────────────────────────────────

	require.NotNil(t, jeremy.PlayerSeasonID, "need resolved player to confirm")

	confirmResult := execQuery(t, server, `mutation {
		confirmFFLTeamSubmission(input: {
			clubMatchId: "`+clubMatchID+`"
			players: [{
				playerSeasonId: "`+*jeremy.PlayerSeasonID+`"
				position: "goals"
				backupPositions: null
				interchangePosition: null
				score: 15
			}]
		}) {
			id
			position
			score
		}
	}`)

	require.Empty(t, confirmResult.Errors)

	var confirmData struct {
		ConfirmFFLTeamSubmission []struct {
			ID       string `json:"id"`
			Position string `json:"position"`
			Score    int    `json:"score"`
		} `json:"confirmFFLTeamSubmission"`
	}
	require.NoError(t, json.Unmarshal(confirmResult.Data, &confirmData))

	t.Run("confirm creates one player_match record", func(t *testing.T) {
		require.Len(t, confirmData.ConfirmFFLTeamSubmission, 1)
		pm := confirmData.ConfirmFFLTeamSubmission[0]
		assert.Equal(t, "goals", pm.Position)
		assert.Equal(t, 15, pm.Score)
	})

	t.Run("player_match is persisted in DB", func(t *testing.T) {
		var count int
		require.NoError(t, pool.QueryRow(ctx,
			"SELECT COUNT(*) FROM ffl.player_match WHERE club_match_id = $1 AND player_season_id = $2",
			ids.homeClubMatchID, jeremyPSID).Scan(&count))
		assert.Equal(t, 1, count)
	})
}

func toIDStr(id int) string {
	return fmt.Sprintf("%d", id)
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
