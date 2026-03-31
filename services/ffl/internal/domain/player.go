package domain

import "context"

type Player struct {
	ID          int
	Name        string
	AFLPlayerID *int
}

type PlayerRepository interface {
	FindAll(ctx context.Context) ([]Player, error)
	FindByID(ctx context.Context, id int) (Player, error)
	Create(ctx context.Context, name string) (Player, error)
	Update(ctx context.Context, id int, name string) (Player, error)
	Delete(ctx context.Context, id int) error
}
