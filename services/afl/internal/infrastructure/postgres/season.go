package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"xffl/services/afl/internal/domain"
)

type SeasonRepository struct {
	pool *pgxpool.Pool
}

func NewSeasonRepository(pool *pgxpool.Pool) *SeasonRepository {
	return &SeasonRepository{pool: pool}
}

func (r *SeasonRepository) FindAll(ctx context.Context) ([]domain.Season, error) {
	rows, err := r.pool.Query(ctx,
		"SELECT id, name, league_id FROM afl.season WHERE deleted_at IS NULL ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seasons []domain.Season
	for rows.Next() {
		var s domain.Season
		if err := rows.Scan(&s.ID, &s.Name, &s.LeagueID); err != nil {
			return nil, err
		}
		seasons = append(seasons, s)
	}
	return seasons, rows.Err()
}

func (r *SeasonRepository) FindByID(ctx context.Context, id int) (domain.Season, error) {
	var s domain.Season
	err := r.pool.QueryRow(ctx,
		"SELECT id, name, league_id FROM afl.season WHERE id = $1 AND deleted_at IS NULL", id).
		Scan(&s.ID, &s.Name, &s.LeagueID)
	return s, err
}
