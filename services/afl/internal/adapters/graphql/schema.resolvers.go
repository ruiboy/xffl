package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.74

import (
	"context"
	"fmt"
	"strconv"
	"xffl/services/afl/internal/adapters/graphql/model"
	"xffl/services/afl/internal/domain"
)

// UpdateAFLPlayerMatch is the resolver for the updateAFLPlayerMatch field.
func (r *mutationResolver) UpdateAFLPlayerMatch(ctx context.Context, input model.UpdateAFLPlayerMatchInput) (*model.AFLPlayerMatch, error) {
	// Convert input IDs to uint
	playerSeasonID, err := strconv.ParseUint(input.PlayerSeasonID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid player season ID: %w", err)
	}
	
	clubMatchID, err := strconv.ParseUint(input.ClubMatchID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid club match ID: %w", err)
	}

	// Build stats from input (only update provided fields)
	stats := domain.PlayerMatch{
		PlayerSeasonID: uint(playerSeasonID),
		ClubMatchID:    uint(clubMatchID),
	}
	
	if input.Kicks != nil {
		stats.Kicks = int(*input.Kicks)
	}
	if input.Handballs != nil {
		stats.Handballs = int(*input.Handballs)
	}
	if input.Marks != nil {
		stats.Marks = int(*input.Marks)
	}
	if input.Hitouts != nil {
		stats.Hitouts = int(*input.Hitouts)
	}
	if input.Tackles != nil {
		stats.Tackles = int(*input.Tackles)
	}
	if input.Goals != nil {
		stats.Goals = int(*input.Goals)
	}
	if input.Behinds != nil {
		stats.Behinds = int(*input.Behinds)
	}

	// Update the player match
	updatedPlayerMatch, err := r.playerMatchService.UpdatePlayerMatch(uint(playerSeasonID), uint(clubMatchID), stats)
	if err != nil {
		return nil, fmt.Errorf("failed to update player match: %w", err)
	}

	return ConvertPlayerMatchToGraphQL(*updatedPlayerMatch), nil
}

// AflClubs is the resolver for the aflClubs field.
func (r *queryResolver) AflClubs(ctx context.Context) ([]*model.AFLClub, error) {
	clubs, err := r.clubService.GetAllClubs()
	if err != nil {
		return nil, err
	}
	return ConvertClubsToGraphQL(clubs), nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
