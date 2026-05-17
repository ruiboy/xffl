package events

import (
	"context"
	"encoding/json"
	"fmt"

	contractevents "xffl/contracts/events"
	"xffl/services/ffl/internal/application"
)

// Handlers translates incoming integration event payloads into application use case calls.
type Handlers struct {
	commands *application.Commands
}

func NewHandlers(commands *application.Commands) *Handlers {
	return &Handlers{commands: commands}
}

func (h *Handlers) HandlePlayerMatchUpdated(ctx context.Context, payload []byte) error {
	var p contractevents.PlayerMatchUpdatedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal PlayerMatchUpdated: %w", err)
	}
	return h.commands.ProcessPlayerMatchUpdated(ctx, application.PlayerMatchUpdate{
		AFLPlayerMatchID:  p.PlayerMatchID,
		AFLPlayerSeasonID: p.PlayerSeasonID,
		ClubMatchID:       p.ClubMatchID,
		RoundID:           p.RoundID,
		Status:            p.Status,
		Goals:             p.Goals,
		Kicks:             p.Kicks,
		Handballs:         p.Handballs,
		Marks:             p.Marks,
		Tackles:           p.Tackles,
		Hitouts:           p.Hitouts,
	})
}

func (h *Handlers) HandleAflMatchFinalized(ctx context.Context, payload []byte) error {
	var p contractevents.AflMatchFinalizedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal AflMatchFinalized: %w", err)
	}
	return h.commands.ProcessAFLMatchFinalized(ctx, p.RoundID)
}

func (h *Handlers) HandleFflTeamFinalized(ctx context.Context, payload []byte) error {
	var p contractevents.FflTeamFinalizedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal FflTeamFinalized: %w", err)
	}
	return h.commands.ProcessFflTeamFinalized(ctx, p.ClubMatchID, p.MatchID)
}

func (h *Handlers) HandleFflClubMatchScoreFinalized(ctx context.Context, payload []byte) error {
	var p contractevents.FflClubMatchScoreFinalizedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal FflClubMatchScoreFinalized: %w", err)
	}
	return h.commands.ProcessFflClubMatchScoreFinalized(ctx, p.ClubMatchID, p.MatchID)
}

func (h *Handlers) HandleFflMatchFinalized(ctx context.Context, payload []byte) error {
	var p contractevents.FflMatchFinalizedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal FflMatchFinalized: %w", err)
	}
	return h.commands.ProcessFflMatchFinalized(ctx, p.MatchID, p.RoundID)
}
