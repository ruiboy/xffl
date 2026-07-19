//go:build integration

package graphql_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSingleEntityQueries_ReturnNullForUnknownID proves that a well-formed id
// with no matching row resolves to null rather than a GraphQL error. The
// frontend treats "settled, no error, no entity" as its 404 signal, so an error
// here would surface as a red banner instead of the not-found page.
func TestSingleEntityQueries_ReturnNullForUnknownID(t *testing.T) {
	pool := connectDB(t)
	seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	// An id far beyond anything the seed creates.
	const missingID = "99999999"

	cases := []struct {
		field string
		query string
	}{
		{"fflSeason", `{ fflSeason(id: "` + missingID + `") { id } }`},
		{"fflRound", `{ fflRound(id: "` + missingID + `") { id } }`},
		{"fflMatch", `{ fflMatch(id: "` + missingID + `") { id } }`},
		{"fflClubSeason", `{ fflClubSeason(id: "` + missingID + `") { id } }`},
		{"fflClubMatch", `{ fflClubMatch(id: "` + missingID + `") { id } }`},
	}

	for _, tc := range cases {
		t.Run(tc.field+" resolves to null instead of erroring", func(t *testing.T) {
			result := execQuery(t, server, tc.query)
			require.Empty(t, result.Errors, "unknown id must not produce a GraphQL error")

			var data map[string]*json.RawMessage
			require.NoError(t, json.Unmarshal(result.Data, &data))
			assert.Nil(t, data[tc.field], "expected %s to be null", tc.field)
		})
	}
}

// TestSingleEntityQueries_ReturnNullForMalformedID covers the other route into
// the not-found page: an id that isn't parseable at all (e.g. /ffl/matches/banana).
// From the user's side that is the same "no such page", so it must not error either.
func TestSingleEntityQueries_ReturnNullForMalformedID(t *testing.T) {
	pool := connectDB(t)
	seedTestData(t, pool)
	server := setupTestServer(t, pool)
	defer server.Close()

	cases := []struct {
		field string
		query string
	}{
		{"fflSeason", `{ fflSeason(id: "banana") { id } }`},
		{"fflRound", `{ fflRound(id: "banana") { id } }`},
		{"fflMatch", `{ fflMatch(id: "banana") { id } }`},
		{"fflClubSeason", `{ fflClubSeason(id: "banana") { id } }`},
		{"fflClubMatch", `{ fflClubMatch(id: "banana") { id } }`},
	}

	for _, tc := range cases {
		t.Run(tc.field+" treats an unparseable id as absent", func(t *testing.T) {
			result := execQuery(t, server, tc.query)
			require.Empty(t, result.Errors, "malformed id must not produce a GraphQL error")

			var data map[string]*json.RawMessage
			require.NoError(t, json.Unmarshal(result.Data, &data))
			assert.Nil(t, data[tc.field], "expected %s to be null", tc.field)
		})
	}
}
