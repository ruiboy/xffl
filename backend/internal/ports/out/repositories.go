package out

import (
	"gffl/internal/domain"
)

// ClubRepository defines the interface for club data operations
type ClubRepository interface {
	FindAll() ([]domain.Club, error)
	FindByID(id uint) (*domain.Club, error)
	Create(club *domain.Club) error
	Update(club *domain.Club) error
	Delete(id uint) error
}

// PlayerRepository defines the interface for player data operations
type PlayerRepository interface {
	FindAll() ([]domain.Player, error)
	FindByID(id uint) (*domain.Player, error)
	FindByClubID(clubID uint) ([]domain.Player, error)
	Create(player *domain.Player) (*domain.Player, error)
	Update(player *domain.Player) (*domain.Player, error)
	Delete(id uint) error
}