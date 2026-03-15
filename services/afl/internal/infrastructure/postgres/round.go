package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"xffl/services/afl/internal/domain"
)

type RoundRepository struct {
	pool *pgxpool.Pool
}

func NewRoundRepository(pool *pgxpool.Pool) *RoundRepository {
	return &RoundRepository{pool: pool}
}

func (r *RoundRepository) FindBySeasonID(ctx context.Context, seasonID int) ([]domain.Round, error) {
	rows, err := r.pool.Query(ctx,
		"SELECT id, name, season_id FROM afl.round WHERE season_id = $1 AND deleted_at IS NULL ORDER BY name",
		seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rounds []domain.Round
	for rows.Next() {
		var rd domain.Round
		if err := rows.Scan(&rd.ID, &rd.Name, &rd.SeasonID); err != nil {
			return nil, err
		}
		rounds = append(rounds, rd)
	}
	return rounds, rows.Err()
}

func (r *RoundRepository) FindByID(ctx context.Context, id int) (domain.Round, error) {
	var rd domain.Round
	err := r.pool.QueryRow(ctx,
		"SELECT id, name, season_id FROM afl.round WHERE id = $1 AND deleted_at IS NULL", id).
		Scan(&rd.ID, &rd.Name, &rd.SeasonID)
	return rd, err
}
