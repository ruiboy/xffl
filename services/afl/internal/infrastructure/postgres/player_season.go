package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"xffl/services/afl/internal/domain"
)

type PlayerSeasonRepository struct {
	pool *pgxpool.Pool
}

func NewPlayerSeasonRepository(pool *pgxpool.Pool) *PlayerSeasonRepository {
	return &PlayerSeasonRepository{pool: pool}
}

func (r *PlayerSeasonRepository) FindByID(ctx context.Context, id int) (domain.PlayerSeason, error) {
	var ps domain.PlayerSeason
	err := r.pool.QueryRow(ctx,
		"SELECT id, player_id, club_season_id FROM afl.player_season WHERE id = $1 AND deleted_at IS NULL", id).
		Scan(&ps.ID, &ps.PlayerID, &ps.ClubSeasonID)
	return ps, err
}
