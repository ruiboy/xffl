package application

import (
	"context"
	"time"

	"xffl/services/afl/internal/domain"
)

// aflTimezone is the canonical timezone for AFL round window boundaries.
const aflTimezone = "Australia/Adelaide"

var adelaideLoc = mustLoadLocation(aflTimezone)

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic("failed to load timezone " + name + ": " + err.Error())
	}
	return loc
}

// midnightBefore returns the start of the calendar day in Adelaide time containing t.
func midnightBefore(t time.Time) time.Time {
	local := t.In(adelaideLoc)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, adelaideLoc)
}

// LiveRound returns the contextually relevant AFL round for the current time,
// or nil when no round is live (e.g. pre-season before the first round day).
//
// The DB returns at most two neighbours: the most recently started round and
// the first upcoming round. The selection rule:
//   - No started round → nil (frontend shows a placeholder).
//   - If now >= midnightBefore(upcoming.firstMatchTime) → upcoming round is live.
//   - Otherwise → most recently started round is live.
func (q *Queries) LiveRound(ctx context.Context) (*domain.RoundWithStart, error) {
	now := q.clock.Now()
	neighbours, err := q.rounds.FindNeighbours(ctx, now)
	if err != nil {
		return nil, err
	}

	var before, after *domain.RoundWithStart
	for i := range neighbours {
		r := &neighbours[i]
		if !r.FirstMatchTime.After(now) {
			before = r
		} else {
			after = r
		}
	}

	if before == nil {
		return nil, nil
	}
	if after != nil && !now.Before(midnightBefore(after.FirstMatchTime)) {
		return after, nil
	}
	return before, nil
}
