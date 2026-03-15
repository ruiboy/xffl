package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"xffl/services/afl/internal/domain"
)

type ClubRepository struct {
	pool *pgxpool.Pool
}

func NewClubRepository(pool *pgxpool.Pool) *ClubRepository {
	return &ClubRepository{pool: pool}
}

func (r *ClubRepository) FindAll(ctx context.Context) ([]domain.Club, error) {
	rows, err := r.pool.Query(ctx,
		"SELECT id, name FROM afl.club WHERE deleted_at IS NULL ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clubs []domain.Club
	for rows.Next() {
		var c domain.Club
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		clubs = append(clubs, c)
	}
	return clubs, rows.Err()
}

func (r *ClubRepository) FindByID(ctx context.Context, id int) (domain.Club, error) {
	var c domain.Club
	err := r.pool.QueryRow(ctx,
		"SELECT id, name FROM afl.club WHERE id = $1 AND deleted_at IS NULL", id).
		Scan(&c.ID, &c.Name)
	return c, err
}
