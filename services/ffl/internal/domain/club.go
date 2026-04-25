package domain

import "context"

type Club struct {
	ID   int
	Name string
}

type ClubRepository interface {
	FindAll(ctx context.Context) ([]Club, error)
	FindByID(ctx context.Context, id int) (Club, error)
	FindByIDs(ctx context.Context, ids []int) (map[int]Club, error)
}
