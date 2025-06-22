package graphql

import (
	"xffl/services/afl/internal/services"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	clubService        *services.ClubService
	playerMatchService *services.PlayerMatchService
}

// NewResolver creates a new GraphQL resolver with injected dependencies
func NewResolver(clubService *services.ClubService, playerMatchService *services.PlayerMatchService) *Resolver {
	return &Resolver{
		clubService:        clubService,
		playerMatchService: playerMatchService,
	}
}
