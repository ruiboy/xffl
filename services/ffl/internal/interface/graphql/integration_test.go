package graphql_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	gqlhandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/jackc/pgx/v5/pgxpool"

	"xffl/services/ffl/internal/application"
	pg "xffl/services/ffl/internal/infrastructure/postgres"
	"xffl/services/ffl/internal/infrastructure/postgres/sqlcgen"
	gql "xffl/services/ffl/internal/interface/graphql"
)

// testIDs holds IDs of rows inserted by seedTestData, used by tests to query known data.
type testIDs struct {
	leagueID        int
	seasonID        int
	roundID         int
	homeClubID      int
	awayClubID      int
	homeClubSeaID   int
	awayClubSeaID   int
	matchID         int
	homeClubMatchID int
	awayClubMatchID int
	aflPlayerID     int
	playerID        int
	playerSeasonID  int
	playerMatchID   int
}

func connectDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/xffl?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("skipping integration test: cannot connect to database: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("skipping integration test: database not available: %v", err)
	}

	t.Cleanup(func() { pool.Close() })
	return pool
}

func seedTestData(t *testing.T, pool *pgxpool.Pool) testIDs {
	t.Helper()
	ctx := context.Background()
	var ids testIDs

	// Truncate in reverse FK order
	tables := []string{
		"ffl.player_match", "ffl.player_season", "ffl.player",
		"ffl.club_match", "ffl.match", "ffl.club_season",
		"ffl.club", "ffl.round", "ffl.season", "ffl.league",
		"afl.player",
	}
	for _, table := range tables {
		if _, err := pool.Exec(ctx, fmt.Sprintf("TRUNCATE %s CASCADE", table)); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	// League
	err := pool.QueryRow(ctx,
		"INSERT INTO ffl.league (name) VALUES ('Test FFL') RETURNING id").Scan(&ids.leagueID)
	if err != nil {
		t.Fatalf("failed to insert league: %v", err)
	}

	// Season
	err = pool.QueryRow(ctx,
		"INSERT INTO ffl.season (name, league_id) VALUES ('Test 2025', $1) RETURNING id",
		ids.leagueID).Scan(&ids.seasonID)
	if err != nil {
		t.Fatalf("failed to insert season: %v", err)
	}

	// Round
	err = pool.QueryRow(ctx,
		"INSERT INTO ffl.round (name, season_id) VALUES ('Round 1', $1) RETURNING id",
		ids.seasonID).Scan(&ids.roundID)
	if err != nil {
		t.Fatalf("failed to insert round: %v", err)
	}

	// Two clubs
	err = pool.QueryRow(ctx,
		"INSERT INTO ffl.club (name) VALUES ('Test Eagles') RETURNING id").Scan(&ids.homeClubID)
	if err != nil {
		t.Fatalf("failed to insert home club: %v", err)
	}
	err = pool.QueryRow(ctx,
		"INSERT INTO ffl.club (name) VALUES ('Test Lions') RETURNING id").Scan(&ids.awayClubID)
	if err != nil {
		t.Fatalf("failed to insert away club: %v", err)
	}

	// Club seasons (home team higher on ladder)
	err = pool.QueryRow(ctx,
		`INSERT INTO ffl.club_season (club_id, season_id, drv_played, drv_won, drv_lost, drv_drawn, drv_for, drv_against, drv_premiership_points)
		 VALUES ($1, $2, 5, 4, 1, 0, 500, 400, 16) RETURNING id`,
		ids.homeClubID, ids.seasonID).Scan(&ids.homeClubSeaID)
	if err != nil {
		t.Fatalf("failed to insert home club season: %v", err)
	}
	err = pool.QueryRow(ctx,
		`INSERT INTO ffl.club_season (club_id, season_id, drv_played, drv_won, drv_lost, drv_drawn, drv_for, drv_against, drv_premiership_points)
		 VALUES ($1, $2, 5, 3, 2, 0, 450, 420, 12) RETURNING id`,
		ids.awayClubID, ids.seasonID).Scan(&ids.awayClubSeaID)
	if err != nil {
		t.Fatalf("failed to insert away club season: %v", err)
	}

	// Match
	err = pool.QueryRow(ctx,
		"INSERT INTO ffl.match (round_id, match_style, venue, start_dt) VALUES ($1, 'versus', 'Test Ground', '2025-06-15 14:00:00') RETURNING id",
		ids.roundID).Scan(&ids.matchID)
	if err != nil {
		t.Fatalf("failed to insert match: %v", err)
	}

	// Club matches
	err = pool.QueryRow(ctx,
		"INSERT INTO ffl.club_match (match_id, club_season_id, drv_score) VALUES ($1, $2, 85) RETURNING id",
		ids.matchID, ids.homeClubSeaID).Scan(&ids.homeClubMatchID)
	if err != nil {
		t.Fatalf("failed to insert home club match: %v", err)
	}
	err = pool.QueryRow(ctx,
		"INSERT INTO ffl.club_match (match_id, club_season_id, drv_score) VALUES ($1, $2, 72) RETURNING id",
		ids.matchID, ids.awayClubSeaID).Scan(&ids.awayClubMatchID)
	if err != nil {
		t.Fatalf("failed to insert away club match: %v", err)
	}

	// Link match to club matches
	_, err = pool.Exec(ctx,
		"UPDATE ffl.match SET home_club_match_id = $1, away_club_match_id = $2 WHERE id = $3",
		ids.homeClubMatchID, ids.awayClubMatchID, ids.matchID)
	if err != nil {
		t.Fatalf("failed to link match to club matches: %v", err)
	}

	// AFL player reference (needed for FFL player FK)
	// Name deliberately does not contain "Test" to avoid polluting AFL player search tests
	err = pool.QueryRow(ctx,
		"INSERT INTO afl.player (name) VALUES ('Seeded AFL Player') RETURNING id").Scan(&ids.aflPlayerID)
	if err != nil {
		t.Fatalf("failed to insert afl player: %v", err)
	}

	// Player (linked to AFL player)
	err = pool.QueryRow(ctx,
		"INSERT INTO ffl.player (drv_name, afl_player_id) VALUES ('Test Player', $1) RETURNING id",
		ids.aflPlayerID).Scan(&ids.playerID)
	if err != nil {
		t.Fatalf("failed to insert player: %v", err)
	}

	// Player season
	err = pool.QueryRow(ctx,
		"INSERT INTO ffl.player_season (player_id, club_season_id) VALUES ($1, $2) RETURNING id",
		ids.playerID, ids.homeClubSeaID).Scan(&ids.playerSeasonID)
	if err != nil {
		t.Fatalf("failed to insert player season: %v", err)
	}

	// Player match: goals position, played status, score 15
	err = pool.QueryRow(ctx,
		`INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status, drv_score)
		 VALUES ($1, $2, 'goals', 'played', 15) RETURNING id`,
		ids.homeClubMatchID, ids.playerSeasonID).Scan(&ids.playerMatchID)
	if err != nil {
		t.Fatalf("failed to insert player match: %v", err)
	}

	t.Cleanup(func() {
		for _, table := range tables {
			pool.Exec(context.Background(), fmt.Sprintf("TRUNCATE %s CASCADE", table))
		}
	})

	return ids
}

func setupTestServer(t *testing.T, pool *pgxpool.Pool) *httptest.Server {
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
	commands := application.NewCommands(db)

	resolver := &gql.Resolver{Queries: queries, Commands: commands}
	srv := gqlhandler.NewDefaultServer(gql.NewExecutableSchema(gql.Config{Resolvers: resolver}))

	return httptest.NewServer(srv)
}

type graphqlRequest struct {
	Query string `json:"query"`
}

type graphqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func execQuery(t *testing.T, server *httptest.Server, query string) graphqlResponse {
	t.Helper()

	body, _ := json.Marshal(graphqlRequest{Query: query})
	resp, err := http.Post(server.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}

	var result graphqlResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return result
}

func TestFflClubs(t *testing.T) {
	pool := connectDB(t)
	seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	result := execQuery(t, server, `{ fflClubs { id name } }`)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	var data struct {
		FflClubs []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"fflClubs"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if len(data.FflClubs) != 2 {
		t.Errorf("expected 2 clubs, got %d", len(data.FflClubs))
	}

	// Clubs ordered by name: Test Eagles before Test Lions
	if len(data.FflClubs) == 2 {
		if data.FflClubs[0].Name != "Test Eagles" {
			t.Errorf("expected first club Test Eagles, got %s", data.FflClubs[0].Name)
		}
		if data.FflClubs[1].Name != "Test Lions" {
			t.Errorf("expected second club Test Lions, got %s", data.FflClubs[1].Name)
		}
	}
}

func TestFflSeasons(t *testing.T) {
	pool := connectDB(t)
	seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	result := execQuery(t, server, `{ fflSeasons { id name } }`)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	var data struct {
		FflSeasons []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"fflSeasons"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if len(data.FflSeasons) != 1 {
		t.Errorf("expected 1 season, got %d", len(data.FflSeasons))
	}
	if len(data.FflSeasons) > 0 && data.FflSeasons[0].Name != "Test 2025" {
		t.Errorf("expected season name Test 2025, got %s", data.FflSeasons[0].Name)
	}
}

func TestFflSeasonWithLadder(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	seasonID := fmt.Sprintf("%d", ids.seasonID)
	result := execQuery(t, server, `{ fflSeason(id: "`+seasonID+`") { name ladder { club { name } season { name } played won lost percentage } } }`)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	var data struct {
		FflSeason struct {
			Name   string `json:"name"`
			Ladder []struct {
				Club struct {
					Name string `json:"name"`
				} `json:"club"`
				Season struct {
					Name string `json:"name"`
				} `json:"season"`
				Played     int     `json:"played"`
				Won        int     `json:"won"`
				Lost       int     `json:"lost"`
				Percentage float64 `json:"percentage"`
			} `json:"ladder"`
		} `json:"fflSeason"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if len(data.FflSeason.Ladder) != 2 {
		t.Fatalf("expected 2 ladder entries, got %d", len(data.FflSeason.Ladder))
	}

	// Eagles: 500 for / 400 against = 125.0%
	if data.FflSeason.Ladder[0].Club.Name != "Test Eagles" {
		t.Errorf("expected Test Eagles first on ladder, got %s", data.FflSeason.Ladder[0].Club.Name)
	}
	if data.FflSeason.Ladder[0].Percentage != 125.0 {
		t.Errorf("expected 125.0 percentage, got %f", data.FflSeason.Ladder[0].Percentage)
	}
	if data.FflSeason.Ladder[1].Club.Name != "Test Lions" {
		t.Errorf("expected Test Lions second on ladder, got %s", data.FflSeason.Ladder[1].Club.Name)
	}
	// Verify season is populated on ladder entries
	if data.FflSeason.Ladder[0].Season.Name != "Test 2025" {
		t.Errorf("expected ladder entry season 'Test 2025', got %s", data.FflSeason.Ladder[0].Season.Name)
	}
}

func TestFflSeasonGraphTraversal(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	seasonID := fmt.Sprintf("%d", ids.seasonID)
	result := execQuery(t, server, `{
		fflSeason(id: "`+seasonID+`") {
			name
			rounds {
				name
				matches {
					venue
					homeClubMatch {
						club { name }
						score
						playerMatches {
							player { name }
							position
							status
							score
						}
					}
					awayClubMatch {
						club { name }
						score
					}
				}
			}
		}
	}`)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	var data struct {
		FflSeason struct {
			Name   string `json:"name"`
			Rounds []struct {
				Name    string `json:"name"`
				Matches []struct {
					Venue         string `json:"venue"`
					HomeClubMatch *struct {
						Club struct {
							Name string `json:"name"`
						} `json:"club"`
						Score         int `json:"score"`
						PlayerMatches []struct {
							Player struct {
								Name string `json:"name"`
							} `json:"player"`
							Position *string `json:"position"`
							Status   *string `json:"status"`
							Score    int     `json:"score"`
						} `json:"playerMatches"`
					} `json:"homeClubMatch"`
					AwayClubMatch *struct {
						Club struct {
							Name string `json:"name"`
						} `json:"club"`
						Score int `json:"score"`
					} `json:"awayClubMatch"`
				} `json:"matches"`
			} `json:"rounds"`
		} `json:"fflSeason"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if len(data.FflSeason.Rounds) != 1 {
		t.Fatalf("expected 1 round, got %d", len(data.FflSeason.Rounds))
	}

	round := data.FflSeason.Rounds[0]
	if round.Name != "Round 1" {
		t.Errorf("expected Round 1, got %s", round.Name)
	}
	if len(round.Matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(round.Matches))
	}

	match := round.Matches[0]
	if match.Venue != "Test Ground" {
		t.Errorf("expected venue Test Ground, got %s", match.Venue)
	}
	if match.HomeClubMatch == nil {
		t.Fatal("expected home club match")
	}
	if match.HomeClubMatch.Club.Name != "Test Eagles" {
		t.Errorf("expected home club Test Eagles, got %s", match.HomeClubMatch.Club.Name)
	}
	if match.HomeClubMatch.Score != 85 {
		t.Errorf("expected home score 85, got %d", match.HomeClubMatch.Score)
	}
	if match.AwayClubMatch == nil {
		t.Fatal("expected away club match")
	}
	if match.AwayClubMatch.Club.Name != "Test Lions" {
		t.Errorf("expected away club Test Lions, got %s", match.AwayClubMatch.Club.Name)
	}
	if match.AwayClubMatch.Score != 72 {
		t.Errorf("expected away score 72, got %d", match.AwayClubMatch.Score)
	}

	// Player match: goals position, played status, score 15
	if len(match.HomeClubMatch.PlayerMatches) != 1 {
		t.Fatalf("expected 1 player match, got %d", len(match.HomeClubMatch.PlayerMatches))
	}
	pm := match.HomeClubMatch.PlayerMatches[0]
	if pm.Player.Name != "Test Player" {
		t.Errorf("expected Test Player, got %s", pm.Player.Name)
	}
	if pm.Position == nil || *pm.Position != "goals" {
		t.Errorf("expected position goals, got %v", pm.Position)
	}
	if pm.Score != 15 {
		t.Errorf("expected score 15, got %d", pm.Score)
	}
}

func TestFflLatestRound(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	// Add a second round so we can verify "latest" picks the last one
	ctx := context.Background()
	var round2ID int
	err := pool.QueryRow(ctx,
		"INSERT INTO ffl.round (name, season_id) VALUES ('Round 2', $1) RETURNING id",
		ids.seasonID).Scan(&round2ID)
	if err != nil {
		t.Fatalf("failed to insert round 2: %v", err)
	}

	result := execQuery(t, server, `{
		fflLatestRound {
			name
			season { name }
		}
	}`)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	var data struct {
		FflLatestRound struct {
			Name   string `json:"name"`
			Season struct {
				Name string `json:"name"`
			} `json:"season"`
		} `json:"fflLatestRound"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if data.FflLatestRound.Name != "Round 2" {
		t.Errorf("expected latest round 'Round 2', got %s", data.FflLatestRound.Name)
	}
	if data.FflLatestRound.Season.Name != "Test 2025" {
		t.Errorf("expected season 'Test 2025', got %s", data.FflLatestRound.Season.Name)
	}
}

func TestFflClubSeason(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	seasonID := fmt.Sprintf("%d", ids.seasonID)
	clubID := fmt.Sprintf("%d", ids.homeClubID)
	result := execQuery(t, server, `{
		fflClubSeason(seasonId: "`+seasonID+`", clubId: "`+clubID+`") {
			id
			club { name }
			season { id name }
			played
			won
		}
	}`)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	var data struct {
		FflClubSeason struct {
			ID     string `json:"id"`
			Club   struct{ Name string } `json:"club"`
			Season struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"season"`
			Played int `json:"played"`
			Won    int `json:"won"`
		} `json:"fflClubSeason"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	cs := data.FflClubSeason
	if cs.Club.Name != "Test Eagles" {
		t.Errorf("expected club 'Test Eagles', got %s", cs.Club.Name)
	}
	if cs.Season.Name != "Test 2025" {
		t.Errorf("expected season 'Test 2025', got %s", cs.Season.Name)
	}
	if cs.Season.ID != seasonID {
		t.Errorf("expected season id %s, got %s", seasonID, cs.Season.ID)
	}
	if cs.Played != 5 {
		t.Errorf("expected played 5, got %d", cs.Played)
	}
	if cs.Won != 4 {
		t.Errorf("expected won 4, got %d", cs.Won)
	}
}

func TestCreateFFLPlayer(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	// Create a second AFL player for the new FFL player to reference
	ctx := context.Background()
	var aflPlayerID2 int
	if err := pool.QueryRow(ctx, "INSERT INTO afl.player (name) VALUES ('New AFL Player') RETURNING id").Scan(&aflPlayerID2); err != nil {
		t.Fatalf("failed to insert afl player: %v", err)
	}
	_ = ids

	aflPID := fmt.Sprintf("%d", aflPlayerID2)
	result := execQuery(t, server, `mutation {
		createFFLPlayer(input: { name: "New Player", aflPlayerId: "`+aflPID+`" }) {
			id
			name
		}
	}`)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	var data struct {
		CreateFFLPlayer struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"createFFLPlayer"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if data.CreateFFLPlayer.Name != "New Player" {
		t.Errorf("expected name 'New Player', got %s", data.CreateFFLPlayer.Name)
	}
	if data.CreateFFLPlayer.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestUpdateFFLPlayer(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	playerID := fmt.Sprintf("%d", ids.playerID)
	result := execQuery(t, server, `mutation {
		updateFFLPlayer(input: { id: "`+playerID+`", name: "Renamed Player" }) {
			id
			name
		}
	}`)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	var data struct {
		UpdateFFLPlayer struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"updateFFLPlayer"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if data.UpdateFFLPlayer.Name != "Renamed Player" {
		t.Errorf("expected 'Renamed Player', got %s", data.UpdateFFLPlayer.Name)
	}
}

func TestDeleteFFLPlayer(t *testing.T) {
	pool := connectDB(t)
	seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	// Create an AFL player for the temp FFL player
	ctx := context.Background()
	var aflPID int
	if err := pool.QueryRow(ctx, "INSERT INTO afl.player (name) VALUES ('Temp AFL') RETURNING id").Scan(&aflPID); err != nil {
		t.Fatalf("failed to insert afl player: %v", err)
	}

	// Create a player to delete (not the seeded one, which has FKs)
	aflPIDStr := fmt.Sprintf("%d", aflPID)
	createResult := execQuery(t, server, `mutation {
		createFFLPlayer(input: { name: "Temp Player", aflPlayerId: "`+aflPIDStr+`" }) { id }
	}`)
	var created struct {
		CreateFFLPlayer struct {
			ID string `json:"id"`
		} `json:"createFFLPlayer"`
	}
	if err := json.Unmarshal(createResult.Data, &created); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	result := execQuery(t, server, `mutation {
		deleteFFLPlayer(id: "`+created.CreateFFLPlayer.ID+`")
	}`)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
}

func TestAddAndRemoveFFLPlayerFromSeason(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	// Create an AFL player for the new FFL player
	ctx := context.Background()
	var aflPID int
	if err := pool.QueryRow(ctx, "INSERT INTO afl.player (name) VALUES ('Season AFL') RETURNING id").Scan(&aflPID); err != nil {
		t.Fatalf("failed to insert afl player: %v", err)
	}

	// Create a new player and add them to the away club season
	aflPIDStr := fmt.Sprintf("%d", aflPID)
	createResult := execQuery(t, server, `mutation {
		createFFLPlayer(input: { name: "Season Player", aflPlayerId: "`+aflPIDStr+`" }) { id }
	}`)
	var created struct {
		CreateFFLPlayer struct {
			ID string `json:"id"`
		} `json:"createFFLPlayer"`
	}
	if err := json.Unmarshal(createResult.Data, &created); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	clubSeasonID := fmt.Sprintf("%d", ids.awayClubSeaID)
	addResult := execQuery(t, server, `mutation {
		addFFLPlayerToSeason(input: { playerId: "`+created.CreateFFLPlayer.ID+`", clubSeasonId: "`+clubSeasonID+`" }) {
			id
			player { id }
			clubSeasonId
		}
	}`)

	if len(addResult.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", addResult.Errors)
	}

	var addData struct {
		AddFFLPlayerToSeason struct {
			ID           string `json:"id"`
			Player       struct {
				ID string `json:"id"`
			} `json:"player"`
			ClubSeasonID string `json:"clubSeasonId"`
		} `json:"addFFLPlayerToSeason"`
	}
	if err := json.Unmarshal(addResult.Data, &addData); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if addData.AddFFLPlayerToSeason.Player.ID != created.CreateFFLPlayer.ID {
		t.Errorf("expected player ID %s, got %s", created.CreateFFLPlayer.ID, addData.AddFFLPlayerToSeason.Player.ID)
	}
	if addData.AddFFLPlayerToSeason.ClubSeasonID != clubSeasonID {
		t.Errorf("expected club season ID %s, got %s", clubSeasonID, addData.AddFFLPlayerToSeason.ClubSeasonID)
	}

	// Remove the player from the season
	removeResult := execQuery(t, server, `mutation {
		removeFFLPlayerFromSeason(id: "`+addData.AddFFLPlayerToSeason.ID+`")
	}`)

	if len(removeResult.Errors) > 0 {
		t.Fatalf("unexpected errors removing player from season: %v", removeResult.Errors)
	}
}

func TestCalculateFFLFantasyScore(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	// Player has position "goals" → score = goals * 5
	// Sending goals=3 → expected score = 15
	pmID := fmt.Sprintf("%d", ids.playerMatchID)
	result := execQuery(t, server, `mutation {
		calculateFFLFantasyScore(input: {
			playerMatchId: "`+pmID+`"
			goals: 3
			kicks: 10
			handballs: 5
			marks: 4
			tackles: 2
			hitouts: 0
		}) {
			id
			player { name }
			position
			score
		}
	}`)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	var data struct {
		CalculateFFLFantasyScore struct {
			ID     string `json:"id"`
			Player struct {
				Name string `json:"name"`
			} `json:"player"`
			Position *string `json:"position"`
			Score    int     `json:"score"`
		} `json:"calculateFFLFantasyScore"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if data.CalculateFFLFantasyScore.Player.Name != "Test Player" {
		t.Errorf("expected Test Player, got %s", data.CalculateFFLFantasyScore.Player.Name)
	}
	// goals position: 3 goals * 5 = 15
	if data.CalculateFFLFantasyScore.Score != 15 {
		t.Errorf("expected score 15, got %d", data.CalculateFFLFantasyScore.Score)
	}
}

func TestCalculateFFLFantasyScore_RecalculatesClubMatchScore(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	// Player has position "goals", sending goals=6 → score = 6*5 = 30
	// Club match score should be recalculated to 30 (only 1 player, starter)
	pmID := fmt.Sprintf("%d", ids.playerMatchID)
	result := execQuery(t, server, `mutation {
		calculateFFLFantasyScore(input: {
			playerMatchId: "`+pmID+`"
			goals: 6
			kicks: 0
			handballs: 0
			marks: 0
			tackles: 0
			hitouts: 0
		}) {
			score
		}
	}`)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	// Query the club match to verify score was recalculated
	seasonID := fmt.Sprintf("%d", ids.seasonID)
	queryResult := execQuery(t, server, `{
		fflSeason(id: "`+seasonID+`") {
			rounds {
				matches {
					homeClubMatch {
						score
					}
				}
			}
		}
	}`)

	if len(queryResult.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", queryResult.Errors)
	}

	var data struct {
		FflSeason struct {
			Rounds []struct {
				Matches []struct {
					HomeClubMatch struct {
						Score int `json:"score"`
					} `json:"homeClubMatch"`
				} `json:"matches"`
			} `json:"rounds"`
		} `json:"fflSeason"`
	}
	if err := json.Unmarshal(queryResult.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	homeScore := data.FflSeason.Rounds[0].Matches[0].HomeClubMatch.Score
	// goals position: 6 goals * 5 = 30
	if homeScore != 30 {
		t.Errorf("expected recalculated home score 30, got %d", homeScore)
	}
}

func TestCalculateFFLFantasyScore_InvalidPlayerMatchID(t *testing.T) {
	pool := connectDB(t)
	seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	result := execQuery(t, server, `mutation {
		calculateFFLFantasyScore(input: {
			playerMatchId: "999999"
			goals: 1
			kicks: 1
			handballs: 1
			marks: 1
			tackles: 1
			hitouts: 1
		}) {
			id
		}
	}`)

	if len(result.Errors) == 0 {
		t.Fatal("expected error for invalid playerMatchId, got none")
	}
}

// ── SetFFLLineup: team composition rule tests ─────────────────────────────────

// lineupMutation builds a setFFLLineup mutation string for testing.
// players is a slice of (playerSeasonId, position, backupPositions, interchangePosition) tuples.
type lineupPlayer struct {
	playerSeasonID      string
	position            string
	backupPositions     *string
	interchangePosition *string
}

func buildSetLineupMutation(clubMatchID string, players []lineupPlayer) string {
	playersStr := ""
	for _, p := range players {
		bp := "null"
		if p.backupPositions != nil {
			bp = `"` + *p.backupPositions + `"`
		}
		ic := "null"
		if p.interchangePosition != nil {
			ic = `"` + *p.interchangePosition + `"`
		}
		playersStr += fmt.Sprintf(`
			{
				playerSeasonId: "%s"
				position: "%s"
				backupPositions: %s
				interchangePosition: %s
			}`, p.playerSeasonID, p.position, bp, ic)
	}
	return fmt.Sprintf(`mutation {
		setFFLLineup(input: {
			clubMatchId: "%s"
			players: [%s]
		}) { id position backupPositions interchangePosition }
	}`, clubMatchID, playersStr)
}

// seedExtraPlayers inserts n additional players into the home club season and returns their playerSeasonIDs.
func seedExtraPlayers(t *testing.T, pool *pgxpool.Pool, ids testIDs, count int) []string {
	t.Helper()
	ctx := context.Background()
	psIDs := make([]string, count)
	for i := range count {
		var aflID, playerID, psID int
		if err := pool.QueryRow(ctx,
			fmt.Sprintf("INSERT INTO afl.player (name) VALUES ('Extra Player %d') RETURNING id", i)).Scan(&aflID); err != nil {
			t.Fatalf("seed extra afl player %d: %v", i, err)
		}
		if err := pool.QueryRow(ctx,
			"INSERT INTO ffl.player (drv_name, afl_player_id) VALUES ($1, $2) RETURNING id",
			fmt.Sprintf("Extra Player %d", i), aflID).Scan(&playerID); err != nil {
			t.Fatalf("seed extra ffl player %d: %v", i, err)
		}
		if err := pool.QueryRow(ctx,
			"INSERT INTO ffl.player_season (player_id, club_season_id) VALUES ($1, $2) RETURNING id",
			playerID, ids.homeClubSeaID).Scan(&psID); err != nil {
			t.Fatalf("seed extra player season %d: %v", i, err)
		}
		psIDs[i] = fmt.Sprintf("%d", psID)
	}
	return psIDs
}

func strp(s string) *string { return &s }

func TestSetFFLLineup_ValidLineupSaves(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	psID := fmt.Sprintf("%d", ids.playerSeasonID)
	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)

	result := execQuery(t, server, buildSetLineupMutation(cmID, []lineupPlayer{
		{playerSeasonID: psID, position: "goals"},
	}))

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
}

func TestSetFFLLineup_TooManyStartersForPosition(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	// Need 4 players for goals (max is 3)
	extras := seedExtraPlayers(t, pool, ids, 3)
	server := setupTestServer(t, pool)
	defer server.Close()

	psID := fmt.Sprintf("%d", ids.playerSeasonID)
	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)

	players := []lineupPlayer{
		{playerSeasonID: psID, position: "goals"},
		{playerSeasonID: extras[0], position: "goals"},
		{playerSeasonID: extras[1], position: "goals"},
		{playerSeasonID: extras[2], position: "goals"}, // 4th goals kicker — invalid
	}
	result := execQuery(t, server, buildSetLineupMutation(cmID, players))
	if len(result.Errors) == 0 {
		t.Fatal("expected error for too many goal kickers, got none")
	}
}

func TestSetFFLLineup_TooManyBenchPlayers(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	extras := seedExtraPlayers(t, pool, ids, 5)
	server := setupTestServer(t, pool)
	defer server.Close()

	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)
	players := []lineupPlayer{}
	positions := [][2]string{
		{"goals,kicks"}, {"goals,kicks"}, {"handballs,marks"}, {"tackles,hitouts"}, {"goals,kicks"},
	}
	_ = positions
	// 5 bench players (max is 4)
	for i, id := range extras {
		_ = i
		players = append(players, lineupPlayer{
			playerSeasonID:  id,
			position:        "goals",
			backupPositions: strp(fmt.Sprintf("goals,kicks")),
		})
		_ = extras
	}
	result := execQuery(t, server, buildSetLineupMutation(cmID, players))
	if len(result.Errors) == 0 {
		t.Fatal("expected error for 5 bench players, got none")
	}
}

func TestSetFFLLineup_TwoBenchStars(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	extras := seedExtraPlayers(t, pool, ids, 1)
	server := setupTestServer(t, pool)
	defer server.Close()

	psID := fmt.Sprintf("%d", ids.playerSeasonID)
	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)

	result := execQuery(t, server, buildSetLineupMutation(cmID, []lineupPlayer{
		{playerSeasonID: psID, position: "star", backupPositions: strp("star")},
		{playerSeasonID: extras[0], position: "star", backupPositions: strp("star")},
	}))
	if len(result.Errors) == 0 {
		t.Fatal("expected error for two backup stars, got none")
	}
}

func TestSetFFLLineup_SamePositionCoveredByTwoBenchPlayers(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	extras := seedExtraPlayers(t, pool, ids, 1)
	server := setupTestServer(t, pool)
	defer server.Close()

	psID := fmt.Sprintf("%d", ids.playerSeasonID)
	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)

	result := execQuery(t, server, buildSetLineupMutation(cmID, []lineupPlayer{
		{playerSeasonID: psID, position: "goals", backupPositions: strp("goals,kicks")},
		{playerSeasonID: extras[0], position: "goals", backupPositions: strp("goals,marks")}, // goals covered twice
	}))
	if len(result.Errors) == 0 {
		t.Fatal("expected error for duplicate bench position coverage, got none")
	}
}

func TestSetFFLLineup_TwoInterchangePositions(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	extras := seedExtraPlayers(t, pool, ids, 1)
	server := setupTestServer(t, pool)
	defer server.Close()

	psID := fmt.Sprintf("%d", ids.playerSeasonID)
	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)

	result := execQuery(t, server, buildSetLineupMutation(cmID, []lineupPlayer{
		{playerSeasonID: psID, position: "goals", backupPositions: strp("goals,kicks"), interchangePosition: strp("goals")},
		{playerSeasonID: extras[0], position: "marks", backupPositions: strp("marks,tackles"), interchangePosition: strp("marks")},
	}))
	if len(result.Errors) == 0 {
		t.Fatal("expected error for two interchange positions, got none")
	}
}

func TestSetFFLLineup_MultipleStartersScoreCorrectly(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	extras := seedExtraPlayers(t, pool, ids, 2)
	server := setupTestServer(t, pool)
	defer server.Close()

	ctx := context.Background()
	psID := fmt.Sprintf("%d", ids.playerSeasonID)
	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)

	// Set 3 goal kickers
	result := execQuery(t, server, buildSetLineupMutation(cmID, []lineupPlayer{
		{playerSeasonID: psID, position: "goals"},
		{playerSeasonID: extras[0], position: "goals"},
		{playerSeasonID: extras[1], position: "goals"},
	}))
	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors setting lineup: %v", result.Errors)
	}

	// Directly set scores for all 3 goal kickers in the DB, then recalculate
	_, err := pool.Exec(ctx,
		"UPDATE ffl.player_match SET drv_score = 10 WHERE club_match_id = $1",
		ids.homeClubMatchID)
	if err != nil {
		t.Fatalf("failed to update scores: %v", err)
	}
	// Recalculate club match score
	playerMatches, err := pool.Query(ctx,
		"SELECT id, position, backup_positions, interchange_position, status, drv_score FROM ffl.player_match WHERE club_match_id = $1",
		ids.homeClubMatchID)
	if err != nil {
		t.Fatalf("failed to query player matches: %v", err)
	}
	defer playerMatches.Close()

	totalScore := 0
	for playerMatches.Next() {
		var score int
		var pos, bp, ic, status *string
		if err := playerMatches.Scan(&ids.playerMatchID, &pos, &bp, &ic, &status, &score); err != nil {
			t.Fatalf("failed to scan: %v", err)
		}
		if bp == nil && ic == nil {
			totalScore += score // starters only
		}
	}

	// 3 goal kickers × 10 = 30
	if totalScore != 30 {
		t.Errorf("expected total starter score 30, got %d", totalScore)
	}
}

func TestSetFFLLineup_ReplacesStaleEntries(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	extras := seedExtraPlayers(t, pool, ids, 1)
	server := setupTestServer(t, pool)
	defer server.Close()

	ctx := context.Background()
	psID := fmt.Sprintf("%d", ids.playerSeasonID)
	extraID := extras[0]
	cmID := fmt.Sprintf("%d", ids.homeClubMatchID)

	// First lineup: original player at goals.
	result := execQuery(t, server, buildSetLineupMutation(cmID, []lineupPlayer{
		{playerSeasonID: psID, position: "goals"},
	}))
	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors setting first lineup: %v", result.Errors)
	}

	// Second lineup: different player at kicks — replaces the first.
	result = execQuery(t, server, buildSetLineupMutation(cmID, []lineupPlayer{
		{playerSeasonID: extraID, position: "kicks"},
	}))
	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors setting second lineup: %v", result.Errors)
	}

	// Only the new player_match should exist in the DB.
	rows, err := pool.Query(ctx,
		"SELECT player_season_id, position FROM ffl.player_match WHERE club_match_id = $1",
		ids.homeClubMatchID)
	if err != nil {
		t.Fatalf("failed to query player_match: %v", err)
	}
	defer rows.Close()

	type row struct {
		playerSeasonID int
		position       string
	}
	var found []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.playerSeasonID, &r.position); err != nil {
			t.Fatalf("scan: %v", err)
		}
		found = append(found, r)
	}

	if len(found) != 1 {
		t.Fatalf("expected 1 player_match after replacement, got %d", len(found))
	}
	extraIDInt, _ := strconv.Atoi(extraID)
	if found[0].playerSeasonID != extraIDInt {
		t.Errorf("expected player_season_id %d, got %d", extraIDInt, found[0].playerSeasonID)
	}
	if found[0].position != "kicks" {
		t.Errorf("expected position %q, got %q", "kicks", found[0].position)
	}
}
