package domain

import (
	"context"
	"fmt"
	"time"
)

type Round struct {
	ID       int
	Name     string
	SeasonID int
}

// RoundStatus indicates whether a round's match window is currently open.
type RoundStatus string

const (
	RoundStatusOpen   RoundStatus = "open"
	RoundStatusClosed RoundStatus = "closed"
)

// LiveRoundResult is the contextually relevant round for a given point in time.
// Status is Open when the round's match window contains that time, Closed when
// it is the most recently completed round and no window is currently open.
type LiveRoundResult struct {
	Round  Round
	Status RoundStatus
}

// RoundWithBounds pairs a round with the start times of its first and last matches,
// used by the LiveRound use case to compute window boundaries.
type RoundWithBounds struct {
	Round          Round
	FirstMatchTime time.Time
	LastMatchTime  time.Time
}

type RoundRepository interface {
	FindBySeasonID(ctx context.Context, seasonID int) ([]Round, error)
	FindByID(ctx context.Context, id int) (Round, error)
	FindLatest(ctx context.Context) (Round, error)
	FindWithMatchBoundsBySeasonID(ctx context.Context, seasonID int) ([]RoundWithBounds, error)
}

// adelaideLoc is the canonical timezone for AFL round window boundaries.
var adelaideLoc = mustLoadLocation("Australia/Adelaide")

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic("failed to load timezone " + name + ": " + err.Error())
	}
	return loc
}

// midnightBefore returns the start of the calendar day (Adelaide time) containing t.
func midnightBefore(t time.Time) time.Time {
	local := t.In(adelaideLoc)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, adelaideLoc)
}

// midnightAfter returns the start of the calendar day after t (Adelaide time).
func midnightAfter(t time.Time) time.Time {
	return midnightBefore(t).AddDate(0, 0, 1)
}

// FindLiveRound returns the contextually relevant round for time t.
// It returns Open if t falls within a round's match window, or Closed for the
// most recently completed round when no window is active.
func FindLiveRound(rounds []RoundWithBounds, t time.Time) (LiveRoundResult, error) {
	var mostRecent *RoundWithBounds
	for i := range rounds {
		r := &rounds[i]
		open := midnightBefore(r.FirstMatchTime)
		close := midnightAfter(r.LastMatchTime)
		if !t.Before(open) && t.Before(close) {
			return LiveRoundResult{Round: r.Round, Status: RoundStatusOpen}, nil
		}
		if t.After(close) {
			mostRecent = r
		}
	}
	if mostRecent != nil {
		return LiveRoundResult{Round: mostRecent.Round, Status: RoundStatusClosed}, nil
	}
	// All rounds are in the future — return the first as closed.
	if len(rounds) > 0 {
		return LiveRoundResult{Round: rounds[0].Round, Status: RoundStatusClosed}, nil
	}
	return LiveRoundResult{}, fmt.Errorf("no rounds found")
}
