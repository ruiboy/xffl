package in

import (
	"gffl/internal/domain/ffl"
)

// ClubUseCase defines the interface for club business operations
type ClubUseCase interface {
	GetAllClubs() ([]ffl.Club, error)
	GetClubByID(id uint) (*ffl.Club, error)
	CreateClub(name string) (*ffl.Club, error)
	UpdateClub(id uint, name string) (*ffl.Club, error)
	DeleteClub(id uint) error
}

// PlayerUseCase defines the interface for player business operations
type PlayerUseCase interface {
	GetAllPlayers() ([]ffl.Player, error)
	GetPlayerByID(id uint) (*ffl.Player, error)
	GetPlayersByClubID(clubID uint) ([]ffl.Player, error)
	CreatePlayer(name string, clubID uint) (*ffl.Player, error)
	UpdatePlayer(id uint, name string) (*ffl.Player, error)
	DeletePlayer(id uint) error
}

// CreatePlayerInput represents the input for creating a player
type CreatePlayerInput struct {
	Name   string
	ClubID uint
}

// UpdatePlayerInput represents the input for updating a player
type UpdatePlayerInput struct {
	ID   uint
	Name string
}
