package graphql

import (
	"strconv"
	"xffl/services/afl/internal/adapters/graphql/model"
	"xffl/services/afl/internal/domain/afl"
)

// ConvertClubToGraphQL converts a domain Club to a GraphQL Club
func ConvertClubToGraphQL(club afl.Club) *model.AFLClub {
	return &model.AFLClub{
		ID:        strconv.FormatUint(uint64(club.ID), 10),
		Name:      club.Name,
		CreatedAt: club.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: club.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ConvertClubsToGraphQL converts a slice of domain Clubs to GraphQL Clubs
func ConvertClubsToGraphQL(clubs []afl.Club) []*model.AFLClub {
	var result []*model.AFLClub
	for _, club := range clubs {
		result = append(result, ConvertClubToGraphQL(club))
	}
	return result
}

// ConvertPlayerMatchToGraphQL converts a domain PlayerMatch to a GraphQL PlayerMatch
func ConvertPlayerMatchToGraphQL(playerMatch afl.PlayerMatch) *model.AFLPlayerMatch {
	return &model.AFLPlayerMatch{
		ID:             strconv.FormatUint(uint64(playerMatch.ID), 10),
		PlayerSeasonID: strconv.FormatUint(uint64(playerMatch.PlayerSeasonID), 10),
		ClubMatchID:    strconv.FormatUint(uint64(playerMatch.ClubMatchID), 10),
		Kicks:          int32(playerMatch.Kicks),
		Handballs:      int32(playerMatch.Handballs),
		Marks:          int32(playerMatch.Marks),
		Hitouts:        int32(playerMatch.Hitouts),
		Tackles:        int32(playerMatch.Tackles),
		Goals:          int32(playerMatch.Goals),
		Behinds:        int32(playerMatch.Behinds),
		CreatedAt:      playerMatch.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      playerMatch.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}