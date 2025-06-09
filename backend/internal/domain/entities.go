package domain

import (
	"time"
)

// Club represents a fantasy football club domain entity
type Club struct {
	ID        uint
	Name      string
	Players   []Player
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// Player represents a fantasy football player domain entity  
type Player struct {
	ID        uint
	Name      string
	ClubID    uint
	Club      *Club
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// NewClub creates a new Club domain entity
func NewClub(name string) *Club {
	now := time.Now()
	return &Club{
		Name:      name,
		Players:   make([]Player, 0),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewPlayer creates a new Player domain entity
func NewPlayer(name string, clubID uint) *Player {
	now := time.Now()
	return &Player{
		Name:      name,
		ClubID:    clubID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddPlayer adds a player to a club
func (c *Club) AddPlayer(player *Player) {
	player.ClubID = c.ID
	player.Club = c
	c.Players = append(c.Players, *player)
}

// UpdateName updates the club's name
func (c *Club) UpdateName(name string) {
	c.Name = name
	c.UpdatedAt = time.Now()
}

// UpdateName updates the player's name
func (p *Player) UpdateName(name string) {
	p.Name = name
	p.UpdatedAt = time.Now()
}