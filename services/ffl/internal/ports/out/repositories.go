package out

import (
	"xffl/internal/domain/ffl"
)

// ClubRepository defines the interface for club data operations
type ClubRepository interface {
	FindAll() ([]ffl.Club, error)
	FindByID(id uint) (*ffl.Club, error)
	Create(club *ffl.Club) error
	Update(club *ffl.Club) error
	Delete(id uint) error
}

// PlayerRepository defines the interface for player data operations
type PlayerRepository interface {
	FindAll() ([]ffl.Player, error)
	FindByID(id uint) (*ffl.Player, error)
	FindByClubID(clubID uint) ([]ffl.Player, error)
	Create(player *ffl.Player) (*ffl.Player, error)
	Update(player *ffl.Player) (*ffl.Player, error)
	Delete(id uint) error
}

// ClubSeasonRepository defines the interface for club season data operations
type ClubSeasonRepository interface {
	FindBySeasonID(seasonID uint) ([]ffl.ClubSeason, error)
	FindByID(id uint) (*ffl.ClubSeason, error)
	Create(clubSeason *ffl.ClubSeason) error
	Update(clubSeason *ffl.ClubSeason) error
	Delete(id uint) error
}
