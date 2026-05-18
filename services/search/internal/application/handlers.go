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

// HandleAflPlayerMatchUpdated handles AFL.PlayerMatchUpdated events.
func (h *Handlers) HandleAflPlayerMatchUpdated(ctx context.Context, payload []byte) error {
	var p contractevents.AflPlayerMatchUpdatedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("HandleAflPlayerMatchUpdated: unmarshal: %w", err)
	}

	slog.DebugContext(ctx, "event received",
		slog.String("event_type", contractevents.AflPlayerMatchUpdated),
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

// HandleFflPlayerMatchUpdated handles FFL.PlayerMatchUpdated events.
func (h *Handlers) HandleFflPlayerMatchUpdated(ctx context.Context, payload []byte) error {
	var p contractevents.FflPlayerMatchUpdatedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("HandleFflPlayerMatchUpdated: unmarshal: %w", err)
	}

	slog.DebugContext(ctx, "event received",
		slog.String("event_type", contractevents.FflPlayerMatchUpdated),
		slog.Int("player_match_id", p.PlayerMatchID),
		slog.Int("score", p.Score),
	)

	doc := domain.SearchDocument{
		ID:     fmt.Sprintf("ffl_player_match_%d", p.PlayerMatchID),
		Source: domain.SourceFFL,
		Type:   domain.TypeFantasyScore,
		Data: map[string]any{
			"player_match_id": p.PlayerMatchID,
			"club_match_id":   p.ClubMatchID,
			"score":           p.Score,
		},
	}

	return h.index.Execute(ctx, doc)
}
