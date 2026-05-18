package events

import (
	"context"
	"encoding/json"
	"fmt"

	contractevents "xffl/contracts/events"
	"xffl/services/ffl/internal/application"
	"xffl/services/ffl/internal/domain"
)

// Handlers translates incoming integration event payloads into application use case calls.
type Handlers struct {
	commands *application.Commands
}

func NewHandlers(commands *application.Commands) *Handlers {
	return &Handlers{commands: commands}
}

func (h *Handlers) HandleAflPlayerMatchUpdated(ctx context.Context, payload []byte) error {
	var p contractevents.AflPlayerMatchUpdatedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal AflPlayerMatchUpdated: %w", err)
	}
	return h.commands.ProcessPlayerMatchUpdated(ctx, application.PlayerMatchUpdate{
		AFLPlayerMatchID:  p.PlayerMatchID,
		AFLPlayerSeasonID: p.PlayerSeasonID,
		ClubMatchID:       p.ClubMatchID,
		RoundID:           p.RoundID,
		Goals:             p.Goals,
		Kicks:             p.Kicks,
		Handballs:         p.Handballs,
		Marks:             p.Marks,
		Tackles:           p.Tackles,
		Hitouts:           p.Hitouts,
	})
}

func (h *Handlers) HandleAflMatchUpdated(ctx context.Context, payload []byte) error {
	var p contractevents.AflMatchUpdatedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal AflMatchUpdated: %w", err)
	}
	return h.commands.ProcessAFLMatchUpdated(ctx, p)
}

func (h *Handlers) HandleFflClubMatchUpdated(ctx context.Context, payload []byte) error {
	var p contractevents.FflClubMatchUpdatedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal FflClubMatchUpdated: %w", err)
	}
	return h.commands.ProcessFflClubMatchUpdated(ctx, p.ClubMatchID, p.MatchID, domain.ClubMatchDataStatus(p.DataStatus))
}

func (h *Handlers) HandleFflClubMatchScoreFinalized(ctx context.Context, payload []byte) error {
	var p contractevents.FflClubMatchScoreFinalizedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal FflClubMatchScoreFinalized: %w", err)
	}
	return h.commands.ProcessFflClubMatchScoreFinalized(ctx, p.ClubMatchID, p.MatchID)
}

func (h *Handlers) HandleFflMatchScoreFinalized(ctx context.Context, payload []byte) error {
	var p contractevents.FflMatchScoreFinalizedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal FflMatchScoreFinalized: %w", err)
	}
	return h.commands.ProcessFflMatchScoreFinalized(ctx, p.MatchID, p.RoundID)
}
