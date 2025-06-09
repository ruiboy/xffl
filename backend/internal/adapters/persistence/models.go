package persistence

import (
	"time"
	
	"gffl/internal/domain/ffl"
	"gorm.io/gorm"
)

// FFLClub represents the database model for Club
type FFLClub struct {
	gorm.Model
	Name    string      `gorm:"uniqueIndex;not null"`
	Players []FFLPlayer `gorm:"foreignKey:ClubID"`
}

// TableName specifies the table name for FFLClub
func (*FFLClub) TableName() string {
	return "ffl.club"
}

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

// ToDomain converts FFLClub to ffl.Club
func (c *FFLClub) ToDomain() ffl.Club {
	players := make([]ffl.Player, len(c.Players))
	for i, p := range c.Players {
		players[i] = p.ToDomain()
	}
	
	var deletedAt *time.Time
	if c.DeletedAt.Valid {
		deletedAt = &c.DeletedAt.Time
	}
	
	return ffl.Club{
		ID:        c.ID,
		Name:      c.Name,
		Players:   players,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		DeletedAt: deletedAt,
	}
}

// FromDomain converts ffl.Club to FFLClub
func (c *FFLClub) FromDomain(club *ffl.Club) {
	c.ID = club.ID
	c.Name = club.Name
	c.CreatedAt = club.CreatedAt
	c.UpdatedAt = club.UpdatedAt
	if club.DeletedAt != nil {
		c.DeletedAt = gorm.DeletedAt{Time: *club.DeletedAt, Valid: true}
	}
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