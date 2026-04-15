package application_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/search/internal/application"
	"xffl/services/search/internal/domain"
)

// stubRepo captures the last document passed to Index.
type stubRepo struct {
	indexed []domain.SearchDocument
}

func (s *stubRepo) Index(_ context.Context, doc domain.SearchDocument) error {
	s.indexed = append(s.indexed, doc)
	return nil
}

func (s *stubRepo) Search(_ context.Context, _ domain.SearchQuery) (domain.SearchResult, error) {
	return domain.SearchResult{}, nil
}

func TestHandlePlayerMatchUpdated(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		wantDoc domain.SearchDocument
		wantErr bool
	}{
		{
			name: "valid payload",
			payload: mustMarshal(t, map[string]any{
				"player_match_id":  42,
				"player_season_id": 7,
				"club_match_id":    3,
				"round_id":         1,
				"kicks":            10,
				"handballs":        5,
				"marks":            4,
				"hitouts":          0,
				"tackles":          3,
				"goals":            2,
				"behinds":          1,
			}),
			wantDoc: domain.SearchDocument{
				ID:     "afl_player_match_42",
				Source: domain.SourceAFL,
				Type:   domain.TypePlayerMatch,
				Data: map[string]any{
					"player_match_id":  42,
					"player_season_id": 7,
					"club_match_id":    3,
					"round_id":         1,
					"kicks":            10,
					"handballs":        5,
					"marks":            4,
					"hitouts":          0,
					"tackles":          3,
					"goals":            2,
					"behinds":          1,
				},
			},
		},
		{
			name: "zero values",
			payload: mustMarshal(t, map[string]any{
				"player_match_id": 1,
			}),
			wantDoc: domain.SearchDocument{
				ID:     "afl_player_match_1",
				Source: domain.SourceAFL,
				Type:   domain.TypePlayerMatch,
				Data: map[string]any{
					"player_match_id":  1,
					"player_season_id": 0,
					"club_match_id":    0,
					"round_id":         0,
					"kicks":            0,
					"handballs":        0,
					"marks":            0,
					"hitouts":          0,
					"tackles":          0,
					"goals":            0,
					"behinds":          0,
				},
			},
		},
		{
			name:    "malformed JSON",
			payload: []byte(`{bad`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubRepo{}
			h := application.NewHandlers(application.NewIndexDocument(repo))

			err := h.HandlePlayerMatchUpdated(context.Background(), tt.payload)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Len(t, repo.indexed, 1)
			assert.Equal(t, tt.wantDoc.ID, repo.indexed[0].ID)
			assert.Equal(t, tt.wantDoc.Source, repo.indexed[0].Source)
			assert.Equal(t, tt.wantDoc.Type, repo.indexed[0].Type)
			// compare individual Data fields to avoid JSON number type mismatches
			for k, want := range tt.wantDoc.Data {
				assert.EqualValues(t, want, repo.indexed[0].Data[k], "Data[%q]", k)
			}
		})
	}
}

func TestHandleFantasyScoreCalculated(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		wantDoc domain.SearchDocument
		wantErr bool
	}{
		{
			name: "valid payload",
			payload: mustMarshal(t, map[string]any{
				"player_match_id": 99,
				"score":           120,
			}),
			wantDoc: domain.SearchDocument{
				ID:     "ffl_fantasy_score_99",
				Source: domain.SourceFFL,
				Type:   domain.TypeFantasyScore,
				Data: map[string]any{
					"player_match_id": 99,
					"score":           120,
				},
			},
		},
		{
			name: "zero values",
			payload: mustMarshal(t, map[string]any{
				"player_match_id": 5,
			}),
			wantDoc: domain.SearchDocument{
				ID:     "ffl_fantasy_score_5",
				Source: domain.SourceFFL,
				Type:   domain.TypeFantasyScore,
				Data: map[string]any{
					"player_match_id": 5,
					"score":           0,
				},
			},
		},
		{
			name:    "malformed JSON",
			payload: []byte(`not json`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubRepo{}
			h := application.NewHandlers(application.NewIndexDocument(repo))

			err := h.HandleFantasyScoreCalculated(context.Background(), tt.payload)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Len(t, repo.indexed, 1)
			assert.Equal(t, tt.wantDoc.ID, repo.indexed[0].ID)
			assert.Equal(t, tt.wantDoc.Source, repo.indexed[0].Source)
			assert.Equal(t, tt.wantDoc.Type, repo.indexed[0].Type)
			for k, want := range tt.wantDoc.Data {
				assert.EqualValues(t, want, repo.indexed[0].Data[k], "Data[%q]", k)
			}
		})
	}
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}
