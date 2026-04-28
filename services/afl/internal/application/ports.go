package application

import "context"

// PlayerCandidate is a known AFL player available for fuzzy name matching.
type PlayerCandidate struct {
	PlayerSeasonID int
	Name           string
	Club           string
}

// PlayerMatch is the result of resolving one parsed name against a candidate pool.
type PlayerMatch struct {
	Candidate  PlayerCandidate
	Confidence float64 // 0.0–1.0
}

// PlayerResolver fuzzy-matches a parsed name (with optional club hint) against
// a caller-supplied candidate pool.
type PlayerResolver interface {
	Resolve(ctx context.Context, name, clubHint string, candidates []PlayerCandidate) ([]PlayerMatch, error)
}

// MatchStats is the parsed output from a single match's stats page.
type MatchStats struct {
	HomeClubName    string
	AwayClubName    string
	HomeTeamGoals   int
	HomeTeamBehinds int
	AwayTeamGoals   int
	AwayTeamBehinds int
	Players         []PlayerStats
}

// PlayerStats is one player's stats from a match page.
type PlayerStats struct {
	Name          string // display name as shown on the source page
	CanonicalName string // slug-derived name (e.g. "Jason Horne Francis" from URL); preferred for matching when non-empty
	ClubName      string
	Kicks         int
	Handballs     int
	Marks         int
	Hitouts       int
	Tackles       int
	Goals         int
	Behinds       int
}

// StatsParser parses player stats for a single match from an external source.
// mid is the source-specific match identifier (e.g. FootyWire's mid= query param).
type StatsParser interface {
	ParseMatch(ctx context.Context, mid string) (MatchStats, error)
}

// FixtureDiscovery finds the external source ID for a match when not already cached.
type FixtureDiscovery interface {
	FindMatchMid(ctx context.Context, roundName, homeClub, awayClub string) (string, error)
}

// DataopsMatchSourceRepository persists the mapping of afl.match_id → external source ID
// per the ACL pattern (ADR-016). Source identifies the integration (e.g. "footywire").
type DataopsMatchSourceRepository interface {
	FindByMatchID(ctx context.Context, source string, matchID int) (externalID string, found bool, err error)
	Store(ctx context.Context, source, externalID string, matchID int) error
}
