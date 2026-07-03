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

// playerSeasonStatsSQL builds the full query for aggregating player-season stats.
//
// Parameters passed to the query:
//
//	$1 int[]  — player_season_id list
//	$2 int    — upToRoundID (NULL = no filter); restricts to final matches whose
//	            start_dt is before the earliest start_dt in the given round, so
//	            IDs remain opaque to callers
//	$3 int    — lastN (NULL = no filter); keeps only the most recent N matches
//	            per player after the round filter is applied
func playerSeasonStatsSQL(method domain.StatMethod) string {
	const cte = `
WITH ranked AS (
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
`
	const meanSelect = `
SELECT
  player_season_id,
  COUNT(*)::int          AS games,
  AVG(goals)::float8     AS goals,
  AVG(kicks)::float8     AS kicks,
  AVG(handballs)::float8 AS handballs,
  AVG(marks)::float8     AS marks,
  AVG(tackles)::float8   AS tackles,
  AVG(hitouts)::float8   AS hitouts
FROM ranked
WHERE $3::int IS NULL OR rn <= $3::int
GROUP BY player_season_id
`
	const medianSelect = `
SELECT
  player_season_id,
  COUNT(*)::int AS games,
  PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY goals)::float8     AS goals,
  PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY kicks)::float8     AS kicks,
  PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY handballs)::float8 AS handballs,
  PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY marks)::float8     AS marks,
  PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY tackles)::float8   AS tackles,
  PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY hitouts)::float8   AS hitouts
FROM ranked
WHERE $3::int IS NULL OR rn <= $3::int
GROUP BY player_season_id
`
	if method == domain.StatMethodMedian {
		return cte + medianSelect
	}
	return cte + meanSelect
}

func (r *PlayerSeasonStatsRepository) GetSeasonStats(ctx context.Context, params domain.PlayerSeasonStatsParams) ([]domain.PlayerSeasonStats, error) {
	int32IDs := make([]int32, len(params.PlayerSeasonIDs))
	for i, id := range params.PlayerSeasonIDs {
		int32IDs[i] = int32(id)
	}
	rows, err := r.pool.Query(ctx, playerSeasonStatsSQL(params.Method), int32IDs, params.UpToRoundID, params.LastN)
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
