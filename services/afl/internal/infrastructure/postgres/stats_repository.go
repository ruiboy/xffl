package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"xffl/services/afl/internal/domain"
)

type PlayerSeasonStatsRepository struct {
	pool *pgxpool.Pool
}

func NewPlayerSeasonStatsRepository(pool *pgxpool.Pool) *PlayerSeasonStatsRepository {
	return &PlayerSeasonStatsRepository{pool: pool}
}

// getSeasonStatsMean aggregates stats for each player_season across final matches.
// Both upToRoundId ($2) and lastN ($3) are optional (pass nil to omit).
// upToRoundId filters to matches whose start_dt is before the earliest match in that round,
// using start_dt ordering rather than round ID ordering so IDs remain opaque.
// lastN is applied after the round filter, keeping only the most recent N matches per player.
const getSeasonStatsMean = `
WITH qualified AS (
  SELECT pm.player_season_id,
         pm.goals, pm.kicks, pm.handballs, pm.marks, pm.tackles, pm.hitouts,
         ROW_NUMBER() OVER (PARTITION BY pm.player_season_id ORDER BY m.start_dt DESC) AS rn
  FROM afl.player_match pm
  JOIN afl.club_match cm ON cm.id = pm.club_match_id AND cm.deleted_at IS NULL
  JOIN afl.match m ON m.id = cm.match_id AND m.deleted_at IS NULL
  WHERE pm.player_season_id = ANY($1::int[])
    AND pm.deleted_at IS NULL
    AND m.data_status = 'final'
    AND ($2::int IS NULL OR m.start_dt < (
      SELECT MIN(m2.start_dt)
      FROM afl.match m2
      WHERE m2.round_id = $2::int AND m2.deleted_at IS NULL
    ))
)
SELECT
  player_season_id,
  COUNT(*)::int          AS games,
  AVG(goals)::float8     AS avg_goals,
  AVG(kicks)::float8     AS avg_kicks,
  AVG(handballs)::float8 AS avg_handballs,
  AVG(marks)::float8     AS avg_marks,
  AVG(tackles)::float8   AS avg_tackles,
  AVG(hitouts)::float8   AS avg_hitouts
FROM qualified
WHERE $3::int IS NULL OR rn <= $3::int
GROUP BY player_season_id
`

func (r *PlayerSeasonStatsRepository) GetSeasonStats(ctx context.Context, params domain.PlayerSeasonStatsParams) ([]domain.PlayerSeasonStats, error) {
	int32IDs := make([]int32, len(params.PlayerSeasonIDs))
	for i, id := range params.PlayerSeasonIDs {
		int32IDs[i] = int32(id)
	}
	rows, err := r.pool.Query(ctx, getSeasonStatsMean, int32IDs, params.UpToRoundID, params.LastN)
	if err != nil {
		return nil, fmt.Errorf("get season stats: %w", err)
	}
	defer rows.Close()

	var out []domain.PlayerSeasonStats
	for rows.Next() {
		var s domain.PlayerSeasonStats
		if err := rows.Scan(&s.PlayerSeasonID, &s.Games,
			&s.Goals, &s.Kicks, &s.Handballs, &s.Marks, &s.Tackles, &s.Hitouts,
		); err != nil {
			return nil, fmt.Errorf("scan season stats: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
