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
	commands *application.Commands
}

func NewHandlers(commands *application.Commands) *Handlers {
	return &Handlers{commands: commands}
}

func (h *Handlers) HandleAflMatchFinalized(ctx context.Context, payload []byte) error {
	var p contractevents.AflMatchFinalizedPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal AflMatchFinalized: %w", err)
	}
	return h.commands.ProcessAFLMatchFinalized(ctx, p.MatchID, p.SeasonID)
}
