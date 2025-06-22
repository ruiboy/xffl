package services

import (
	"context"
	"xffl/pkg/events"
	"xffl/services/afl/internal/domain"
	domainEvents "xffl/services/afl/internal/domain/events"
)

// playerMatchRepository defines the interface for player match data operations needed by PlayerMatchService
type playerMatchRepository interface {
	UpdatePlayerMatch(playerSeasonID, clubMatchID uint, stats domain.PlayerMatch) (*domain.PlayerMatch, error)
	FindByPlayerSeasonAndClubMatch(playerSeasonID, clubMatchID uint) (*domain.PlayerMatch, error)
}

// PlayerMatchService implements player match business logic
type PlayerMatchService struct {
	playerMatchRepo playerMatchRepository
	eventDispatcher events.EventDispatcher
}

// NewPlayerMatchService creates a new player match service
func NewPlayerMatchService(playerMatchRepo playerMatchRepository, eventDispatcher events.EventDispatcher) *PlayerMatchService {
	return &PlayerMatchService{
		playerMatchRepo: playerMatchRepo,
		eventDispatcher: eventDispatcher,
	}
}

// UpdatePlayerMatch updates or creates a player match record
func (s *PlayerMatchService) UpdatePlayerMatch(playerSeasonID, clubMatchID uint, stats domain.PlayerMatch) (*domain.PlayerMatch, error) {
	// Get existing stats for comparison
	oldPlayerMatch, _ := s.playerMatchRepo.FindByPlayerSeasonAndClubMatch(playerSeasonID, clubMatchID)
	
	// Update the player match
	updatedPlayerMatch, err := s.playerMatchRepo.UpdatePlayerMatch(playerSeasonID, clubMatchID, stats)
	if err != nil {
		return nil, err
	}
	
	// Convert to event stats format
	oldStats := domainEvents.PlayerStats{}
	if oldPlayerMatch != nil {
		oldStats = domainEvents.PlayerStats{
			Kicks:     oldPlayerMatch.Kicks,
			Handballs: oldPlayerMatch.Handballs,
			Marks:     oldPlayerMatch.Marks,
			Hitouts:   oldPlayerMatch.Hitouts,
			Tackles:   oldPlayerMatch.Tackles,
			Goals:     oldPlayerMatch.Goals,
			Behinds:   oldPlayerMatch.Behinds,
		}
	}
	
	newStats := domainEvents.PlayerStats{
		Kicks:     updatedPlayerMatch.Kicks,
		Handballs: updatedPlayerMatch.Handballs,
		Marks:     updatedPlayerMatch.Marks,
		Hitouts:   updatedPlayerMatch.Hitouts,
		Tackles:   updatedPlayerMatch.Tackles,
		Goals:     updatedPlayerMatch.Goals,
		Behinds:   updatedPlayerMatch.Behinds,
	}
	
	// Publish domain event
	event := domainEvents.NewPlayerMatchUpdatedEvent(playerSeasonID, clubMatchID, oldStats, newStats)
	if err := s.eventDispatcher.Publish(context.Background(), event); err != nil {
		// Log error but don't fail the operation
		// In production, you might want more sophisticated error handling
		// such as storing events for retry
	}
	
	return updatedPlayerMatch, nil
}