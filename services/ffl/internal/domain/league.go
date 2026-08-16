package domain

import "context"

type League struct {
	ID   int
	Name string
}

type LeagueRepository interface {
	FindAll(ctx context.Context) ([]League, error)
	Create(ctx context.Context, name string) (League, error)
}
