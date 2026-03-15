package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"xffl/services/afl/internal/domain"
)

type PlayerMatchRepository struct {
	pool *pgxpool.Pool
}

func NewPlayerMatchRepository(pool *pgxpool.Pool) *PlayerMatchRepository {
	return &PlayerMatchRepository{pool: pool}
}

func (r *PlayerMatchRepository) FindByClubMatchID(ctx context.Context, clubMatchID int) ([]domain.PlayerMatch, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, club_match_id, player_season_id, kicks, handballs, marks, hitouts, tackles, goals, behinds
		 FROM afl.player_match WHERE club_match_id = $1 AND deleted_at IS NULL`,
		clubMatchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pms []domain.PlayerMatch
	for rows.Next() {
		var pm domain.PlayerMatch
		if err := rows.Scan(&pm.ID, &pm.ClubMatchID, &pm.PlayerSeasonID,
			&pm.Kicks, &pm.Handballs, &pm.Marks, &pm.Hitouts, &pm.Tackles, &pm.Goals, &pm.Behinds); err != nil {
			return nil, err
		}
		pms = append(pms, pm)
	}
	return pms, rows.Err()
}

func (r *PlayerMatchRepository) FindByID(ctx context.Context, id int) (domain.PlayerMatch, error) {
	var pm domain.PlayerMatch
	err := r.pool.QueryRow(ctx,
		`SELECT id, club_match_id, player_season_id, kicks, handballs, marks, hitouts, tackles, goals, behinds
		 FROM afl.player_match WHERE id = $1 AND deleted_at IS NULL`, id).
		Scan(&pm.ID, &pm.ClubMatchID, &pm.PlayerSeasonID,
			&pm.Kicks, &pm.Handballs, &pm.Marks, &pm.Hitouts, &pm.Tackles, &pm.Goals, &pm.Behinds)
	return pm, err
}
