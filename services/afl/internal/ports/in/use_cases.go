package in

import (
	"xffl/services/afl/internal/domain/afl"
)

// ClubUseCase defines the interface for club business operations
type ClubUseCase interface {
	GetAllClubs() ([]afl.Club, error)
}

// PlayerMatchUseCase defines the interface for player match business operations
type PlayerMatchUseCase interface {
	UpdatePlayerMatch(playerSeasonID, clubMatchID uint, stats afl.PlayerMatch) (*afl.PlayerMatch, error)
}