package domain

import "context"

type Player struct {
	ID   int
	Name string
}

type PlayerWithClub struct {
	ID       int
	Name     string
	ClubName string
}

type PlayerRepository interface {
	Create(ctx context.Context, name string) (Player, error)
	FindByID(ctx context.Context, id int) (Player, error)
	FindByIDs(ctx context.Context, ids []int) ([]Player, error)
	FindByIDsWithClub(ctx context.Context, ids []int) ([]PlayerWithClub, error)
	Search(ctx context.Context, query string) ([]Player, error)
}
