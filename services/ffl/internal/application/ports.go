package application

import "context"

// PlayerCandidate is a known player that can be matched against a parsed name.
type PlayerCandidate struct {
	PlayerID    int
	AFLPlayerID int
	Name        string
	Club        string // AFL club name from afl.club
}

// PlayerMatch is the result of resolving a parsed name against a candidate pool.
type PlayerMatch struct {
	Candidate  PlayerCandidate
	Confidence float64 // 0.0–1.0
}

// PlayerLookup fetches player names from the AFL service by AFL player ID.
// Used to build candidate pools without reading ffl.player.drv_name.
type PlayerLookup interface {
	LookupPlayers(ctx context.Context, aflPlayerIDs []int) ([]PlayerCandidate, error)
}

// PlayerResolver fuzzy-matches a parsed name (with optional club hint) against
// a caller-supplied candidate pool. Decoupled from the record type being matched.
type PlayerResolver interface {
	Resolve(ctx context.Context, name, clubHint string, candidates []PlayerCandidate) ([]PlayerMatch, error)
}

// TeamParser parses a raw forum post into structured player rows.
// The caller supplies the FFL team name and round number; the parser
// identifies player names, positions, and optional scores.
type TeamParser interface {
	Parse(ctx context.Context, teamName, post string) ([]ParsedPlayerRow, error)
}

// ParsedPlayerRow is one player line extracted from a forum post.
type ParsedPlayerRow struct {
	Name                string
	ClubHint            string // AFL club code as written in the post (e.g. "Geel", "WB")
	Position            string // primary position ("goals", "kicks", …, "bench")
	BackupPositions     string // comma-separated, bench players only
	InterchangePosition string // bench players with interchange designation
	Score               *int   // nil if not present in the post
	Notes               string
}
