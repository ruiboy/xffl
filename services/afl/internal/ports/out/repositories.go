package out

import (
	"xffl/services/afl/internal/domain/afl"
)

// ClubRepository defines the interface for club data operations
type ClubRepository interface {
	FindAll() ([]afl.Club, error)
	FindByID(id uint) (*afl.Club, error)
}

// PlayerMatchRepository defines the interface for player match data operations
type PlayerMatchRepository interface {
	UpdatePlayerMatch(playerSeasonID, clubMatchID uint, stats afl.PlayerMatch) (*afl.PlayerMatch, error)
	FindByPlayerSeasonAndClubMatch(playerSeasonID, clubMatchID uint) (*afl.PlayerMatch, error)
}