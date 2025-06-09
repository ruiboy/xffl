package ffl

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