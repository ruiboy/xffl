package events

import (
	"context"
	"encoding/json"
	"fmt"

	contractevents "xffl/contracts/events"
	"xffl/services/afl/internal/application"
)

// Handlers translates incoming integration event payloads into application use case calls.
type Handlers struct {
	scoreCommands *application.ScoreCommands
}

func NewHandlers(scoreCommands *application.ScoreCommands) *Handlers {
	return &Handlers{scoreCommands: scoreCommands}
}

func (h *Handlers) HandleAflMatchFinalized(ctx context.Context, payload []byte) error {
	var p contractevents.AflMatchFinalizedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal AflMatchFinalized: %w", err)
	}
	return h.scoreCommands.ProcessAFLMatchFinalized(ctx, p.MatchID, p.SeasonID)
}
