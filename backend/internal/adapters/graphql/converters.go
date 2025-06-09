package graphql

import (
	"strconv"
	"time"

	"gffl/internal/adapters/graphql/model"
	"gffl/internal/domain/ffl"
)

// ClubToGraphQL converts a ffl Club to GraphQL FFLClub
func ClubToGraphQL(club *ffl.Club) *model.FFLClub {
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
func PlayerToGraphQL(player *ffl.Player) *model.FFLPlayer {
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
func ClubsToGraphQL(clubs []ffl.Club) []*model.FFLClub {
	result := make([]*model.FFLClub, len(clubs))
	for i, club := range clubs {
		result[i] = ClubToGraphQL(&club)
	}
	return result
}

// PlayersToGraphQL converts a slice of ffl Players to GraphQL FFLPlayers
func PlayersToGraphQL(players []ffl.Player) []*model.FFLPlayer {
	result := make([]*model.FFLPlayer, len(players))
	for i, player := range players {
		result[i] = PlayerToGraphQL(&player)
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