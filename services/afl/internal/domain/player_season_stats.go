package domain

import "context"

// PlayerSeasonStats holds aggregated stats for a player_season over a qualifying set of matches.
type PlayerSeasonStats struct {
	PlayerSeasonID int
	Goals          float64
	Kicks          float64
	Handballs      float64
	Marks          float64
	Tackles        float64
	Hitouts        float64
	Games          int
}

// PlayerSeasonStatsParams controls which matches are included in the aggregation.
// UpToRoundID and LastN are both optional and may be combined.
type PlayerSeasonStatsParams struct {
	PlayerSeasonIDs []int
	UpToRoundID     *int // only final matches before the earliest start_dt in this round
	LastN           *int // restrict to the most recent N qualifying matches
}

// PlayerSeasonStatsRepository is a read-only repository for aggregated player stats queries.
// Kept separate from PlayerMatchRepository to avoid mixing write-path concerns.
type PlayerSeasonStatsRepository interface {
	GetSeasonStats(ctx context.Context, params PlayerSeasonStatsParams) ([]PlayerSeasonStats, error)
}
