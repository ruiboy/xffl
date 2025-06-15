package application

import (
	"xffl/services/afl/internal/domain/afl"
	"xffl/services/afl/internal/ports/out"
)

// PlayerMatchService implements player match business logic
type PlayerMatchService struct {
	playerMatchRepo out.PlayerMatchRepository
}

// NewPlayerMatchService creates a new player match service
func NewPlayerMatchService(playerMatchRepo out.PlayerMatchRepository) *PlayerMatchService {
	return &PlayerMatchService{
		playerMatchRepo: playerMatchRepo,
	}
}

// UpdatePlayerMatch updates or creates a player match record
func (s *PlayerMatchService) UpdatePlayerMatch(playerSeasonID, clubMatchID uint, stats afl.PlayerMatch) (*afl.PlayerMatch, error) {
	// Delegate to repository for upsert logic
	return s.playerMatchRepo.UpdatePlayerMatch(playerSeasonID, clubMatchID, stats)
}