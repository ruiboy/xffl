package graphql

import (
	"xffl/services/ffl/internal/services"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver holds the services for GraphQL resolvers
type Resolver struct {
	clubService       *services.ClubService
	playerService     *services.PlayerService
	clubSeasonService *services.ClubSeasonService
}

// NewResolver creates a new GraphQL resolver
func NewResolver(clubService *services.ClubService, playerService *services.PlayerService, clubSeasonService *services.ClubSeasonService) *Resolver {
	return &Resolver{
		clubService:       clubService,
		playerService:     playerService,
		clubSeasonService: clubSeasonService,
	}
}
