package graphql

import "xffl/services/ffl/internal/application"

// Resolver is the dependency injection container for GraphQL resolvers.
type Resolver struct {
	Queries     *application.Queries
	Commands    *application.Commands
	DataOps     *application.DataOpsCommands
}
