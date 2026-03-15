package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"xffl/services/afl/internal/domain"
)

type ClubSeasonRepository struct {
	pool *pgxpool.Pool
}

func NewClubSeasonRepository(pool *pgxpool.Pool) *ClubSeasonRepository {
	return &ClubSeasonRepository{pool: pool}
}

func (r *ClubSeasonRepository) FindBySeasonID(ctx context.Context, seasonID int) ([]domain.ClubSeason, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, club_id, season_id, drv_played, drv_won, drv_lost, drv_drawn,
		        drv_for, drv_against, drv_premiership_points
		 FROM afl.club_season WHERE season_id = $1 AND deleted_at IS NULL
		 ORDER BY drv_premiership_points DESC, (drv_for - drv_against) DESC`,
		seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clubSeasons []domain.ClubSeason
	for rows.Next() {
		var cs domain.ClubSeason
		if err := rows.Scan(&cs.ID, &cs.ClubID, &cs.SeasonID, &cs.Played, &cs.Won,
			&cs.Lost, &cs.Drawn, &cs.For, &cs.Against, &cs.PremiershipPoints); err != nil {
			return nil, err
		}
		clubSeasons = append(clubSeasons, cs)
	}
	return clubSeasons, rows.Err()
}

func (r *ClubSeasonRepository) FindByID(ctx context.Context, id int) (domain.ClubSeason, error) {
	var cs domain.ClubSeason
	err := r.pool.QueryRow(ctx,
		`SELECT id, club_id, season_id, drv_played, drv_won, drv_lost, drv_drawn,
		        drv_for, drv_against, drv_premiership_points
		 FROM afl.club_season WHERE id = $1 AND deleted_at IS NULL`, id).
		Scan(&cs.ID, &cs.ClubID, &cs.SeasonID, &cs.Played, &cs.Won,
			&cs.Lost, &cs.Drawn, &cs.For, &cs.Against, &cs.PremiershipPoints)
	return cs, err
}
