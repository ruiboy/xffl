package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"xffl/services/afl/internal/domain"
)

type ClubMatchRepository struct {
	pool *pgxpool.Pool
}

func NewClubMatchRepository(pool *pgxpool.Pool) *ClubMatchRepository {
	return &ClubMatchRepository{pool: pool}
}

func (r *ClubMatchRepository) FindByMatchID(ctx context.Context, matchID int) ([]domain.ClubMatch, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, match_id, club_season_id, rushed_behinds, drv_score
		 FROM afl.club_match WHERE match_id = $1 AND deleted_at IS NULL`,
		matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clubMatches []domain.ClubMatch
	for rows.Next() {
		var cm domain.ClubMatch
		if err := rows.Scan(&cm.ID, &cm.MatchID, &cm.ClubSeasonID, &cm.RushedBehinds, &cm.Score); err != nil {
			return nil, err
		}
		clubMatches = append(clubMatches, cm)
	}
	return clubMatches, rows.Err()
}

func (r *ClubMatchRepository) FindByID(ctx context.Context, id int) (domain.ClubMatch, error) {
	var cm domain.ClubMatch
	err := r.pool.QueryRow(ctx,
		`SELECT id, match_id, club_season_id, rushed_behinds, drv_score
		 FROM afl.club_match WHERE id = $1 AND deleted_at IS NULL`, id).
		Scan(&cm.ID, &cm.MatchID, &cm.ClubSeasonID, &cm.RushedBehinds, &cm.Score)
	return cm, err
}
