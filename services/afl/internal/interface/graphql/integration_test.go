package graphql_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	gqlhandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/jackc/pgx/v5/pgxpool"

	"xffl/services/afl/internal/application"
	pg "xffl/services/afl/internal/infrastructure/postgres"
	"xffl/services/afl/internal/infrastructure/postgres/sqlcgen"
	gql "xffl/services/afl/internal/interface/graphql"
)

// testIDs holds IDs of rows inserted by seedTestData, used by tests to query known data.
type testIDs struct {
	leagueID       int
	seasonID       int
	roundID        int
	homeClubID     int
	awayClubID     int
	homeClubSeaID  int
	awayClubSeaID  int
	matchID        int
	homeClubMatchID int
	awayClubMatchID int
	playerID       int
	playerSeasonID int
	playerMatchID  int
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
		"afl.player_match", "afl.player_season", "afl.player",
		"afl.club_match", "afl.match", "afl.club_season",
		"afl.club", "afl.round", "afl.season", "afl.league",
	}
	for _, table := range tables {
		if _, err := pool.Exec(ctx, fmt.Sprintf("TRUNCATE %s CASCADE", table)); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	// League
	err := pool.QueryRow(ctx,
		"INSERT INTO afl.league (name) VALUES ('Test AFL') RETURNING id").Scan(&ids.leagueID)
	if err != nil {
		t.Fatalf("failed to insert league: %v", err)
	}

	// Season
	err = pool.QueryRow(ctx,
		"INSERT INTO afl.season (name, league_id) VALUES ('Test 2025', $1) RETURNING id",
		ids.leagueID).Scan(&ids.seasonID)
	if err != nil {
		t.Fatalf("failed to insert season: %v", err)
	}

	// Round
	err = pool.QueryRow(ctx,
		"INSERT INTO afl.round (name, season_id) VALUES ('Round 1', $1) RETURNING id",
		ids.seasonID).Scan(&ids.roundID)
	if err != nil {
		t.Fatalf("failed to insert round: %v", err)
	}

	// Two clubs
	err = pool.QueryRow(ctx,
		"INSERT INTO afl.club (name) VALUES ('Test Hawks') RETURNING id").Scan(&ids.homeClubID)
	if err != nil {
		t.Fatalf("failed to insert home club: %v", err)
	}
	err = pool.QueryRow(ctx,
		"INSERT INTO afl.club (name) VALUES ('Test Cats') RETURNING id").Scan(&ids.awayClubID)
	if err != nil {
		t.Fatalf("failed to insert away club: %v", err)
	}

	// Club seasons (home team higher on ladder)
	err = pool.QueryRow(ctx,
		`INSERT INTO afl.club_season (club_id, season_id, drv_played, drv_won, drv_lost, drv_drawn, drv_for, drv_against, drv_premiership_points)
		 VALUES ($1, $2, 5, 4, 1, 0, 500, 400, 16) RETURNING id`,
		ids.homeClubID, ids.seasonID).Scan(&ids.homeClubSeaID)
	if err != nil {
		t.Fatalf("failed to insert home club season: %v", err)
	}
	err = pool.QueryRow(ctx,
		`INSERT INTO afl.club_season (club_id, season_id, drv_played, drv_won, drv_lost, drv_drawn, drv_for, drv_against, drv_premiership_points)
		 VALUES ($1, $2, 5, 3, 2, 0, 450, 420, 12) RETURNING id`,
		ids.awayClubID, ids.seasonID).Scan(&ids.awayClubSeaID)
	if err != nil {
		t.Fatalf("failed to insert away club season: %v", err)
	}

	// Match
	err = pool.QueryRow(ctx,
		"INSERT INTO afl.match (round_id, venue, start_dt) VALUES ($1, 'Test Ground', '2025-06-15 14:00:00') RETURNING id",
		ids.roundID).Scan(&ids.matchID)
	if err != nil {
		t.Fatalf("failed to insert match: %v", err)
	}

	// Club matches
	err = pool.QueryRow(ctx,
		"INSERT INTO afl.club_match (match_id, club_season_id, drv_score, rushed_behinds) VALUES ($1, $2, 85, 2) RETURNING id",
		ids.matchID, ids.homeClubSeaID).Scan(&ids.homeClubMatchID)
	if err != nil {
		t.Fatalf("failed to insert home club match: %v", err)
	}
	err = pool.QueryRow(ctx,
		"INSERT INTO afl.club_match (match_id, club_season_id, drv_score, rushed_behinds) VALUES ($1, $2, 72, 1) RETURNING id",
		ids.matchID, ids.awayClubSeaID).Scan(&ids.awayClubMatchID)
	if err != nil {
		t.Fatalf("failed to insert away club match: %v", err)
	}

	// Link match to club matches
	_, err = pool.Exec(ctx,
		"UPDATE afl.match SET home_club_match_id = $1, away_club_match_id = $2 WHERE id = $3",
		ids.homeClubMatchID, ids.awayClubMatchID, ids.matchID)
	if err != nil {
		t.Fatalf("failed to link match to club matches: %v", err)
	}

	// Player
	err = pool.QueryRow(ctx,
		"INSERT INTO afl.player (name) VALUES ('Test Player') RETURNING id").Scan(&ids.playerID)
	if err != nil {
		t.Fatalf("failed to insert player: %v", err)
	}

	// Player season
	err = pool.QueryRow(ctx,
		"INSERT INTO afl.player_season (player_id, club_season_id) VALUES ($1, $2) RETURNING id",
		ids.playerID, ids.homeClubSeaID).Scan(&ids.playerSeasonID)
	if err != nil {
		t.Fatalf("failed to insert player season: %v", err)
	}

	// Player match (10 kicks, 5 handballs, 3 marks, 0 hitouts, 2 tackles, 2 goals, 1 behind)
	err = pool.QueryRow(ctx,
		`INSERT INTO afl.player_match (club_match_id, player_season_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
		 VALUES ($1, $2, 10, 5, 3, 0, 2, 2, 1) RETURNING id`,
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

func TestAflClubs(t *testing.T) {
	pool := connectDB(t)
	seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	result := execQuery(t, server, `{ aflClubs { id name } }`)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	var data struct {
		AflClubs []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"aflClubs"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if len(data.AflClubs) != 2 {
		t.Errorf("expected 2 clubs, got %d", len(data.AflClubs))
	}

	// Clubs ordered by name: Test Cats before Test Hawks
	if len(data.AflClubs) == 2 {
		if data.AflClubs[0].Name != "Test Cats" {
			t.Errorf("expected first club Test Cats, got %s", data.AflClubs[0].Name)
		}
		if data.AflClubs[1].Name != "Test Hawks" {
			t.Errorf("expected second club Test Hawks, got %s", data.AflClubs[1].Name)
		}
	}
}

func TestAflSeasons(t *testing.T) {
	pool := connectDB(t)
	seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	result := execQuery(t, server, `{ aflSeasons { id name } }`)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	var data struct {
		AflSeasons []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"aflSeasons"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if len(data.AflSeasons) != 1 {
		t.Errorf("expected 1 season, got %d", len(data.AflSeasons))
	}
	if len(data.AflSeasons) > 0 && data.AflSeasons[0].Name != "Test 2025" {
		t.Errorf("expected season name Test 2025, got %s", data.AflSeasons[0].Name)
	}
}

func TestAflSeasonWithLadder(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	seasonID := fmt.Sprintf("%d", ids.seasonID)
	result := execQuery(t, server, `{ aflSeason(id: "`+seasonID+`") { name ladder { club { name } played won lost premiershipPoints } } }`)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	var data struct {
		AflSeason struct {
			Name   string `json:"name"`
			Ladder []struct {
				Club struct {
					Name string `json:"name"`
				} `json:"club"`
				Played            int `json:"played"`
				Won               int `json:"won"`
				Lost              int `json:"lost"`
				PremiershipPoints int `json:"premiershipPoints"`
			} `json:"ladder"`
		} `json:"aflSeason"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if len(data.AflSeason.Ladder) != 2 {
		t.Fatalf("expected 2 ladder entries, got %d", len(data.AflSeason.Ladder))
	}

	// Ladder sorted by premiership points: Hawks (16) before Cats (12)
	if data.AflSeason.Ladder[0].Club.Name != "Test Hawks" {
		t.Errorf("expected Test Hawks first on ladder, got %s", data.AflSeason.Ladder[0].Club.Name)
	}
	if data.AflSeason.Ladder[0].PremiershipPoints != 16 {
		t.Errorf("expected 16 premiership points, got %d", data.AflSeason.Ladder[0].PremiershipPoints)
	}
	if data.AflSeason.Ladder[1].Club.Name != "Test Cats" {
		t.Errorf("expected Test Cats second on ladder, got %s", data.AflSeason.Ladder[1].Club.Name)
	}
}

func TestAflSeasonGraphTraversal(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	seasonID := fmt.Sprintf("%d", ids.seasonID)
	result := execQuery(t, server, `{
		aflSeason(id: "`+seasonID+`") {
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
							kicks
							handballs
							disposals
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
		AflSeason struct {
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
							Kicks     int `json:"kicks"`
							Handballs int `json:"handballs"`
							Disposals int `json:"disposals"`
							Score     int `json:"score"`
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
		} `json:"aflSeason"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	// Verify full graph: 1 round → 1 match → home/away club matches → player matches
	if len(data.AflSeason.Rounds) != 1 {
		t.Fatalf("expected 1 round, got %d", len(data.AflSeason.Rounds))
	}

	round := data.AflSeason.Rounds[0]
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
	if match.HomeClubMatch.Club.Name != "Test Hawks" {
		t.Errorf("expected home club Test Hawks, got %s", match.HomeClubMatch.Club.Name)
	}
	if match.HomeClubMatch.Score != 85 {
		t.Errorf("expected home score 85, got %d", match.HomeClubMatch.Score)
	}
	if match.AwayClubMatch == nil {
		t.Fatal("expected away club match")
	}
	if match.AwayClubMatch.Club.Name != "Test Cats" {
		t.Errorf("expected away club Test Cats, got %s", match.AwayClubMatch.Club.Name)
	}
	if match.AwayClubMatch.Score != 72 {
		t.Errorf("expected away score 72, got %d", match.AwayClubMatch.Score)
	}

	// Player match: 10 kicks + 5 handballs = 15 disposals, 2*6 + 1 = 13 score
	if len(match.HomeClubMatch.PlayerMatches) != 1 {
		t.Fatalf("expected 1 player match, got %d", len(match.HomeClubMatch.PlayerMatches))
	}
	pm := match.HomeClubMatch.PlayerMatches[0]
	if pm.Player.Name != "Test Player" {
		t.Errorf("expected Test Player, got %s", pm.Player.Name)
	}
	if pm.Kicks != 10 {
		t.Errorf("expected 10 kicks, got %d", pm.Kicks)
	}
	if pm.Disposals != 15 {
		t.Errorf("expected 15 disposals, got %d", pm.Disposals)
	}
	if pm.Score != 13 {
		t.Errorf("expected score 13, got %d", pm.Score)
	}
}

func TestAflClubWithPlayers(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	clubID := fmt.Sprintf("%d", ids.homeClubID)
	result := execQuery(t, server, `{ aflClub(id: "`+clubID+`") { name players { name } } }`)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	var data struct {
		AflClub struct {
			Name    string `json:"name"`
			Players []struct {
				Name string `json:"name"`
			} `json:"players"`
		} `json:"aflClub"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if data.AflClub.Name != "Test Hawks" {
		t.Errorf("expected Test Hawks, got %s", data.AflClub.Name)
	}

	if len(data.AflClub.Players) != 1 {
		t.Fatalf("expected 1 player, got %d", len(data.AflClub.Players))
	}
	if data.AflClub.Players[0].Name != "Test Player" {
		t.Errorf("expected Test Player, got %s", data.AflClub.Players[0].Name)
	}
}

func TestUpdatePlayerMatch_Update(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	// Update existing player match: change kicks from 10 to 20, leave other fields nil (unchanged)
	mutation := fmt.Sprintf(`mutation {
		updateAFLPlayerMatch(input: {
			playerSeasonId: "%d"
			clubMatchId: "%d"
			kicks: 20
		}) {
			id
			player { name }
			kicks
			handballs
			disposals
			goals
			behinds
			score
		}
	}`, ids.playerSeasonID, ids.homeClubMatchID)

	result := execQuery(t, server, mutation)
	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	var data struct {
		UpdateAFLPlayerMatch struct {
			ID     string `json:"id"`
			Player struct {
				Name string `json:"name"`
			} `json:"player"`
			Kicks     int `json:"kicks"`
			Handballs int `json:"handballs"`
			Disposals int `json:"disposals"`
			Goals     int `json:"goals"`
			Behinds   int `json:"behinds"`
			Score     int `json:"score"`
		} `json:"updateAFLPlayerMatch"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	pm := data.UpdateAFLPlayerMatch
	if pm.Player.Name != "Test Player" {
		t.Errorf("expected Test Player, got %s", pm.Player.Name)
	}
	if pm.Kicks != 20 {
		t.Errorf("expected 20 kicks, got %d", pm.Kicks)
	}
	// Handballs should remain unchanged at 5
	if pm.Handballs != 5 {
		t.Errorf("expected 5 handballs (unchanged), got %d", pm.Handballs)
	}
	// Disposals = 20 kicks + 5 handballs = 25
	if pm.Disposals != 25 {
		t.Errorf("expected 25 disposals, got %d", pm.Disposals)
	}
	// Goals and behinds unchanged: 2 goals, 1 behind → score = 13
	if pm.Score != 13 {
		t.Errorf("expected score 13, got %d", pm.Score)
	}
}

func TestUpdatePlayerMatch_Create(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	// Insert a second player + player_season for the away team
	ctx := context.Background()
	var player2ID, playerSeason2ID int
	err := pool.QueryRow(ctx,
		"INSERT INTO afl.player (name) VALUES ('Test Player 2') RETURNING id").Scan(&player2ID)
	if err != nil {
		t.Fatalf("failed to insert player 2: %v", err)
	}
	err = pool.QueryRow(ctx,
		"INSERT INTO afl.player_season (player_id, club_season_id) VALUES ($1, $2) RETURNING id",
		player2ID, ids.awayClubSeaID).Scan(&playerSeason2ID)
	if err != nil {
		t.Fatalf("failed to insert player season 2: %v", err)
	}

	// Create a new player match via mutation (no existing record for this player+clubMatch)
	mutation := fmt.Sprintf(`mutation {
		updateAFLPlayerMatch(input: {
			playerSeasonId: "%d"
			clubMatchId: "%d"
			kicks: 8
			handballs: 4
			marks: 6
			hitouts: 0
			tackles: 3
			goals: 1
			behinds: 2
		}) {
			id
			player { name }
			kicks
			handballs
			disposals
			goals
			behinds
			score
		}
	}`, playerSeason2ID, ids.awayClubMatchID)

	result := execQuery(t, server, mutation)
	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	var data struct {
		UpdateAFLPlayerMatch struct {
			ID     string `json:"id"`
			Player struct {
				Name string `json:"name"`
			} `json:"player"`
			Kicks     int `json:"kicks"`
			Handballs int `json:"handballs"`
			Disposals int `json:"disposals"`
			Goals     int `json:"goals"`
			Behinds   int `json:"behinds"`
			Score     int `json:"score"`
		} `json:"updateAFLPlayerMatch"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	pm := data.UpdateAFLPlayerMatch
	if pm.Player.Name != "Test Player 2" {
		t.Errorf("expected Test Player 2, got %s", pm.Player.Name)
	}
	if pm.Kicks != 8 {
		t.Errorf("expected 8 kicks, got %d", pm.Kicks)
	}
	if pm.Disposals != 12 {
		t.Errorf("expected 12 disposals, got %d", pm.Disposals)
	}
	// 1 goal * 6 + 2 behinds = 8
	if pm.Score != 8 {
		t.Errorf("expected score 8, got %d", pm.Score)
	}
}

func TestUpdatePlayerMatch_RecalculatesClubMatchScore(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	// Update the player's goals from 2 to 5, behinds from 1 to 3
	// New player score = 5*6 + 3 = 33
	// Club match score = player scores + rushed_behinds(2) = 33 + 2 = 35
	mutation := fmt.Sprintf(`mutation {
		updateAFLPlayerMatch(input: {
			playerSeasonId: "%d"
			clubMatchId: "%d"
			goals: 5
			behinds: 3
		}) {
			id
			goals
			behinds
			score
		}
	}`, ids.playerSeasonID, ids.homeClubMatchID)

	result := execQuery(t, server, mutation)
	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	// Now query the club match to verify drv_score was recalculated
	seasonID := fmt.Sprintf("%d", ids.seasonID)
	queryResult := execQuery(t, server, `{
		aflSeason(id: "`+seasonID+`") {
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
		AflSeason struct {
			Rounds []struct {
				Matches []struct {
					HomeClubMatch struct {
						Score int `json:"score"`
					} `json:"homeClubMatch"`
				} `json:"matches"`
			} `json:"rounds"`
		} `json:"aflSeason"`
	}
	if err := json.Unmarshal(queryResult.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	homeScore := data.AflSeason.Rounds[0].Matches[0].HomeClubMatch.Score
	// Expected: 5 goals * 6 + 3 behinds + 2 rushed_behinds = 35
	if homeScore != 35 {
		t.Errorf("expected recalculated home score 35, got %d", homeScore)
	}
}

func TestAflLatestRound(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	// Add a second round so we can verify "latest" picks the last one
	ctx := context.Background()
	var round2ID int
	err := pool.QueryRow(ctx,
		"INSERT INTO afl.round (name, season_id) VALUES ('Round 2', $1) RETURNING id",
		ids.seasonID).Scan(&round2ID)
	if err != nil {
		t.Fatalf("failed to insert round 2: %v", err)
	}

	result := execQuery(t, server, `{
		aflLatestRound {
			name
			season { name }
		}
	}`)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	var data struct {
		AflLatestRound struct {
			Name   string `json:"name"`
			Season struct {
				Name string `json:"name"`
			} `json:"season"`
		} `json:"aflLatestRound"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if data.AflLatestRound.Name != "Round 2" {
		t.Errorf("expected latest round 'Round 2', got %s", data.AflLatestRound.Name)
	}
	if data.AflLatestRound.Season.Name != "Test 2025" {
		t.Errorf("expected season 'Test 2025', got %s", data.AflLatestRound.Season.Name)
	}
}

func TestAflLatestRound_MultipleSeasons(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	// Add a second season with its own round
	ctx := context.Background()
	var season2ID, round2ID int
	err := pool.QueryRow(ctx,
		"INSERT INTO afl.season (name, league_id) VALUES ('Test 2026', $1) RETURNING id",
		ids.leagueID).Scan(&season2ID)
	if err != nil {
		t.Fatalf("failed to insert season 2: %v", err)
	}
	err = pool.QueryRow(ctx,
		"INSERT INTO afl.round (name, season_id) VALUES ('Round 1', $1) RETURNING id",
		season2ID).Scan(&round2ID)
	if err != nil {
		t.Fatalf("failed to insert round for season 2: %v", err)
	}

	result := execQuery(t, server, `{
		aflLatestRound {
			name
			season { name }
		}
	}`)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	var data struct {
		AflLatestRound struct {
			Name   string `json:"name"`
			Season struct {
				Name string `json:"name"`
			} `json:"season"`
		} `json:"aflLatestRound"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	// Should return round from the latest season (2026), not the first (2025)
	if data.AflLatestRound.Season.Name != "Test 2026" {
		t.Errorf("expected season 'Test 2026', got %s", data.AflLatestRound.Season.Name)
	}
	if data.AflLatestRound.Name != "Round 1" {
		t.Errorf("expected 'Round 1' from latest season, got %s", data.AflLatestRound.Name)
	}
}

func TestAflPlayerSearch(t *testing.T) {
	pool := connectDB(t)
	seedTestData(t, pool) // seeds "Test Player"
	server := setupTestServer(t, pool)
	defer server.Close()

	// Search matching
	result := execQuery(t, server, `{ aflPlayerSearch(query: "Test") { id name } }`)
	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	var data struct {
		AflPlayerSearch []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"aflPlayerSearch"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if len(data.AflPlayerSearch) != 1 {
		t.Fatalf("expected 1 result, got %d", len(data.AflPlayerSearch))
	}
	if data.AflPlayerSearch[0].Name != "Test Player" {
		t.Errorf("expected 'Test Player', got %s", data.AflPlayerSearch[0].Name)
	}

	// Search not matching
	result2 := execQuery(t, server, `{ aflPlayerSearch(query: "Nonexistent") { id name } }`)
	if len(result2.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result2.Errors)
	}

	var data2 struct {
		AflPlayerSearch []struct {
			ID string `json:"id"`
		} `json:"aflPlayerSearch"`
	}
	if err := json.Unmarshal(result2.Data, &data2); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	if len(data2.AflPlayerSearch) != 0 {
		t.Errorf("expected 0 results for non-matching search, got %d", len(data2.AflPlayerSearch))
	}
}

func TestUpdatePlayerMatch_InvalidPlayerSeasonID(t *testing.T) {
	pool := connectDB(t)
	ids := seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	mutation := fmt.Sprintf(`mutation {
		updateAFLPlayerMatch(input: {
			playerSeasonId: "999999"
			clubMatchId: "%d"
			kicks: 5
		}) {
			id
		}
	}`, ids.homeClubMatchID)

	result := execQuery(t, server, mutation)
	if len(result.Errors) == 0 {
		t.Fatal("expected error for invalid playerSeasonId, got none")
	}
}
