package in

import (
	"xffl/services/afl/internal/domain/afl"
)

// ClubUseCase defines the interface for club business operations
type ClubUseCase interface {
	GetAllClubs() ([]afl.Club, error)
}