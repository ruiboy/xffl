package in

import (
	"gffl/internal/domain"
)

// ClubUseCase defines the interface for club business operations
type ClubUseCase interface {
	GetAllClubs() ([]domain.Club, error)
	GetClubByID(id uint) (*domain.Club, error)
	CreateClub(name string) (*domain.Club, error)
	UpdateClub(id uint, name string) (*domain.Club, error)
	DeleteClub(id uint) error
}

// PlayerUseCase defines the interface for player business operations
type PlayerUseCase interface {
	GetAllPlayers() ([]domain.Player, error)
	GetPlayerByID(id uint) (*domain.Player, error)
	GetPlayersByClubID(clubID uint) ([]domain.Player, error)
	CreatePlayer(name string, clubID uint) (*domain.Player, error)
	UpdatePlayer(id uint, name string) (*domain.Player, error)
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