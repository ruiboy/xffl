package graphql

import (
	"strconv"
	"time"

	"xffl/services/ffl/internal/adapters/graphql/model"
	"xffl/services/ffl/internal/domain"
)

// ClubToGraphQL converts a ffl Club to GraphQL FFLClub
func ClubToGraphQL(club *domain.Club) *model.FFLClub {
	players := make([]*model.FFLPlayer, len(club.Players))
	for i, player := range club.Players {
		players[i] = PlayerToGraphQL(&player)
	}

	var deletedAt *string
	if club.DeletedAt != nil {
		str := club.DeletedAt.Format(time.RFC3339)
		deletedAt = &str
	}

	return &model.FFLClub{
		ID:        strconv.FormatUint(uint64(club.ID), 10),
		Name:      club.Name,
		CreatedAt: club.CreatedAt.Format(time.RFC3339),
		UpdatedAt: club.UpdatedAt.Format(time.RFC3339),
		DeletedAt: deletedAt,
		Players:   players,
	}
}

// PlayerToGraphQL converts a ffl Player to GraphQL FFLPlayer
func PlayerToGraphQL(player *domain.Player) *model.FFLPlayer {
	var deletedAt *string
	if player.DeletedAt != nil {
		str := player.DeletedAt.Format(time.RFC3339)
		deletedAt = &str
	}

	return &model.FFLPlayer{
		ID:        strconv.FormatUint(uint64(player.ID), 10),
		Name:      player.Name,
		ClubID:    strconv.FormatUint(uint64(player.ClubID), 10),
		CreatedAt: player.CreatedAt.Format(time.RFC3339),
		UpdatedAt: player.UpdatedAt.Format(time.RFC3339),
		DeletedAt: deletedAt,
	}
}

// ClubsToGraphQL converts a slice of ffl Clubs to GraphQL FFLClubs
func ClubsToGraphQL(clubs []domain.Club) []*model.FFLClub {
	result := make([]*model.FFLClub, len(clubs))
	for i, club := range clubs {
		result[i] = ClubToGraphQL(&club)
	}
	return result
}

// PlayersToGraphQL converts a slice of ffl Players to GraphQL FFLPlayers
func PlayersToGraphQL(players []domain.Player) []*model.FFLPlayer {
	result := make([]*model.FFLPlayer, len(players))
	for i, player := range players {
		result[i] = PlayerToGraphQL(&player)
	}
	return result
}

// ClubSeasonToGraphQL converts a ffl ClubSeason to GraphQL FFLClubSeason
func ClubSeasonToGraphQL(clubSeason *domain.ClubSeason) *model.FFLClubSeason {
	var deletedAt *string
	if clubSeason.DeletedAt != nil {
		str := clubSeason.DeletedAt.Format(time.RFC3339)
		deletedAt = &str
	}

	return &model.FFLClubSeason{
		ID:                strconv.FormatUint(uint64(clubSeason.ID), 10),
		ClubID:            strconv.FormatUint(uint64(clubSeason.ClubID), 10),
		SeasonID:          strconv.FormatUint(uint64(clubSeason.SeasonID), 10),
		ClubName:          clubSeason.Club.Name,
		Played:            int32(clubSeason.Played),
		Won:               int32(clubSeason.Won),
		Lost:              int32(clubSeason.Lost),
		Drawn:             int32(clubSeason.Drawn),
		PointsFor:         int32(clubSeason.PointsFor),
		PointsAgainst:     int32(clubSeason.PointsAgainst),
		ExtraPoints:       int32(clubSeason.ExtraPoints),
		PremiershipPoints: int32(clubSeason.PremiershipPoints),
		Percentage:        clubSeason.Percentage(),
		CreatedAt:         clubSeason.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         clubSeason.UpdatedAt.Format(time.RFC3339),
		DeletedAt:         deletedAt,
	}
}

// ClubSeasonsToGraphQL converts a slice of ffl ClubSeasons to GraphQL FFLClubSeasons
func ClubSeasonsToGraphQL(clubSeasons []domain.ClubSeason) []*model.FFLClubSeason {
	result := make([]*model.FFLClubSeason, len(clubSeasons))
	for i, clubSeason := range clubSeasons {
		result[i] = ClubSeasonToGraphQL(&clubSeason)
	}
	return result
}

// ParseID converts a string ID to uint
func ParseID(id string) (uint, error) {
	parsed, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}
