package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	contractevents "xffl/contracts/events"
	"xffl/services/search/internal/domain"
)

// Handlers processes incoming events and indexes them as search documents.
type Handlers struct {
	index *IndexDocument
}

func NewHandlers(index *IndexDocument) *Handlers {
	return &Handlers{index: index}
}

// HandlePlayerMatchUpdated handles AFL.PlayerMatchUpdated events.
func (h *Handlers) HandlePlayerMatchUpdated(ctx context.Context, payload []byte) error {
	var p contractevents.PlayerMatchUpdatedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("HandlePlayerMatchUpdated: unmarshal: %w", err)
	}

	slog.DebugContext(ctx, "event received",
		slog.String("event_type", contractevents.PlayerMatchUpdated),
		slog.Int("player_match_id", p.PlayerMatchID),
	)

	doc := domain.SearchDocument{
		ID:     fmt.Sprintf("afl_player_match_%d", p.PlayerMatchID),
		Source: domain.SourceAFL,
		Type:   domain.TypePlayerMatch,
		Data: map[string]any{
			"player_match_id":  p.PlayerMatchID,
			"player_season_id": p.PlayerSeasonID,
			"club_match_id":    p.ClubMatchID,
			"round_id":         p.RoundID,
			"kicks":            p.Kicks,
			"handballs":        p.Handballs,
			"marks":            p.Marks,
			"hitouts":          p.Hitouts,
			"tackles":          p.Tackles,
			"goals":            p.Goals,
			"behinds":          p.Behinds,
		},
	}

	return h.index.Execute(ctx, doc)
}

// HandleFantasyScoreCalculated handles FFL.FantasyScoreCalculated events.
func (h *Handlers) HandleFantasyScoreCalculated(ctx context.Context, payload []byte) error {
	var p contractevents.FantasyScoreCalculatedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("HandleFantasyScoreCalculated: unmarshal: %w", err)
	}

	slog.DebugContext(ctx, "event received",
		slog.String("event_type", contractevents.FantasyScoreCalculated),
		slog.Int("player_match_id", p.PlayerMatchID),
		slog.Int("score", p.Score),
	)

	doc := domain.SearchDocument{
		ID:     fmt.Sprintf("ffl_fantasy_score_%d", p.PlayerMatchID),
		Source: domain.SourceFFL,
		Type:   domain.TypeFantasyScore,
		Data: map[string]any{
			"player_match_id": p.PlayerMatchID,
			"score":           p.Score,
		},
	}

	return h.index.Execute(ctx, doc)
}
