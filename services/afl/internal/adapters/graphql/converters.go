package graphql

import (
	"strconv"
	"xffl/services/afl/internal/adapters/graphql/model"
	"xffl/services/afl/internal/domain/afl"
)

// ConvertClubToGraphQL converts a domain Club to a GraphQL Club
func ConvertClubToGraphQL(club afl.Club) *model.Club {
	return &model.Club{
		ID:           strconv.FormatUint(uint64(club.ID), 10),
		Name:         club.Name,
		Abbreviation: club.Abbreviation,
		CreatedAt:    club.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    club.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ConvertClubsToGraphQL converts a slice of domain Clubs to GraphQL Clubs
func ConvertClubsToGraphQL(clubs []afl.Club) []*model.Club {
	var result []*model.Club
	for _, club := range clubs {
		result = append(result, ConvertClubToGraphQL(club))
	}
	return result
}