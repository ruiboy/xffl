package domain

import "context"

type Player struct {
	ID   int
	Name string
}

type PlayerRepository interface {
	FindAll(ctx context.Context) ([]Player, error)
	FindByID(ctx context.Context, id int) (Player, error)
}
