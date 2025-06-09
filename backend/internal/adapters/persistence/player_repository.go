package persistence

import (
	"gffl/internal/domain/ffl"
	"gorm.io/gorm"
)

// PlayerRepositoryImpl implements the PlayerRepository interface
type PlayerRepositoryImpl struct {
	db *gorm.DB
}

// NewPlayerRepository creates a new PlayerRepositoryImpl
func NewPlayerRepository(db *gorm.DB) *PlayerRepositoryImpl {
	return &PlayerRepositoryImpl{
		db: db,
	}
}

// FindAll retrieves all players from the database
func (r *PlayerRepositoryImpl) FindAll() ([]ffl.Player, error) {
	var fflPlayers []FFLPlayer
	err := r.db.Preload("Club").Find(&fflPlayers).Error
	if err != nil {
		return nil, err
	}
	
	players := make([]ffl.Player, len(fflPlayers))
	for i, fflPlayer := range fflPlayers {
		players[i] = fflPlayer.ToDomain()
	}
	
	return players, nil
}

// FindByID retrieves a player by its ID
func (r *PlayerRepositoryImpl) FindByID(id uint) (*ffl.Player, error) {
	var fflPlayer FFLPlayer
	err := r.db.Preload("Club").First(&fflPlayer, id).Error
	if err != nil {
		return nil, err
	}
	
	player := fflPlayer.ToDomain()
	return &player, nil
}

// FindByClubID retrieves all players for a specific club
func (r *PlayerRepositoryImpl) FindByClubID(clubID uint) ([]ffl.Player, error) {
	var fflPlayers []FFLPlayer
	err := r.db.Preload("Club").Where("club_id = ?", clubID).Find(&fflPlayers).Error
	if err != nil {
		return nil, err
	}
	
	players := make([]ffl.Player, len(fflPlayers))
	for i, fflPlayer := range fflPlayers {
		players[i] = fflPlayer.ToDomain()
	}
	
	return players, nil
}

// Create creates a new player in the database
func (r *PlayerRepositoryImpl) Create(player *ffl.Player) (*ffl.Player, error) {
	var fflPlayer FFLPlayer
	fflPlayer.FromDomain(player)
	
	err := r.db.Create(&fflPlayer).Error
	if err != nil {
		return nil, err
	}
	
	// Update the domain entity with the generated ID and timestamps
	player.ID = fflPlayer.ID
	player.CreatedAt = fflPlayer.CreatedAt
	player.UpdatedAt = fflPlayer.UpdatedAt
	
	return player, nil
}

// Update updates an existing player in the database
func (r *PlayerRepositoryImpl) Update(player *ffl.Player) (*ffl.Player, error) {
	var fflPlayer FFLPlayer
	fflPlayer.FromDomain(player)
	
	err := r.db.Save(&fflPlayer).Error
	if err != nil {
		return nil, err
	}
	
	// Update timestamps
	player.UpdatedAt = fflPlayer.UpdatedAt
	
	return player, nil
}

// Delete deletes a player by its ID
func (r *PlayerRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&FFLPlayer{}, id).Error
}