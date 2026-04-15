package rest_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/search/internal/domain"
	"xffl/services/search/internal/interface/rest"
)

// stubRepo is a test double for domain.DocumentRepository.
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

func TestServeSearch_FullParams(t *testing.T) {
	repo := &stubRepo{
		result: domain.SearchResult{
			Total: 1,
			Documents: []domain.SearchDocument{
				{ID: "afl_player_match_1", Source: "afl", Type: "player_match", Data: map[string]any{"kicks": float64(10)}},
			},
		},
	}
	h := rest.NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/search?q=hello&source=afl&type=player_match", nil)
	rec := httptest.NewRecorder()
	h.ServeSearch(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	// Verify query passed to repo.
	assert.Equal(t, "hello", repo.got.Q)
	assert.Equal(t, "afl", repo.got.Source)
	assert.Equal(t, "player_match", repo.got.Type)

	// Verify JSON response shape.
	var body struct {
		Total     int              `json:"total"`
		Documents []map[string]any `json:"documents"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Equal(t, 1, body.Total)
	assert.Len(t, body.Documents, 1)
	assert.Equal(t, "afl_player_match_1", body.Documents[0]["id"])
}

func TestServeSearch_NoParams(t *testing.T) {
	repo := &stubRepo{result: domain.SearchResult{Total: 0}}
	h := rest.NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/search", nil)
	rec := httptest.NewRecorder()
	h.ServeSearch(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "", repo.got.Q)
	assert.Equal(t, "", repo.got.Source)
	assert.Equal(t, "", repo.got.Type)
}

func TestServeSearch_RepoError(t *testing.T) {
	repo := &stubRepo{err: assert.AnError}
	h := rest.NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/search?q=test", nil)
	rec := httptest.NewRecorder()
	h.ServeSearch(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestServeSearch_MethodNotAllowed(t *testing.T) {
	h := rest.NewHandler(&stubRepo{})

	req := httptest.NewRequest(http.MethodPost, "/search", nil)
	rec := httptest.NewRecorder()
	h.ServeSearch(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}
