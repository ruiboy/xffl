package application

import (
	"xffl/services/ffl/internal/domain/ffl"
	"xffl/services/ffl/internal/ports/out"
)

// PlayerService implements the PlayerUseCase interface
type PlayerService struct {
	playerRepo out.PlayerRepository
	clubRepo   out.ClubRepository
}

// NewPlayerService creates a new PlayerService
func NewPlayerService(playerRepo out.PlayerRepository, clubRepo out.ClubRepository) *PlayerService {
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
