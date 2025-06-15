package out

import (
	"xffl/services/afl/internal/domain/afl"
)

// ClubRepository defines the interface for club data operations
type ClubRepository interface {
	FindAll() ([]afl.Club, error)
	FindByID(id uint) (*afl.Club, error)
}