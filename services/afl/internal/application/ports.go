package application

import "context"

// PlayerMatchStats represents a single player's stats for one match as
// provided by an external source. The adapter resolves external codes to
// domain-friendly names; the application layer resolves names to domain IDs.
type PlayerMatchStats struct {
	ExternalPlayerID string // source's own player identifier (for xref mapping)
	PlayerName       string
	ClubName         string // canonical afl.club.name, resolved by the adapter
	RoundName        string // e.g. "Round 1", "Opening Round"
	SeasonYear       int
	Kicks            int
	Handballs        int
	Marks            int
	Hitouts          int
	Tackles          int
	Goals            int
	Behinds          int
}

// StatsProvider is the outbound port for fetching player match statistics
// from an external source. Adapters in infrastructure implement this interface.
type StatsProvider interface {
	FetchSeasonStats(ctx context.Context, year int) ([]PlayerMatchStats, error)
}
