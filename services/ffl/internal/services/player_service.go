package services

import (
	"xffl/services/ffl/internal/domain/ffl"
)

// playerRepository defines the interface for player data operations needed by PlayerService
type playerRepository interface {
	FindAll() ([]ffl.Player, error)
	FindByID(id uint) (*ffl.Player, error)
	FindByClubID(clubID uint) ([]ffl.Player, error)
	Create(player *ffl.Player) (*ffl.Player, error)
	Update(player *ffl.Player) (*ffl.Player, error)
	Delete(id uint) error
}

// playerServiceClubRepository defines the club repository interface needed by PlayerService
type playerServiceClubRepository interface {
	FindByID(id uint) (*ffl.Club, error)
}

// PlayerService implements player business logic
type PlayerService struct {
	playerRepo playerRepository
	clubRepo   playerServiceClubRepository
}

// NewPlayerService creates a new PlayerService
func NewPlayerService(playerRepo playerRepository, clubRepo playerServiceClubRepository) *PlayerService {
	return &PlayerService{
		playerRepo: playerRepo,
		clubRepo:   clubRepo,
	}
}

// GetAllPlayers retrieves all players
func (s *PlayerService) GetAllPlayers() ([]ffl.Player, error) {
	return s.playerRepo.FindAll()
}

// GetPlayerByID retrieves a player by its ID
func (s *PlayerService) GetPlayerByID(id uint) (*ffl.Player, error) {
	return s.playerRepo.FindByID(id)
}

// GetPlayersByClubID retrieves all players for a specific club
func (s *PlayerService) GetPlayersByClubID(clubID uint) ([]ffl.Player, error) {
	return s.playerRepo.FindByClubID(clubID)
}

// CreatePlayer creates a new player
func (s *PlayerService) CreatePlayer(name string, clubID uint) (*ffl.Player, error) {
	// Verify the club exists
	_, err := s.clubRepo.FindByID(clubID)
	if err != nil {
		return nil, err
	}
	
	player := ffl.NewPlayer(name, clubID)
	createdPlayer, err := s.playerRepo.Create(player)
	if err != nil {
		return nil, err
	}
	
	return createdPlayer, nil
}

// UpdatePlayer updates an existing player
func (s *PlayerService) UpdatePlayer(id uint, name string) (*ffl.Player, error) {
	player, err := s.playerRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	
	player.UpdateName(name)
	updatedPlayer, err := s.playerRepo.Update(player)
	if err != nil {
		return nil, err
	}
	
	return updatedPlayer, nil
}

// DeletePlayer deletes a player by its ID
func (s *PlayerService) DeletePlayer(id uint) error {
	return s.playerRepo.Delete(id)
}
