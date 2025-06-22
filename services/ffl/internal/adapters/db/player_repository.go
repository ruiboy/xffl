package db

import (
	"time"
	"xffl/services/ffl/internal/domain/ffl"
	"gorm.io/gorm"
)

// FFLPlayer represents the database model for Player
type FFLPlayer struct {
	gorm.Model
	Name   string  `gorm:"not null"`
	ClubID uint    `gorm:"not null"`
	Club   FFLClub `gorm:"foreignKey:ClubID"`
}

// TableName specifies the table name for FFLPlayer
func (*FFLPlayer) TableName() string {
	return "ffl.player"
}

// ToDomain converts FFLPlayer to ffl.Player
func (p *FFLPlayer) ToDomain() ffl.Player {
	var deletedAt *time.Time
	if p.DeletedAt.Valid {
		deletedAt = &p.DeletedAt.Time
	}
	
	return ffl.Player{
		ID:        p.ID,
		Name:      p.Name,
		ClubID:    p.ClubID,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		DeletedAt: deletedAt,
	}
}

// FromDomain converts ffl.Player to FFLPlayer
func (p *FFLPlayer) FromDomain(player *ffl.Player) {
	p.ID = player.ID
	p.Name = player.Name
	p.ClubID = player.ClubID
	p.CreatedAt = player.CreatedAt
	p.UpdatedAt = player.UpdatedAt
	if player.DeletedAt != nil {
		p.DeletedAt = gorm.DeletedAt{Time: *player.DeletedAt, Valid: true}
	}
}

// PlayerRepository implements player database operations
type PlayerRepository struct {
	db *gorm.DB
}

// NewPlayerRepository creates a new PlayerRepository
func NewPlayerRepository(db *gorm.DB) *PlayerRepository {
	return &PlayerRepository{
		db: db,
	}
}

// FindAll retrieves all players from the database
func (r *PlayerRepository) FindAll() ([]ffl.Player, error) {
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
func (r *PlayerRepository) FindByID(id uint) (*ffl.Player, error) {
	var fflPlayer FFLPlayer
	err := r.db.Preload("Club").First(&fflPlayer, id).Error
	if err != nil {
		return nil, err
	}
	
	player := fflPlayer.ToDomain()
	return &player, nil
}

// FindByClubID retrieves all players for a specific club
func (r *PlayerRepository) FindByClubID(clubID uint) ([]ffl.Player, error) {
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
func (r *PlayerRepository) Create(player *ffl.Player) (*ffl.Player, error) {
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
func (r *PlayerRepository) Update(player *ffl.Player) (*ffl.Player, error) {
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
func (r *PlayerRepository) Delete(id uint) error {
	return r.db.Delete(&FFLPlayer{}, id).Error
}
