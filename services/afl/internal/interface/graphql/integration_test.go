package graphql_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	gqlhandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/jackc/pgx/v5/pgxpool"

	"xffl/services/afl/internal/application"
	pg "xffl/services/afl/internal/infrastructure/postgres"
	gql "xffl/services/afl/internal/interface/graphql"
)

func setupTestServer(t *testing.T) *httptest.Server {
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

	queries := application.NewQueries(
		pg.NewClubRepository(pool),
		pg.NewSeasonRepository(pool),
		pg.NewRoundRepository(pool),
		pg.NewMatchRepository(pool),
		pg.NewClubSeasonRepository(pool),
		pg.NewClubMatchRepository(pool),
		pg.NewPlayerRepository(pool),
		pg.NewPlayerMatchRepository(pool),
		pg.NewPlayerSeasonRepository(pool),
	)

	resolver := &gql.Resolver{Queries: queries}
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
	server := setupTestServer(t)
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

	if len(data.AflClubs) != 18 {
		t.Errorf("expected 18 clubs, got %d", len(data.AflClubs))
	}

	// Clubs are ordered by name, first should be Adelaide Crows
	if len(data.AflClubs) > 0 && data.AflClubs[0].Name != "Adelaide Crows" {
		t.Errorf("expected first club to be Adelaide Crows, got %s", data.AflClubs[0].Name)
	}
}

func TestAflSeasons(t *testing.T) {
	server := setupTestServer(t)
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

	if len(data.AflSeasons) == 0 {
		t.Error("expected at least one season")
	}
}

func TestAflSeasonWithLadder(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	// First get a season ID
	seasonsResult := execQuery(t, server, `{ aflSeasons { id } }`)
	var seasons struct {
		AflSeasons []struct {
			ID string `json:"id"`
		} `json:"aflSeasons"`
	}
	if err := json.Unmarshal(seasonsResult.Data, &seasons); err != nil {
		t.Fatalf("failed to unmarshal seasons: %v", err)
	}
	if len(seasons.AflSeasons) == 0 {
		t.Skip("no seasons in database")
	}

	seasonID := seasons.AflSeasons[0].ID
	result := execQuery(t, server, `{ aflSeason(id: "`+seasonID+`") { id name ladder { club { name } played won lost premiershipPoints } } }`)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	var data struct {
		AflSeason struct {
			ID     string `json:"id"`
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

	if len(data.AflSeason.Ladder) == 0 {
		t.Error("expected ladder to have entries")
	}

	// Ladder should be ordered by premiership points descending
	for i := 1; i < len(data.AflSeason.Ladder); i++ {
		if data.AflSeason.Ladder[i].PremiershipPoints > data.AflSeason.Ladder[i-1].PremiershipPoints {
			t.Errorf("ladder not sorted by premiership points: %d > %d at position %d",
				data.AflSeason.Ladder[i].PremiershipPoints, data.AflSeason.Ladder[i-1].PremiershipPoints, i)
		}
	}
}

func TestAflSeasonGraphTraversal(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	// Get season ID
	seasonsResult := execQuery(t, server, `{ aflSeasons { id } }`)
	var seasons struct {
		AflSeasons []struct {
			ID string `json:"id"`
		} `json:"aflSeasons"`
	}
	if err := json.Unmarshal(seasonsResult.Data, &seasons); err != nil {
		t.Fatalf("failed to unmarshal seasons: %v", err)
	}
	if len(seasons.AflSeasons) == 0 {
		t.Skip("no seasons in database")
	}

	// Full graph traversal: season → rounds → matches → clubMatches → playerMatches
	seasonID := seasons.AflSeasons[0].ID
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

	if len(data.AflSeason.Rounds) == 0 {
		t.Error("expected at least one round")
	}

	// Verify we can traverse the full graph
	for _, round := range data.AflSeason.Rounds {
		if round.Name == "" {
			t.Error("round name should not be empty")
		}
		for _, match := range round.Matches {
			if match.HomeClubMatch != nil && match.HomeClubMatch.Club.Name == "" {
				t.Error("home club name should not be empty")
			}
			if match.AwayClubMatch != nil && match.AwayClubMatch.Club.Name == "" {
				t.Error("away club name should not be empty")
			}
		}
	}
}

func TestAflClubWithPlayers(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	// Get a club ID (Adelaide Crows should have Jordan Dawson)
	clubsResult := execQuery(t, server, `{ aflClubs { id name } }`)
	var clubs struct {
		AflClubs []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"aflClubs"`
	}
	if err := json.Unmarshal(clubsResult.Data, &clubs); err != nil {
		t.Fatalf("failed to unmarshal clubs: %v", err)
	}

	var crowsID string
	for _, c := range clubs.AflClubs {
		if c.Name == "Adelaide Crows" {
			crowsID = c.ID
			break
		}
	}
	if crowsID == "" {
		t.Skip("Adelaide Crows not found in database")
	}

	result := execQuery(t, server, `{ aflClub(id: "`+crowsID+`") { name players { name } } }`)

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

	if data.AflClub.Name != "Adelaide Crows" {
		t.Errorf("expected Adelaide Crows, got %s", data.AflClub.Name)
	}

	// Jordan Dawson should be in the players list
	found := false
	for _, p := range data.AflClub.Players {
		if p.Name == "Jordan Dawson" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Jordan Dawson in Adelaide Crows players")
	}
}
