// Package clock provides a Clock interface for time-dependent domain logic.
// Use RealClock in production and FixedClock in tests.
package clock

import "time"

// Clock returns the current time.
type Clock interface {
	Now() time.Time
}

// RealClock implements Clock using time.Now().
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

// FixedClock implements Clock with a fixed time, for use in tests.
type FixedClock struct {
	T time.Time
}

func (c FixedClock) Now() time.Time { return c.T }
