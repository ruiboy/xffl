package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"xffl/services/afl/internal/domain"
)

type MatchRepository struct {
	pool *pgxpool.Pool
}

func NewMatchRepository(pool *pgxpool.Pool) *MatchRepository {
	return &MatchRepository{pool: pool}
}

func (r *MatchRepository) FindByRoundID(ctx context.Context, roundID int) ([]domain.Match, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, round_id, COALESCE(home_club_match_id, 0), COALESCE(away_club_match_id, 0),
		        COALESCE(venue, ''), COALESCE(start_dt, $2), COALESCE(drv_result, '')
		 FROM afl.match WHERE round_id = $1 AND deleted_at IS NULL`,
		roundID, time.Time{})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []domain.Match
	for rows.Next() {
		var m domain.Match
		var result string
		if err := rows.Scan(&m.ID, &m.RoundID, &m.HomeClubMatchID, &m.AwayClubMatchID,
			&m.Venue, &m.StartTime, &result); err != nil {
			return nil, err
		}
		m.Result = domain.MatchResult(result)
		matches = append(matches, m)
	}
	return matches, rows.Err()
}

func (r *MatchRepository) FindByID(ctx context.Context, id int) (domain.Match, error) {
	var m domain.Match
	var result string
	err := r.pool.QueryRow(ctx,
		`SELECT id, round_id, COALESCE(home_club_match_id, 0), COALESCE(away_club_match_id, 0),
		        COALESCE(venue, ''), COALESCE(start_dt, $2), COALESCE(drv_result, '')
		 FROM afl.match WHERE id = $1 AND deleted_at IS NULL`,
		id, time.Time{}).
		Scan(&m.ID, &m.RoundID, &m.HomeClubMatchID, &m.AwayClubMatchID,
			&m.Venue, &m.StartTime, &result)
	m.Result = domain.MatchResult(result)
	return m, err
}
