package domain

import (
	"time"
)

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

// UpdateName updates the player's name
func (p *Player) UpdateName(name string) {
	p.Name = name
	p.UpdatedAt = time.Now()
}