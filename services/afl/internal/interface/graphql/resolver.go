package graphql

import "xffl/services/afl/internal/application"

// Resolver is the dependency injection container for GraphQL resolvers.
type Resolver struct {
	Queries *application.Queries
}
