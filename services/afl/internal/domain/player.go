package domain

import "context"

type Player struct {
	ID   int
	Name string
}

type PlayerRepository interface {
	FindByID(ctx context.Context, id int) (Player, error)
	FindByIDs(ctx context.Context, ids []int) ([]Player, error)
	Search(ctx context.Context, query string) ([]Player, error)
}
