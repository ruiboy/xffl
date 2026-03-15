package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"xffl/services/afl/internal/domain"
)

type PlayerRepository struct {
	pool *pgxpool.Pool
}

func NewPlayerRepository(pool *pgxpool.Pool) *PlayerRepository {
	return &PlayerRepository{pool: pool}
}

func (r *PlayerRepository) FindByClubID(ctx context.Context, clubID int) ([]domain.Player, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT p.id, p.name, COALESCE(p.club_id, 0) FROM afl.player p
		 LEFT JOIN afl.player_season ps ON ps.player_id = p.id AND ps.deleted_at IS NULL
		 LEFT JOIN afl.club_season cs ON ps.club_season_id = cs.id AND cs.deleted_at IS NULL
		 WHERE (p.club_id = $1 OR cs.club_id = $1) AND p.deleted_at IS NULL
		 ORDER BY p.name`,
		clubID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []domain.Player
	for rows.Next() {
		var p domain.Player
		if err := rows.Scan(&p.ID, &p.Name, &p.ClubID); err != nil {
			return nil, err
		}
		players = append(players, p)
	}
	return players, rows.Err()
}

func (r *PlayerRepository) FindByID(ctx context.Context, id int) (domain.Player, error) {
	var p domain.Player
	err := r.pool.QueryRow(ctx,
		"SELECT id, name, COALESCE(club_id, 0) FROM afl.player WHERE id = $1 AND deleted_at IS NULL", id).
		Scan(&p.ID, &p.Name, &p.ClubID)
	return p, err
}
