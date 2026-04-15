package graphql_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/search/internal/domain"
	gql "xffl/services/search/internal/interface/graphql"
)

type stubRepo struct {
	result domain.SearchResult
	err    error
	got    domain.SearchQuery
}

func (s *stubRepo) Index(context.Context, domain.SearchDocument) error { return nil }
func (s *stubRepo) Search(_ context.Context, q domain.SearchQuery) (domain.SearchResult, error) {
	s.got = q
	return s.result, s.err
}

func newServer(repo *stubRepo) *httptest.Server {
	resolver := &gql.Resolver{Repo: repo}
	srv := handler.NewDefaultServer(gql.NewExecutableSchema(gql.Config{Resolvers: resolver}))
	return httptest.NewServer(srv)
}

func postGraphQL(t *testing.T, url, query string) map[string]any {
	t.Helper()
	body := `{"query":"` + query + `"}`
	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	return result
}

func TestSearch_FullParams(t *testing.T) {
	repo := &stubRepo{
		result: domain.SearchResult{
			Total: 1,
			Documents: []domain.SearchDocument{
				{ID: "afl_player_match_1", Source: "afl", Type: "player_match", Data: map[string]any{"kicks": float64(10)}},
			},
		},
	}
	srv := newServer(repo)
	defer srv.Close()

	result := postGraphQL(t, srv.URL, `{ search(q: \"hello\", source: \"afl\", type: \"player_match\") { total documents { id source type data } } }`)

	assert.Equal(t, "hello", repo.got.Q)
	assert.Equal(t, "afl", repo.got.Source)
	assert.Equal(t, "player_match", repo.got.Type)

	data := result["data"].(map[string]any)
	search := data["search"].(map[string]any)
	assert.Equal(t, float64(1), search["total"])

	docs := search["documents"].([]any)
	assert.Len(t, docs, 1)
	doc := docs[0].(map[string]any)
	assert.Equal(t, "afl_player_match_1", doc["id"])
	assert.Equal(t, "afl", doc["source"])
}

func TestSearch_NoParams(t *testing.T) {
	repo := &stubRepo{result: domain.SearchResult{Total: 0, Documents: []domain.SearchDocument{}}}
	srv := newServer(repo)
	defer srv.Close()

	result := postGraphQL(t, srv.URL, `{ search { total documents { id } } }`)

	assert.Equal(t, "", repo.got.Q)
	assert.Equal(t, "", repo.got.Source)
	assert.Equal(t, "", repo.got.Type)

	data := result["data"].(map[string]any)
	search := data["search"].(map[string]any)
	assert.Equal(t, float64(0), search["total"])
}

func TestSearch_RepoError(t *testing.T) {
	repo := &stubRepo{err: assert.AnError}
	srv := newServer(repo)
	defer srv.Close()

	result := postGraphQL(t, srv.URL, `{ search(q: \"test\") { total } }`)

	errors, ok := result["errors"].([]any)
	require.True(t, ok)
	assert.NotEmpty(t, errors)
}
